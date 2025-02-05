package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/Kopleman/gophermart/internal/common/log"
)

const baseBackoffMultiplier = 2

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(baseBackoffMultiplier, float64(retries))) * time.Second
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode >= http.StatusInternalServerError {
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
	for shouldRetry(err, resp) && retries < t.retryCount {
		time.Sleep(backoff(retries))
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
	transport  http.RoundTripper
	logger     log.Logger
	retryCount int
}

func NewRetryableTransport(logger log.Logger, retryCount int) http.RoundTripper {
	transport := &retryableTransport{
		transport:  &http.Transport{},
		retryCount: retryCount,
		logger:     logger,
	}

	return transport
}
