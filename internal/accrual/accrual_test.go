package accrual

import (
	"context"
	"encoding/json"
	"errors"
	//"fmt"
	"testing"
	//"time"

	"github.com/Kopleman/gophermart/internal/accrual/mocks"
	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	//"github.com/Kopleman/gophermart/internal/pgxstore"
	//"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Get(url, contentType string) ([]byte, error) {
	args := m.Called(url, contentType)
	return args.Get(0).([]byte), args.Error(1)
}

func TestAccrual_sendRequestToAccrual(t *testing.T) {
	orderNumber := "123"
	expectedURL := "/" + orderNumber

	t.Run("successful request", func(t *testing.T) {
		client := new(mockHTTPClient)
		expectedResponse := dto.AccrualResponseDTO{
			Order:   orderNumber,
			Status:  dto.AccrualStatusTypePROCESSED,
			Accrual: new(float64),
		}
		*expectedResponse.Accrual = 100.5
		responseData, _ := json.Marshal(expectedResponse)

		client.On("Get", expectedURL, "application/json").Return(responseData, nil)

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		response, err := a.sendRequestToAccrual(orderNumber)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Order, response.Order)
		client.AssertExpectations(t)
	})

	t.Run("http error", func(t *testing.T) {
		client := new(mockHTTPClient)
		expectedErr := errors.New("connection error")
		client.On("Get", expectedURL, "application/json").Return([]byte{}, expectedErr)

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		_, err := a.sendRequestToAccrual(orderNumber)

		assert.ErrorContains(t, err, "GET request")
		client.AssertExpectations(t)
	})

	t.Run("invalid json", func(t *testing.T) {
		client := new(mockHTTPClient)
		client.On("Get", expectedURL, "application/json").Return([]byte("{invalid}"), nil)

		a := New(log.MockLogger{}, &config.Config{}, new(mocks.OrderRepoForAccrual), client)
		_, err := a.sendRequestToAccrual(orderNumber)

		assert.ErrorContains(t, err, "unmarshal response")
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
		client := new(mockHTTPClient)
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil)
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
		client := new(mockHTTPClient)
		expectedErr := errors.New("accrual error")
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, expectedErr)

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
		client := new(mockHTTPClient)
		expectedErr := errors.New("database error")
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil)
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
		client := new(mockHTTPClient)

		expectedProcessingResponse := dto.AccrualResponseDTO{
			Order:   order.OrderNumber,
			Status:  dto.AccrualStatusTypePROCESSING,
			Accrual: new(float64),
		}
		responseData, _ := json.Marshal(expectedProcessingResponse)
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil).Once()

		expectedProcessedResponse := dto.AccrualResponseDTO{
			Order:   order.OrderNumber,
			Status:  dto.AccrualStatusTypePROCESSED,
			Accrual: new(float64),
		}
		*expectedProcessedResponse.Accrual = 100.5
		responseData, _ = json.Marshal(expectedProcessedResponse)
		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil).Once()

		repo.On("StoreAccrualCalculation", ctx, mock.Anything).Return(nil)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		err := a.processOrder(ctx, order)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		client.AssertExpectations(t)
	})

	t.Run("invalid status handling", func(t *testing.T) {
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mockHTTPClient)
		expectedResponse := dto.AccrualResponseDTO{
			Order:  order.OrderNumber,
			Status: dto.AccrualStatusTypeINVALID,
		}
		responseData, _ := json.Marshal(expectedResponse)

		client.On("Get", mock.Anything, mock.Anything).Return(responseData, nil)
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
		client := new(mockHTTPClient)
		client.On("Get", mock.Anything, mock.Anything).Return([]byte{}, nil)
		repo.On("RegisterOrderProcessing", ctx, order.OrderNumber).Return(nil)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		go a.startRegisterOrdersToProcessWorker(ctx, ordersChan, make(chan error), 1)
	})

	t.Run("processing worker", func(t *testing.T) {
		repo := new(mocks.OrderRepoForAccrual)
		client := new(mockHTTPClient)
		client.On("Get", mock.Anything, mock.Anything).Return([]byte{}, nil)
		repo.On("StoreAccrualCalculation", ctx, mock.Anything).Return(nil)

		a := New(log.MockLogger{}, &config.Config{}, repo, client)
		go a.startProcessingOrdersToProcessWorker(ctx, ordersChan, make(chan error), 1)
	})
}
