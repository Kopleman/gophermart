package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/Kopleman/gophermart/internal/accrual/mocks"
	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAccrual_sendRequestToAccrual(t *testing.T) {
	orderNumber := "123"
	expectedURL := "/" + orderNumber

	t.Run("successful request", func(t *testing.T) {
		client := new(mocks.HTTPClient)
		expectedResponse := dto.AccrualResponseDTO{
			Order:   orderNumber,
			Status:  dto.AccrualStatusTypePROCESSED,
			Accrual: new(float64),
		}
		*expectedResponse.Accrual = 100.5
		responseData, _ := json.Marshal(expectedResponse)

		client.On("Get", expectedURL, "application/json").Return(responseData, nil, nil)

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		response, err := a.sendRequestToAccrual(orderNumber)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Order, response.Order)
		client.AssertExpectations(t)
	})

	t.Run("http error", func(t *testing.T) {
		client := new(mocks.HTTPClient)
		expectedErr := errors.New("connection error")
		client.On("Get", expectedURL, "application/json").Return([]byte{}, nil, expectedErr)

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		_, err := a.sendRequestToAccrual(orderNumber)

		assert.ErrorContains(t, err, "GET request")
		client.AssertExpectations(t)
	})

	t.Run("invalid json", func(t *testing.T) {
		client := new(mocks.HTTPClient)
		client.On("Get", expectedURL, "application/json").Return([]byte("{invalid}"), nil, nil)

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		_, err := a.sendRequestToAccrual(orderNumber)

		assert.ErrorContains(t, err, "unmarshal response")
		client.AssertExpectations(t)
	})

	t.Run("429 response", func(t *testing.T) {
		client := new(mocks.HTTPClient)
		expectedResponse := dto.AccrualResponseDTO{
			Order:   orderNumber,
			Status:  dto.AccrualStatusTypePROCESSED,
			Accrual: new(float64),
		}
		*expectedResponse.Accrual = 100.5
		responseBytes, _ := json.Marshal(expectedResponse)
		firstResponse := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     make(http.Header),
		}
		firstResponse.Header.Add("Retry-After", "1")
		firstResponse.StatusCode = http.StatusTooManyRequests

		client.On("Get", expectedURL, "application/json").Return(nil, firstResponse, nil).Once()
		client.On("Get", expectedURL, "application/json").Return(responseBytes, nil, nil).Once()

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		response, err := a.sendRequestToAccrual(orderNumber)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Order, response.Order)
		client.AssertExpectations(t)
	})
}

func TestAccrual_registerOrder(t *testing.T) {
	ctx := context.Background()
	order := &pgxstore.OrdersToProcess{OrderNumber: "123"}

	t.Run("successful registration", func(t *testing.T) {
		expectedResponse := dto.AccrualResponseDTO{
			Order:   order.OrderNumber,
			Status:  dto.AccrualStatusTypePROCESSING,
			Accrual: new(float64),
		}
		responseData, _ := json.Marshal(expectedResponse)
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil, nil)
		repo.On("RegisterOrderProcessing", ctx, order.OrderNumber).Return(nil)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		err := a.registerOrder(ctx, order)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		client.AssertExpectations(t)
	})

	t.Run("registration error - accrual error", func(t *testing.T) {
		expectedResponse := struct{}{}
		responseData, _ := json.Marshal(expectedResponse)
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)
		expectedErr := errors.New("accrual error")
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil, expectedErr).Once()

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		err := a.registerOrder(ctx, order)

		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})

	t.Run("registration error - db error", func(t *testing.T) {
		expectedResponse := dto.AccrualResponseDTO{
			Order:   order.OrderNumber,
			Status:  dto.AccrualStatusTypePROCESSING,
			Accrual: new(float64),
		}
		responseData, _ := json.Marshal(expectedResponse)
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)
		expectedErr := errors.New("database error")
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil, nil).Once()
		repo.On("RegisterOrderProcessing", ctx, order.OrderNumber).Return(expectedErr)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		err := a.registerOrder(ctx, order)

		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})
}

func TestAccrual_processOrder(t *testing.T) {
	ctx := context.Background()
	order := &pgxstore.OrdersToProcess{OrderNumber: "123"}

	t.Run("successful processing", func(t *testing.T) {
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)

		expectedProcessingResponse := dto.AccrualResponseDTO{
			Order:   order.OrderNumber,
			Status:  dto.AccrualStatusTypePROCESSING,
			Accrual: new(float64),
		}
		responseData, _ := json.Marshal(expectedProcessingResponse)
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil, nil).Once()

		expectedProcessedResponse := dto.AccrualResponseDTO{
			Order:   order.OrderNumber,
			Status:  dto.AccrualStatusTypePROCESSED,
			Accrual: new(float64),
		}
		*expectedProcessedResponse.Accrual = 100.5
		responseData, _ = json.Marshal(expectedProcessedResponse)
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil, nil).Once()

		repo.On("StoreAccrualCalculation", ctx, mock.Anything).Return(nil)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		err := a.processOrder(ctx, order)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		client.AssertExpectations(t)
	})

	t.Run("invalid status handling", func(t *testing.T) {
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)
		expectedResponse := dto.AccrualResponseDTO{
			Order:  order.OrderNumber,
			Status: dto.AccrualStatusTypeINVALID,
		}
		responseData, _ := json.Marshal(expectedResponse)

		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil, nil).Once()
		repo.On("StoreAccrualCalculation", ctx, mock.Anything).Return(nil)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		err := a.processOrder(ctx, order)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestAccrual_workerFunctions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	order := &pgxstore.OrdersToProcess{OrderNumber: "123"}
	ordersChan := make(chan *pgxstore.OrdersToProcess, 1)
	ordersChan <- order
	close(ordersChan)

	t.Run("register worker", func(t *testing.T) {
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)
		client.On("Get", mock.Anything, mock.Anything).Return([]byte{}, nil)
		repo.On("RegisterOrderProcessing", ctx, order.OrderNumber).Return(nil)

		wg := &sync.WaitGroup{}
		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		wg.Add(1)
		go a.startRegisterOrdersToProcessWorker(ctx, wg, ordersChan, make(chan error), 1)
	})

	t.Run("processing worker", func(t *testing.T) {
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mocks.HTTPClient)
		client.On("Get", mock.Anything, mock.Anything).Return([]byte{}, nil, nil)
		repo.On("StoreAccrualCalculation", ctx, mock.Anything).Return(nil)

		wg := &sync.WaitGroup{}
		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		wg.Add(1)
		go a.startProcessingOrdersToProcessWorker(ctx, wg, ordersChan, make(chan error), 1)
	})
}
