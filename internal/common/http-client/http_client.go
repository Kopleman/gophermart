package httpclient

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/Kopleman/gophermart/internal/common"
	"github.com/Kopleman/gophermart/internal/common/log"
)

func (c *HTTPClient) Post(url, contentType string, bodyBytes []byte) ([]byte, error) {
	body := bytes.NewBuffer(bodyBytes)
	finalURL := c.BaseURL + url
	var respBody []byte

	req, err := http.NewRequest(http.MethodPost, finalURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set(common.ContentType, contentType)
	req.Header.Set(common.AcceptEncoding, "gzip")

	res, respErr := c.client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to send post req to '%s': %w", finalURL, respErr)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("failed to send post req to '%s': status code %d", finalURL, res.StatusCode)
	}

	gz, gzipErr := gzip.NewReader(res.Body)
	if gzipErr != nil {
		return nil, fmt.Errorf("failed to decompress response: %w", err)
	}
	defer func() {
		if gzErr := gz.Close(); gzErr != nil {
			c.logger.Error(gzErr)
		}
	}()

	defer func() {
		if bodyParseErr := res.Body.Close(); bodyParseErr != nil {
			c.logger.Error(bodyParseErr)
		}
	}()

	respBody, err = io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return respBody, nil
}

func (c *HTTPClient) Get(url, contentType string) ([]byte, error) {
	finalURL := c.BaseURL + url
	var respBody []byte

	req, err := http.NewRequest(http.MethodGet, finalURL, bytes.NewBuffer(make([]byte, 0)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set(common.ContentType, contentType)

	res, respErr := c.client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("failed to send post req to '%s': %w", finalURL, respErr)
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("failed to send post req to '%s': status code %d", finalURL, res.StatusCode)
	}

	defer func() {
		if bodyParseErr := res.Body.Close(); bodyParseErr != nil {
			c.logger.Error(bodyParseErr)
		}
	}()

	respBody, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return respBody, nil
}

type HTTPClient struct {
	logger  log.Logger
	client  *http.Client
	BaseURL string
	key     []byte
}

const defaultRetryCount = 3

func NewHTTPClient(baseURL string, logger log.Logger, handle429Status bool) *HTTPClient {
	transport := NewRetryableTransport(logger, defaultRetryCount, handle429Status)

	return &HTTPClient{
		BaseURL: baseURL,
		client: &http.Client{
			Transport: transport,
		},
		logger: logger,
		key:    []byte(""),
	}
}
