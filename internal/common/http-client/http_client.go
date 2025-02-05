package httpclient

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

	bodyHash := c.calcHashForBody(bodyBytes)
	if bodyHash != "" {
		req.Header.Set(common.HashSHA256, bodyHash)
	}

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

func (c *HTTPClient) calcHashForBody(bodyBytes []byte) string {
	if len(c.key) == 0 {
		return ""
	}
	if len(bodyBytes) == 0 {
		return ""
	}

	h := hmac.New(sha256.New, c.key)
	h.Write(bodyBytes)
	hash := h.Sum(nil)
	hashString := base64.StdEncoding.EncodeToString(hash)

	return hashString
}

type HTTPClient struct {
	logger  log.Logger
	client  *http.Client
	BaseURL string
	key     []byte
}

const defaultRetryCount = 3

func NewHTTPClient(baseURL string, logger log.Logger) *HTTPClient {
	transport := NewRetryableTransport(logger, defaultRetryCount)

	return &HTTPClient{
		BaseURL: baseURL,
		client: &http.Client{
			Transport: transport,
		},
		logger: logger,
		key:    []byte(""),
	}
}
