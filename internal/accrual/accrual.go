package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/shopspring/decimal"
)

const baseBackoffMaxWait = 60

type OrderRepoForAccrual interface {
	PickOrdersToProcess(ctx context.Context, limit int32) ([]*pgxstore.OrdersToProcess, error)
	GetRegisteredProcessingOrders(ctx context.Context, limit int32) ([]*pgxstore.OrdersToProcess, error)
	GetStartProcessingOrders(ctx context.Context) ([]*pgxstore.OrdersToProcess, error)
	RegisterOrderProcessing(ctx context.Context, orderNumber string) error
	StoreAccrualCalculation(
		ctx context.Context,
		params pgxstore.AccrualCalculationParams,
	) error
}

type HTTPClient interface {
	Get(url, contentType string) ([]byte, *http.Response, error)
}

type Accrual struct {
	logger        log.Logger
	cfg           *config.Config
	repo          OrderRepoForAccrual
	httpClient    HTTPClient
	mu            *sync.Mutex
	nextRetryTime time.Time
}

func New(logger log.Logger, cfg *config.Config, repo OrderRepoForAccrual, client HTTPClient) *Accrual {
	return &Accrual{
		logger:        logger,
		cfg:           cfg,
		repo:          repo,
		httpClient:    client,
		mu:            &sync.Mutex{},
		nextRetryTime: time.Now(),
	}
}

type processJobParams struct {
	nextJobTime              time.Time
	tickerChan               <-chan time.Time
	errorChan                chan<- error
	ordersToProcessFetchFunc func(ctx context.Context, limit int32) ([]*pgxstore.OrdersToProcess, error)
	backUpFunc               func(ctx context.Context, ordersChan chan *pgxstore.OrdersToProcess) error
	name                     string
	interval                 time.Duration
}

func noopBackupFunc(ctx context.Context, ordersChan chan *pgxstore.OrdersToProcess) error {
	return nil
}

func (a *Accrual) backUpProcessingJob(ctx context.Context, ordersChan chan *pgxstore.OrdersToProcess) error {
	processedOrders, err := a.repo.GetStartProcessingOrders(ctx)
	if err != nil {
		return fmt.Errorf("backUpProcessingJob fetch: %w", err)
	}
	a.logger.Infof("Amount of orders to process from last restart: %v", len(processedOrders))
	for {
		skippedOrders := make([]*pgxstore.OrdersToProcess, 0, len(processedOrders))
		for _, order := range processedOrders {
			if len(ordersChan) == cap(ordersChan) {
				skippedOrders = append(skippedOrders, order)
				continue
			}
			ordersChan <- order
		}
		if len(skippedOrders) == 0 {
			return nil
		}
		processedOrders = skippedOrders
	}
}

func (a *Accrual) genOrdersToProcessChan(ctx context.Context, params *processJobParams) chan *pgxstore.OrdersToProcess {
	ordersChan := make(chan *pgxstore.OrdersToProcess, a.cfg.MaxOrdersInWork)

	go func() {
		defer close(ordersChan)
		// first we push to work orders that have been in work on previous run
		if err := params.backUpFunc(ctx, ordersChan); err != nil {
			params.errorChan <- fmt.Errorf("failed to perform backup processing: %w", err)
			return
		}
		for {
			select {
			case currentTickerTime := <-params.tickerChan:
				if currentTickerTime.After(params.nextJobTime) || currentTickerTime.Equal(params.nextJobTime) {
					params.nextJobTime = currentTickerTime.Add(params.interval)
					currentAmountInWork := len(ordersChan)
					if len(ordersChan) == cap(ordersChan) {
						break
					}
					newButchSize := int(a.cfg.MaxOrdersInWork) - currentAmountInWork
					orders, err := params.ordersToProcessFetchFunc(ctx, int32(newButchSize))
					if err != nil {
						params.errorChan <- fmt.Errorf("failed to pick orders: %w", err)
					}
					a.logger.Infof("fetched orders to process for %s: %d", params.name, len(orders))
					for _, order := range orders {
						ordersChan <- order
					}
				}
			case <-ctx.Done():
				a.logger.Infof("stopping pushing orders to channel for %s", params.name)
				return
			}
		}
	}()

	return ordersChan
}

func (a *Accrual) getReqBackoff(resp *http.Response) time.Duration {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			parsedHeaderValue, err := strconv.Atoi(retryAfter)
			a.mu.Lock()
			a.nextRetryTime = time.Now().Add(baseBackoffMaxWait * time.Second)
			if err == nil {
				a.nextRetryTime = time.Now().Add(time.Duration(parsedHeaderValue) * time.Second)
			}
			a.mu.Unlock()
		}
	}
	retryIn := time.Until(a.nextRetryTime)
	if retryIn <= 0 {
		retryIn = 0
	}
	return retryIn
}

func (a *Accrual) sendRequestToAccrual(orderNumber string) (*dto.AccrualResponseDTO, error) {
	url := "/" + orderNumber
	time.Sleep(a.getReqBackoff(nil))
	bodyBytes, resp, err := a.httpClient.Get(url, "application/json") //nolint:bodyclose // its closed in client
	for resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if bodyParseErr := resp.Body.Close(); bodyParseErr != nil {
			a.logger.Error(bodyParseErr)
		}
		time.Sleep(a.getReqBackoff(resp))
		retriedBodyBytes, retriedResp, retryErr := a.httpClient.Get( //nolint:bodyclose // its closed in client
			url,
			"application/json",
		)
		resp = retriedResp
		err = retryErr
		bodyBytes = retriedBodyBytes
	}

	if err != nil {
		return nil, fmt.Errorf("sendRequestToAccrual GET request: %w", err)
	}
	responseDTO := new(dto.AccrualResponseDTO)
	if unmarshalErr := json.Unmarshal(bodyBytes, responseDTO); unmarshalErr != nil {
		return nil, fmt.Errorf("sendRequestToAccrual unmarshal response: %w", unmarshalErr)
	}

	return responseDTO, nil
}

func (a *Accrual) registerOrder(ctx context.Context, order *pgxstore.OrdersToProcess) error {
	a.logger.Infof("Registering order: %s", order.OrderNumber)
	_, err := a.sendRequestToAccrual(order.OrderNumber)
	if err != nil {
		return fmt.Errorf("registerOrder: %w", err)
	}
	if err = a.repo.RegisterOrderProcessing(ctx, order.OrderNumber); err != nil {
		return fmt.Errorf("registerOrder: %w", err)
	}
	return nil
}

func (a *Accrual) processOrder(ctx context.Context, order *pgxstore.OrdersToProcess) error {
	a.logger.Infof("Processing order: %s", order.OrderNumber)
	accrualDto := new(dto.AccrualResponseDTO)
	orderProcessed := false
	for !orderProcessed {
		responseDto, err := a.sendRequestToAccrual(order.OrderNumber)
		if err != nil {
			return fmt.Errorf("processOrder: %w", err)
		}
		if responseDto.Status == dto.AccrualStatusTypePROCESSING ||
			responseDto.Status == dto.AccrualStatusTypeREGISTERED {
			continue
		}
		accrualDto = responseDto
		orderProcessed = true
	}

	orderStatus := pgxstore.OrderStatusTypePROCESSED
	if accrualDto.Status == dto.AccrualStatusTypeINVALID {
		orderStatus = pgxstore.OrderStatusTypeINVALID
	}
	amount := decimal.Zero
	if accrualDto.Accrual != nil {
		amount = decimal.NewFromFloat(*accrualDto.Accrual)
	}

	storeParams := pgxstore.AccrualCalculationParams{
		OrderNumber: order.OrderNumber,
		Amount:      amount,
		Status:      orderStatus,
	}

	if err := a.repo.StoreAccrualCalculation(ctx, storeParams); err != nil {
		return fmt.Errorf("processOrder: %w", err)
	}

	return nil
}

// This worker register new orders for calculation to Accrual.
// It not handle calculation itself.
func (a *Accrual) startRegisterOrdersToProcessWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	ordersChan <-chan *pgxstore.OrdersToProcess,
	errorChan chan<- error,
	workerID int,
) {
	defer wg.Done()
	for {
		select {
		case order := <-ordersChan:
			if err := a.registerOrder(ctx, order); err != nil {
				errorChan <- fmt.Errorf("failed to register order: %w", err)
			}
		case <-ctx.Done():
			a.logger.Infof("stopping registering orders for %v", workerID)
			return
		}
	}
}

// This worker fetch data from accrual to store deposit.
func (a *Accrual) startProcessingOrdersToProcessWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	ordersChan <-chan *pgxstore.OrdersToProcess,
	errorChan chan<- error,
	workerID int,
) {
	defer wg.Done()
	for {
		select {
		case order := <-ordersChan:
			if err := a.processOrder(ctx, order); err != nil {
				errorChan <- fmt.Errorf("failed to register order: %w", err)
			}
		case <-ctx.Done():
			a.logger.Infof("stopping registering orders to process worker %v", workerID)
			return
		}
	}
}

func (a *Accrual) Run(ctx context.Context) error {
	innerCtx, cancelFunc := context.WithCancel(ctx)
	a.logger.Info("Starting orders processing")
	pollTicker := time.NewTicker(1 * time.Second)
	defer pollTicker.Stop()

	pollDuration := time.Duration(a.cfg.PollInterval) * time.Second

	errChan := make(chan error, a.cfg.WorkerLimit*2) //nolint:all // we just double chan buffer that every worker has chance to push error
	defer close(errChan)

	ordersToRegisterParams := processJobParams{
		name:                     "order register",
		nextJobTime:              time.Now().Add(pollDuration),
		interval:                 pollDuration,
		errorChan:                errChan,
		tickerChan:               pollTicker.C,
		ordersToProcessFetchFunc: a.repo.PickOrdersToProcess,
		backUpFunc:               a.backUpProcessingJob,
	}
	ordersToRegisterChan := a.genOrdersToProcessChan(innerCtx, &ordersToRegisterParams)

	maxWorkerCount := int(a.cfg.WorkerLimit)

	wg := &sync.WaitGroup{}

	for w := 1; w <= maxWorkerCount; w++ {
		wg.Add(1)
		go a.startRegisterOrdersToProcessWorker(innerCtx, wg, ordersToRegisterChan, errChan, w)
	}

	ordersToProcessParams := processJobParams{
		name:                     "order processing",
		nextJobTime:              time.Now().Add(pollDuration),
		interval:                 pollDuration,
		errorChan:                errChan,
		tickerChan:               pollTicker.C,
		ordersToProcessFetchFunc: a.repo.GetRegisteredProcessingOrders,
		backUpFunc:               noopBackupFunc,
	}
	ordersToProcessChan := a.genOrdersToProcessChan(innerCtx, &ordersToProcessParams)

	for w := 1; w <= maxWorkerCount; w++ {
		wg.Add(1)
		go a.startProcessingOrdersToProcessWorker(innerCtx, wg, ordersToProcessChan, errChan, w)
	}

	for {
		select {
		case err := <-errChan:
			if err != nil {
				cancelFunc()
				wg.Wait()
				return fmt.Errorf("accrual order processing failed: %w", err)
			}
		case <-ctx.Done():
			a.logger.Infof("gracefully shutting down accrual service")
			cancelFunc()
			wg.Wait()
			return nil
		}
	}
}
