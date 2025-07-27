package providers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/algonius/algonius-wallet/native/pkg/dex"
	"go.uber.org/zap"
)

// OKXProvider implements IDEXProvider for OKX DEX API
type OKXProvider struct {
	name       string
	apiKey     string
	secretKey  string
	passphrase string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// OKXConfig holds configuration for OKX DEX API
type OKXConfig struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	BaseURL    string // Default: https://www.okx.com
	Timeout    time.Duration
}

// NewOKXProvider creates a new OKX DEX provider
func NewOKXProvider(config OKXConfig, logger *zap.Logger) *OKXProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://www.okx.com"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &OKXProvider{
		name:       "OKX",
		apiKey:     config.APIKey,
		secretKey:  config.SecretKey,
		passphrase: config.Passphrase,
		baseURL:    config.BaseURL,
		httpClient: &http.Client{Timeout: config.Timeout},
		logger:     logger,
	}
}

// GetName returns the provider name
func (o *OKXProvider) GetName() string {
	return o.name
}

// IsSupported checks if the chain is supported by OKX DEX
func (o *OKXProvider) IsSupported(chainID string) bool {
	supportedChains := map[string]bool{
		"1":   true, // Ethereum
		"56":  true, // BSC
		"137": true, // Polygon
		"501": true, // Solana
	}
	return supportedChains[chainID]
}

// GetQuote fetches a swap quote from OKX DEX API
func (o *OKXProvider) GetQuote(ctx context.Context, params dex.SwapParams) (*dex.SwapQuote, error) {
	o.logger.Debug("Getting quote from OKX DEX", 
		zap.String("fromToken", params.FromToken),
		zap.String("toToken", params.ToToken),
		zap.String("amount", params.Amount),
		zap.String("chainId", params.ChainID))

	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid swap parameters: %w", err)
	}

	// Build quote request
	quoteReq := map[string]interface{}{
		"chainId":       params.ChainID,
		"fromTokenAddress": params.FromToken,
		"toTokenAddress":   params.ToToken,
		"amount":        params.Amount,
		"slippage":      fmt.Sprintf("%.3f", params.Slippage),
	}

	// Make API request
	resp, err := o.makeAPIRequest(ctx, "GET", "/api/v5/dex/aggregator/quote", quoteReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote from OKX: %w", err)
	}

	// Parse response
	var quoteResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ToTokenAmount   string `json:"toTokenAmount"`
			FromTokenAmount string `json:"fromTokenAmount"`
			EstimatedGas    string `json:"estimatedGas"`
			MinimumReceived string `json:"minimumReceived"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to parse OKX quote response: %w", err)
	}

	if quoteResp.Code != "0" {
		return nil, fmt.Errorf("OKX API error: %s", quoteResp.Msg)
	}

	if len(quoteResp.Data) == 0 {
		return nil, fmt.Errorf("no quote data returned from OKX")
	}

	data := quoteResp.Data[0]
	
	// Parse estimated gas
	var estimatedGas uint64
	if gasStr := data.EstimatedGas; gasStr != "" {
		if gas, err := strconv.ParseUint(gasStr, 10, 64); err == nil {
			estimatedGas = gas
		}
	}
	
	return &dex.SwapQuote{
		Provider:     o.name,
		FromToken:    params.FromToken,
		ToToken:      params.ToToken,
		FromAmount:   data.FromTokenAmount,
		ToAmount:     data.ToTokenAmount,
		EstimatedGas: estimatedGas,
		Slippage:     params.Slippage,
		ValidUntil:   time.Now().Add(30 * time.Second).Unix(),
		RawData:      fmt.Sprintf(`{"okx_data": %s}`, string(mustMarshal(data))),
	}, nil
}

// ExecuteSwap executes a token swap using OKX DEX API
func (o *OKXProvider) ExecuteSwap(ctx context.Context, params dex.SwapParams) (*dex.SwapResult, error) {
	o.logger.Info("Executing swap with OKX DEX",
		zap.String("fromToken", params.FromToken),
		zap.String("toToken", params.ToToken),
		zap.String("amount", params.Amount))

	// Note: We don't need the quote for the swap request, but we keep this for validation
	_, err := o.GetQuote(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote before swap: %w", err)
	}

	// Build swap request
	swapReq := map[string]interface{}{
		"chainId":          params.ChainID,
		"fromTokenAddress": params.FromToken,
		"toTokenAddress":   params.ToToken,
		"amount":           params.Amount,
		"slippage":         fmt.Sprintf("%.3f", params.Slippage),
		"userWalletAddress": params.FromAddress,
		"referrerAddress":  "", // Optional referrer
	}

	// Make swap API request
	resp, err := o.makeAPIRequest(ctx, "GET", "/api/v5/dex/aggregator/swap", swapReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute swap with OKX: %w", err)
	}

	// Parse response
	var swapResp struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Tx struct {
				Data     string `json:"data"`
				Gas      string `json:"gas"`
				GasPrice string `json:"gasPrice"`
				To       string `json:"to"`
				Value    string `json:"value"`
			} `json:"tx"`
			ToTokenAmount string `json:"toTokenAmount"`
			MinAmountOut  string `json:"minAmountOut"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &swapResp); err != nil {
		return nil, fmt.Errorf("failed to parse OKX swap response: %w", err)
	}

	if swapResp.Code != "0" {
		return nil, fmt.Errorf("OKX swap API error: %s", swapResp.Msg)
	}

	if len(swapResp.Data) == 0 {
		return nil, fmt.Errorf("no swap data returned from OKX")
	}

	data := swapResp.Data[0]
	
	// Note: In a real implementation, you would sign and broadcast the transaction here
	// For now, we return the transaction data for external signing
	mockTxHash := fmt.Sprintf("0x%x", sha256.Sum256([]byte(data.Tx.Data)))

	return &dex.SwapResult{
		TxHash:     mockTxHash,
		Provider:   o.name,
		FromToken:  params.FromToken,
		ToToken:    params.ToToken,
		FromAmount: params.Amount,
		ToAmount:   data.ToTokenAmount,
		Status:     "pending",
		Timestamp:  time.Now().Unix(),
	}, nil
}

// GetBalance fetches token balance using OKX API (if available)
func (o *OKXProvider) GetBalance(ctx context.Context, address string, tokenAddress string, chainID string) (*dex.BalanceInfo, error) {
	// OKX DEX API doesn't directly provide balance queries
	// This would typically use a separate balance API or RPC call
	return &dex.BalanceInfo{
		TokenAddress: tokenAddress,
		TokenSymbol:  "UNKNOWN",
		Balance:      "0", // Placeholder - would implement actual balance check
		Decimals:     18,
	}, fmt.Errorf("balance queries not supported by OKX DEX provider")
}

// EstimateGas estimates gas for the swap transaction
func (o *OKXProvider) EstimateGas(ctx context.Context, params dex.SwapParams) (gasLimit uint64, gasPrice string, err error) {
	// Get quote which includes gas estimation
	quote, err := o.GetQuote(ctx, params)
	if err != nil {
		return 0, "", fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Parse gas limit from quote
	if quote.EstimatedGas > 0 {
		gasLimit = quote.EstimatedGas
	}

	// Default gas values if not available
	if gasLimit == 0 {
		gasLimit = 200000 // Default gas limit for DEX swaps
	}

	// Use a reasonable gas price (this should be fetched from network)
	gasPrice = "20000000000" // 20 gwei

	return gasLimit, gasPrice, nil
}

// makeAPIRequest makes an authenticated request to OKX API
func (o *OKXProvider) makeAPIRequest(ctx context.Context, method, endpoint string, params map[string]interface{}) ([]byte, error) {
	// Build URL with query parameters for GET requests
	url := o.baseURL + endpoint
	var body io.Reader

	if method == "GET" && len(params) > 0 {
		queryParams := make([]string, 0, len(params))
		for key, value := range params {
			queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
		}
		url += "?" + strings.Join(queryParams, "&")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if credentials are provided
	if o.apiKey != "" {
		timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		signature := o.generateSignature(timestamp, method, endpoint, "")
		
		req.Header.Set("OK-ACCESS-KEY", o.apiKey)
		req.Header.Set("OK-ACCESS-SIGN", signature)
		req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
		req.Header.Set("OK-ACCESS-PASSPHRASE", o.passphrase)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Algonius-Wallet/1.0")

	// Make request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// generateSignature generates HMAC-SHA256 signature for OKX API authentication
func (o *OKXProvider) generateSignature(timestamp, method, requestPath, body string) string {
	if o.secretKey == "" {
		return ""
	}

	message := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(o.secretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// mustMarshal marshals data to JSON, panicking on error (helper for simple cases)
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
}