package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Kopleman/gophermart/internal/common/log"
)

const baseBackoffMultiplier = 2
const baseBackoffMaxWait = 60

func (t *retryableTransport) backoff(resp *http.Response, retries int) time.Duration {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests && t.handle429Status {
			retryAfter := resp.Header.Get("Retry-After")
			retryAfterSeconds, err := strconv.Atoi(retryAfter)
			if err != nil {
				retryAfterSeconds = baseBackoffMaxWait
			}
			return time.Duration(retryAfterSeconds) * time.Second
		}
	}
	return time.Duration(math.Pow(baseBackoffMultiplier, float64(retries))) * time.Second
}

func (t *retryableTransport) shouldRetry(err error, resp *http.Response, retries int) bool {
	if err != nil {
		return t.retryCountExceed(retries)
	}

	if resp == nil {
		return false
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return t.retryCountExceed(retries)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return t.handle429Status
	}

	return false
}

func (t *retryableTransport) retryCountExceed(retries int) bool {
	if t.retryCount > retries {
		return true
	}

	return false
}

func closeBody(resp *http.Response) error {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("failed to close response body: %w", err)
		}
	}
	return nil
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var initialBodyBytes []byte
	if req.Body != nil {
		reRedBodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		initialBodyBytes = reRedBodyBytes
		req.Body = io.NopCloser(bytes.NewBuffer(initialBodyBytes))
	}
	resp, err := t.transport.RoundTrip(req)
	if err == nil {
		return resp, nil
	}
	retries := 0
	for t.shouldRetry(err, resp, retries) {
		time.Sleep(t.backoff(resp, retries))
		t.logger.Infof("retry attempt %d", retries+1)
		if err = closeBody(resp); err != nil {
			return nil, fmt.Errorf("round trip: %w", err)
		}
		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(initialBodyBytes))
		}
		resp, err = t.transport.RoundTrip(req)
		if err == nil {
			return resp, nil
		}
		retries++
	}
	return resp, fmt.Errorf("retry amount exeeded: %w", err)
}

type retryableTransport struct {
	transport       http.RoundTripper
	logger          log.Logger
	retryCount      int
	handle429Status bool
}

func NewRetryableTransport(logger log.Logger, retryCount int, handle429Status bool) http.RoundTripper {
	transport := &retryableTransport{
		transport:       &http.Transport{},
		retryCount:      retryCount,
		logger:          logger,
		handle429Status: handle429Status,
	}

	return transport
}
