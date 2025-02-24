//nolint:all // test-cases
package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Kopleman/gophermart/internal/common"
	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/controller"
	"github.com/Kopleman/gophermart/internal/controller/mocks"
	"github.com/Kopleman/gophermart/internal/middlerware"
	"github.com/Kopleman/gophermart/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var nullIdToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAifQ.PwlKpYiTzOArQcGsAKQHO13kyVxZNi3_0G81nXNZ4p0"

func setupServer() (*Server, *mocks.UserService, *mocks.OrderService, *mocks.BalanceService) {
	cfg := &config.Config{JWTSecret: "secret_key"}
	s := NewServer(log.MockLogger{}, cfg)
	s.app = fiber.New()
	validatorInstance := validator.New()
	userService := new(mocks.UserService)
	orderService := new(mocks.OrderService)
	balanceService := new(mocks.BalanceService)
	userController := controller.NewUserController(s.logger, validatorInstance, s.config, userService)
	orderController := controller.NewOrderController(s.logger, validatorInstance, s.config, orderService)
	balanceController := controller.NewBalanceController(s.logger, validatorInstance, s.config, balanceService)
	s.applyRoutes(middlerware.NewAuthMiddleWare(s.config), userController, orderController, balanceController)

	return s, userService, orderService, balanceService
}

func testRequest(t *testing.T, app *fiber.App, method, path string, body io.Reader, authToken *string) (int, string) {
	t.Helper()

	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)
	if authToken != nil {
		req.Header.Set("Authorization", "Bearer "+*authToken)
	}
	require.NoError(t, err)
	if body != http.NoBody {
		req.Header.Set(common.ContentType, "application/json")
	}

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer func() {
		err = errors.Join(err, resp.Body.Close())
		require.NoError(t, err)
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(respBody)
}

func TestBaseRouterPath_Server(t *testing.T) {
	testServer, _, _, _ := setupServer()
	app := testServer.app

	var testTable = []struct {
		method  string
		url     string
		body    io.Reader
		want    string
		status  int
		hasJSON bool
	}{
		{
			"GET",
			"/some-url",
			http.NoBody,
			`Not Found`,
			http.StatusNotFound,
			false,
		},
		{
			"GET",
			"/api",
			http.NoBody,
			`Not Found`,
			http.StatusNotFound,
			false,
		},
		{
			"GET",
			"/api/api-docs",
			http.NoBody,
			``,
			http.StatusMovedPermanently,
			false,
		},
	}
	for _, v := range testTable {
		gotStatusCode, gotResponse := testRequest(t, app, v.method, v.url, v.body, nil)
		assert.Equal(t, v.status, gotStatusCode)

		if !v.hasJSON {
			assert.Equal(t, v.want, gotResponse)
			continue
		}

		assert.JSONEq(t, v.want, gotResponse)
	}
}

func TestUserRegisterRouterPath_Server(t *testing.T) {
	testServer, userService, _, _ := setupServer()
	app := testServer.app

	t.Run("successful user register", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}
		expectedToken := "token-string"

		userService.On("CreateUser", mock.Anything, expectedDto).Return(nil).Once()
		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("token-string", nil).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/register", body, nil)
		assert.Equal(t, http.StatusOK, gotStatusCode)
		assert.JSONEq(t, fmt.Sprintf(`{"token": "%s"}`, expectedToken), gotResponse)
	})

	t.Run("user already exists", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("CreateUser", mock.Anything, expectedDto).Return(service.ErrAlreadyExists).Once()
		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/register", body, nil)
		assert.Equal(t, http.StatusConflict, gotStatusCode)
		assert.Equal(t, "Conflict", gotResponse)
	})

	t.Run("post auth error 1", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("CreateUser", mock.Anything, expectedDto).Return(nil).Once()
		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("", service.ErrNotFound).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/register", body, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("post auth error 2", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("CreateUser", mock.Anything, expectedDto).Return(nil).Once()
		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("", service.ErrInvalidArguments).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/register", body, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("CreateUser", mock.Anything, expectedDto).Return(errors.New("something bad happened")).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/register", body, nil)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})

	var testTable = []struct {
		body string
	}{
		{
			`{"login": "foo"}`,
		},
		{
			`{"password": "foo"}`,
		},
		{
			`{}`,
		},
		{
			``,
		},
		{
			`some plain text`,
		},
	}
	for _, v := range testTable {
		t.Run(fmt.Sprintf("Request validation for - %s", fmt.Sprint(v.body)), func(t *testing.T) {
			gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/register", strings.NewReader(v.body), nil)
			assert.Equal(t, http.StatusBadRequest, gotStatusCode)
			assert.Equal(t, "Bad Request", gotResponse)
		})
	}
}

func TestLoginUserRouterPath_Server(t *testing.T) {
	testServer, userService, _, _ := setupServer()
	app := testServer.app

	t.Run("successful user login", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}
		expectedToken := "token-string"

		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("token-string", nil).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/login", body, nil)
		assert.Equal(t, http.StatusOK, gotStatusCode)
		assert.JSONEq(t, fmt.Sprintf(`{"token": "%s"}`, expectedToken), gotResponse)
	})

	t.Run("user not found", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("CreateUser", mock.Anything, expectedDto).Return(nil).Once()
		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("", service.ErrNotFound).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/login", body, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("bad password", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("CreateUser", mock.Anything, expectedDto).Return(nil).Once()
		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("", service.ErrInvalidArguments).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/login", body, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		expectedDto := &dto.UserCredentialsDTO{
			Login:    "foo",
			Password: "bar",
		}

		userService.On("AuthorizeUser", mock.Anything, expectedDto).Return("", errors.New("something bad happened")).Once()

		body := strings.NewReader(fmt.Sprintf(`{"login": "%s", "password": "%s"}`, expectedDto.Login, expectedDto.Password))

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/login", body, nil)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})

	var testTable = []struct {
		body string
	}{
		{
			`{"login": "foo"}`,
		},
		{
			`{"password": "foo"}`,
		},
		{
			`{}`,
		},
		{
			``,
		},
		{
			`some plain text`,
		},
	}
	for _, v := range testTable {
		t.Run(fmt.Sprintf("Request validation for - %s", fmt.Sprint(v.body)), func(t *testing.T) {
			gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/login", strings.NewReader(v.body), nil)
			assert.Equal(t, http.StatusBadRequest, gotStatusCode)
			assert.Equal(t, "Bad Request", gotResponse)
		})
	}
}

func TestGetWithdrawalsRouterPath_Server(t *testing.T) {
	testServer, userService, _, _ := setupServer()
	app := testServer.app

	t.Run("successful fetch with withdrawals list", func(t *testing.T) {
		expectedDto := &dto.WithdrawalItemDTO{
			Order:       "1",
			ProcessedAt: "2025-02-21T15:02:07.614Z",
			Sum:         100,
		}
		expectedDtos := []*dto.WithdrawalItemDTO{expectedDto}

		userService.On("GetWithdrawals", mock.Anything, uuid.Nil).Return(expectedDtos, nil).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/withdrawals", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusOK, gotStatusCode)
		assert.JSONEq(
			t,
			fmt.Sprintf(
				`[{"order": "%s", "processed_at": "%s", "sum": %v}]`,
				expectedDto.Order,
				expectedDto.ProcessedAt,
				expectedDto.Sum), gotResponse)
	})

	t.Run("successful fetch without items", func(t *testing.T) {

		var expectedDtos []*dto.WithdrawalItemDTO

		userService.On("GetWithdrawals", mock.Anything, uuid.Nil).Return(expectedDtos, nil).Once()

		gotStatusCode, _ := testRequest(t, app, "GET", "/api/user/withdrawals", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusNoContent, gotStatusCode)
	})

	t.Run("auth check", func(t *testing.T) {
		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/withdrawals", http.NoBody, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		userService.On("GetWithdrawals", mock.Anything, uuid.Nil).Return(nil, errors.New("something bad happened")).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/withdrawals", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})
}

func TestGetUserBalanceRouterPath_Server(t *testing.T) {
	testServer, _, _, balanceService := setupServer()
	app := testServer.app

	t.Run("successful balance fetch", func(t *testing.T) {
		expectedDto := &dto.BalanceDTO{
			Current:   1000,
			Withdrawn: 1000,
		}

		balanceService.On("GetUserBalanceDTO", mock.Anything, uuid.Nil).Return(expectedDto, nil).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/balance", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusOK, gotStatusCode)
		assert.JSONEq(
			t,
			fmt.Sprintf(
				`{"current": %v, "withdrawn": %v}`,
				expectedDto.Current,
				expectedDto.Withdrawn), gotResponse)
	})

	t.Run("auth check", func(t *testing.T) {
		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/balance", http.NoBody, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		balanceService.On("GetUserBalanceDTO", mock.Anything, uuid.Nil).Return(nil, errors.New("something bad happened")).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/balance", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})
}

func TestMakeWithdrawRouterPath_Server(t *testing.T) {
	testServer, _, _, balanceService := setupServer()
	app := testServer.app

	t.Run("successful withdraw place", func(t *testing.T) {
		requestDto := &dto.WithdrawRequestDTO{
			Order: "49927398716",
			Sum:   100,
		}
		body := strings.NewReader(fmt.Sprintf(`{"order": "%s", "sum": %v}`, requestDto.Order, requestDto.Sum))
		expectedDto := &dto.WithdrawDTO{
			Order:  requestDto.Order,
			Amount: decimal.NewFromFloat(requestDto.Sum),
			UserID: uuid.Nil,
		}

		balanceService.On("MakeWithdraw", mock.Anything, expectedDto).Return(nil).Once()

		gotStatusCode, _ := testRequest(t, app, "POST", "/api/user/balance/withdraw", body, &nullIdToken)
		assert.Equal(t, http.StatusOK, gotStatusCode)
	})

	t.Run("auth check", func(t *testing.T) {
		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/balance/withdraw", http.NoBody, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		requestDto := &dto.WithdrawRequestDTO{
			Order: "49927398716",
			Sum:   100,
		}
		body := strings.NewReader(fmt.Sprintf(`{"order": "%s", "sum": %v}`, requestDto.Order, requestDto.Sum))
		expectedDto := &dto.WithdrawDTO{
			Order:  requestDto.Order,
			Amount: decimal.NewFromFloat(requestDto.Sum),
			UserID: uuid.Nil,
		}
		balanceService.On("MakeWithdraw", mock.Anything, expectedDto).Return(errors.New("something bad happened")).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/balance/withdraw", body, &nullIdToken)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})

	var testTable = []struct {
		body string
	}{
		{
			`{"order": "foo"}`,
		},
		{
			`{"order": "49927398716", "sum"": "bar"}`,
		},
		{
			`{"order": "49927398716", "sum": "100"}`,
		},
		{
			`{ "sum": 100 }`,
		},
		{
			`{"order": "4992739871", "sum": 100}`,
		},
		{
			`{}`,
		},
		{
			``,
		},
		{
			`some plain text`,
		},
	}
	for _, v := range testTable {
		t.Run(fmt.Sprintf("Request validation for - %s", fmt.Sprint(v.body)), func(t *testing.T) {
			gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/balance/withdraw", strings.NewReader(v.body), &nullIdToken)
			assert.Equal(t, http.StatusUnprocessableEntity, gotStatusCode)
			assert.Equal(t, "Unprocessable Entity", gotResponse)
		})
	}
}

func TestAddOrderRouterPath_Server(t *testing.T) {
	testServer, _, orderService, _ := setupServer()
	app := testServer.app

	t.Run("successful order place", func(t *testing.T) {
		orderNumber := "49927398716"
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		expectedDto := &dto.CreateOrderDTO{
			OrderNumber: orderNumber,
			UserID:      uuid.Nil,
		}

		orderService.On("GetOrderByNumber", mock.Anything, orderNumber).Return(nil, service.ErrNotFound).Once()
		orderService.On("CreateOrder", mock.Anything, expectedDto).Return(nil).Once()

		gotStatusCode, _ := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusAccepted, gotStatusCode)
	})

	t.Run("place same order", func(t *testing.T) {
		orderNumber := "49927398716"
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		expectedDto := &dto.OrderDTO{
			OrderNumber: orderNumber,
			UserID:      uuid.Nil,
		}

		orderService.On("GetOrderByNumber", mock.Anything, orderNumber).Return(expectedDto, nil).Once()

		gotStatusCode, _ := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusOK, gotStatusCode)
	})

	t.Run("auth check", func(t *testing.T) {
		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/orders", http.NoBody, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		orderNumber := "49927398716"
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		orderService.On("GetOrderByNumber", mock.Anything, orderNumber).Return(nil, errors.New("something bad happened")).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})

	t.Run("500 error - 2", func(t *testing.T) {
		orderNumber := "49927398716"
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		expectedDto := &dto.CreateOrderDTO{
			OrderNumber: orderNumber,
			UserID:      uuid.Nil,
		}

		orderService.On("GetOrderByNumber", mock.Anything, orderNumber).Return(nil, service.ErrNotFound).Once()
		orderService.On("CreateOrder", mock.Anything, expectedDto).Return(errors.New("something bad happened")).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})

	t.Run("bad order num", func(t *testing.T) {
		orderNumber := "4992739871"
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusUnprocessableEntity, gotStatusCode)
		assert.Equal(t, "Unprocessable Entity", gotResponse)
	})

	t.Run("bad order num", func(t *testing.T) {
		orderNumber := ""
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusBadRequest, gotStatusCode)
		assert.Equal(t, "Bad Request", gotResponse)
	})

	t.Run("order of another user", func(t *testing.T) {
		orderNumber := "49927398716"
		body := strings.NewReader(fmt.Sprintf(`%s`, orderNumber))
		userUuid := uuid.New()
		expectedDto := &dto.OrderDTO{
			OrderNumber: orderNumber,
			UserID:      userUuid,
		}

		orderService.On("GetOrderByNumber", mock.Anything, orderNumber).Return(expectedDto, nil).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "POST", "/api/user/orders", body, &nullIdToken)
		assert.Equal(t, http.StatusConflict, gotStatusCode)
		assert.Equal(t, "Conflict", gotResponse)
	})
}

func TestGetOrdersRouterPath_Server(t *testing.T) {
	testServer, _, orderService, _ := setupServer()
	app := testServer.app

	t.Run("successful fetch with order list", func(t *testing.T) {
		orderDto := &dto.OrderDTO{
			OrderNumber: "49927398716",
			Status:      "NEW",
			CreatedAt:   "2022-01-01T00:00:00+00:00",
			Accrual:     decimal.NewFromFloat(100),
			ID:          uuid.New(),
			UserID:      uuid.Nil,
		}
		orderDtos := []*dto.OrderDTO{orderDto}
		accrualValue, _ := orderDto.Accrual.Float64()
		expectedDto := &dto.OrderInfoDTO{
			Number:     orderDto.OrderNumber,
			Status:     orderDto.Status,
			Accrual:    &accrualValue,
			UploadedAt: orderDto.CreatedAt,
		}

		orderService.On("GetUserOrders", mock.Anything, uuid.Nil).Return(orderDtos, nil).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/orders", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusOK, gotStatusCode)
		assert.JSONEq(
			t,
			fmt.Sprintf(
				`[{"number": "%s", "uploaded_at": "%s", "accrual": %v, "status": "%s"}]`,
				expectedDto.Number,
				expectedDto.UploadedAt,
				*expectedDto.Accrual,
				expectedDto.Status), gotResponse)
	})

	t.Run("successful fetch with empty order list", func(t *testing.T) {
		orderService.On("GetUserOrders", mock.Anything, uuid.Nil).Return(nil, nil).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/orders", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusNoContent, gotStatusCode)
		assert.Equal(t, "", gotResponse)
	})

	t.Run("auth check", func(t *testing.T) {
		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/orders", http.NoBody, nil)
		assert.Equal(t, http.StatusUnauthorized, gotStatusCode)
		assert.Equal(t, "Unauthorized", gotResponse)
	})

	t.Run("500 error", func(t *testing.T) {
		orderService.On("GetUserOrders", mock.Anything, uuid.Nil).Return(nil, errors.New("something bad happened")).Once()

		gotStatusCode, gotResponse := testRequest(t, app, "GET", "/api/user/orders", http.NoBody, &nullIdToken)
		assert.Equal(t, http.StatusInternalServerError, gotStatusCode)
		assert.Equal(t, "Internal Server Error", gotResponse)
	})
}
