package okex

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	client     *http.Client
	apiBaseURL string
	projectID  string
	apiKey     string
	apiSecret  string
	passphrase string

	logger *zap.Logger
}

const defaultTimeout = 30 * time.Second

func NewClient(baseURL, projectID string, apiKey, apiSecret, passphrase string, opts ...ClientOption) IOKEXClient {
	c := &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		apiBaseURL: baseURL,
		projectID:  projectID,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		passphrase: passphrase,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) sign(timestamp, method, requestPath string, body []byte) string {
	message := timestamp + method + requestPath
	if len(body) > 0 {
		message += string(body)
	}

	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *Client) setHeaders(req *http.Request, method string, body []byte) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	sign := c.sign(timestamp, method, req.URL.RequestURI(), body)

	req.Header.Set("OK-ACCESS-KEY", c.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", sign)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", c.passphrase)
	req.Header.Set("OK-ACCESS-PROJECT", c.projectID)
	req.Header.Set("Content-Type", "application/json")

	if c.logger != nil {
		c.logger.Debug("okex_client_request",
			zap.String("method", method),
			zap.String("requestURI", req.URL.RequestURI()),
			zap.Int("body_length", len(body)),
			zap.String("timestamp", timestamp),
		)
	}
}

func (c *Client) BroadcastTransaction(ctx context.Context, params BroadcastTransactionParams) (*BroadcastTransactionResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid broadcast transaction parameters: %w", err)
	}

	url := fmt.Sprintf("%s/api/v5/wallet/pre-transaction/broadcast-transaction", c.apiBaseURL)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, "POST", body)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("okex_broadcast_transaction_response",
			zap.Int("status_code", resp.StatusCode),
			zap.String("chain_index", params.ChainIndex),
			zap.String("address", params.Address),
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("broadcast transaction failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var broadcastResp BroadcastTransactionResponse
	if err := json.Unmarshal(respBody, &broadcastResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &broadcastResp, nil
}

func (c *Client) GetOrders(ctx context.Context, params QueryOrdersParams) (*QueryOrdersResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query parameters: %w", err)
	}

	url := fmt.Sprintf("%s/api/v5/wallet/post-transaction/orders", c.apiBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if params.Address != "" {
		q.Add("address", params.Address)
	}
	if params.AccountID != "" {
		q.Add("accountId", params.AccountID)
	}
	if params.ChainIndex != "" {
		q.Add("chainIndex", params.ChainIndex)
	}
	if params.TxStatus != "" {
		q.Add("txStatus", params.TxStatus)
	}
	if params.OrderID != "" {
		q.Add("orderId", params.OrderID)
	}
	if params.Cursor != "" {
		q.Add("cursor", params.Cursor)
	}
	if params.Limit != "" {
		q.Add("limit", params.Limit)
	} else {
		q.Add("limit", DefaultOrderLimit) // Use default limit
	}

	req.URL.RawQuery = q.Encode()
	c.setHeaders(req, "GET", nil)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("okex_get_orders_response",
			zap.Int("status_code", resp.StatusCode),
			zap.String("address", params.Address),
			zap.String("chain_index", params.ChainIndex),
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get orders request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var ordersResp QueryOrdersResponse
	if err := json.Unmarshal(respBody, &ordersResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ordersResp, nil
}

func (c *Client) GetSupportedChains(ctx context.Context, chainID string) (*SupportedChainResponse, error) {
	url := fmt.Sprintf("%s/api/v5/dex/aggregator/supported/chain", c.apiBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if chainID != "" {
		q.Add("chainId", chainID)
	}
	req.URL.RawQuery = q.Encode()

	c.setHeaders(req, "GET", nil)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported chains: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("okex_supported_chains_response",
			zap.Int("status_code", resp.StatusCode),
			zap.String("chain_id", chainID),
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supported chains request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var chainResp SupportedChainResponse
	if err := json.Unmarshal(respBody, &chainResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chainResp, nil
}

func (c *Client) GetApproveTransaction(ctx context.Context, params ApproveTransactionParams) (*ApproveTransactionResponse, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid approve transaction parameters: %w", err)
	}

	url := fmt.Sprintf("%s/api/v5/dex/aggregator/approve-transaction", c.apiBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("chainId", params.ChainID)
	q.Add("tokenContractAddress", params.TokenContractAddress)
	q.Add("approveAmount", params.ApproveAmount)
	req.URL.RawQuery = q.Encode()

	c.setHeaders(req, "GET", nil)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get approve transaction: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("okex_approve_transaction_response",
			zap.Int("status_code", resp.StatusCode),
			zap.String("chain_id", params.ChainID),
			zap.String("token_address", params.TokenContractAddress),
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("approve transaction request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var approveResp ApproveTransactionResponse
	if err := json.Unmarshal(respBody, &approveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &approveResp, nil
}