package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a webhook testing client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new webhook testing client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// TestRequest represents a webhook test request
type TestRequest struct {
	FormID  string
	Format  string // "json" or "form"
	Fields  map[string]string
}

// TestResponse represents a webhook test response
type TestResponse struct {
	StatusCode   int
	Headers      map[string][]string
	Body         string
	Duration     time.Duration
	Error        error
}

// TestWebhook sends a test webhook request
func (c *Client) TestWebhook(req *TestRequest) *TestResponse {
	startTime := time.Now()
	
	// Build the URL
	webhookURL := fmt.Sprintf("%s/webhook/%s", c.baseURL, req.FormID)
	
	// Prepare the request based on format
	var httpReq *http.Request
	var err error
	
	if req.Format == "json" {
		// JSON format
		body, _ := json.Marshal(req.Fields)
		httpReq, err = http.NewRequest("POST", webhookURL, bytes.NewReader(body))
		if err != nil {
			return &TestResponse{
				Error:    fmt.Errorf("failed to create request: %w", err),
				Duration: time.Since(startTime),
			}
		}
		httpReq.Header.Set("Content-Type", "application/json")
	} else {
		// Form-encoded format
		formData := url.Values{}
		for key, value := range req.Fields {
			// Elementor sends nested fields
			formData.Set(fmt.Sprintf("fields[%s][value]", key), value)
		}
		httpReq, err = http.NewRequest("POST", webhookURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return &TestResponse{
				Error:    fmt.Errorf("failed to create request: %w", err),
				Duration: time.Since(startTime),
			}
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	
	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &TestResponse{
			Error:    fmt.Errorf("request failed: %w", err),
			Duration: time.Since(startTime),
		}
	}
	defer resp.Body.Close()
	
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Error:      fmt.Errorf("failed to read response: %w", err),
			Duration:   time.Since(startTime),
		}
	}
	
	return &TestResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(body),
		Duration:   time.Since(startTime),
	}
}

// GenerateSampleData generates sample data for testing
func GenerateSampleData() map[string]string {
	return map[string]string{
		"name":    "John Doe",
		"email":   "john.doe@example.com",
		"phone":   "+1 234 567 8900",
		"company": "Acme Corp",
		"message": "This is a test message from the webhook testing tool.",
	}
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// FormatHeaders formats HTTP headers for display
func FormatHeaders(headers map[string][]string) string {
	var result strings.Builder
	for key, values := range headers {
		result.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(values, ", ")))
	}
	return result.String()
}

// PrettyJSON formats JSON for display
func PrettyJSON(jsonStr string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return jsonStr // Return as-is if not valid JSON
	}
	
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return jsonStr
	}
	
	return string(pretty)
}