package adminservices

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ExchangeRateService handles currency API requests
type ExchangeRateService struct{}

// NewExchangeRateService creates a new ExchangeRateService
func NewExchangeRateService() *ExchangeRateService {
	return &ExchangeRateService{}
}

// GetCurrencyAPI calls the external FX rates API with a base currency
func (c *ExchangeRateService) GetCurrencyAPI(baseCode string) (interface{}, error) {
	if baseCode == "" {
		return nil, fmt.Errorf("base currency code cannot be empty")
	}

	// Construct the API URL
	url := fmt.Sprintf("https://api.fxratesapi.com/latest?base=%s", baseCode)

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call FX API: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 HTTP response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FX API error: %s", string(body))
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Optional: validate if API returned an error for invalid currency code
	if _, ok := result["error"]; ok {
		return nil, fmt.Errorf("invalid currency code: %s", baseCode)
	}

	return result, nil
}
