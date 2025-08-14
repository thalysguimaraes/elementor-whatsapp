package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
)

// Client represents a Cloudflare D1 database client
type Client struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new D1 database client
func NewClient(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	baseURL := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s",
		cfg.Cloudflare.AccountID,
		cfg.Cloudflare.DatabaseID,
	)

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}, nil
}

// Query executes a SQL query against the D1 database
func (c *Client) Query(sql string, params ...interface{}) (*D1Result, error) {
	req := QueryRequest{
		SQL:    sql,
		Params: params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/query", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.config.Cloudflare.APIToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var d1Resp D1Response
	if err := json.Unmarshal(respBody, &d1Resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !d1Resp.Success || len(d1Resp.Errors) > 0 {
		if len(d1Resp.Errors) > 0 {
			return nil, fmt.Errorf("D1 error: %s", d1Resp.Errors[0].Message)
		}
		return nil, fmt.Errorf("D1 query failed")
	}

	if len(d1Resp.Result) == 0 {
		return nil, fmt.Errorf("no results returned")
	}

	return &d1Resp.Result[0], nil
}

// GetStats retrieves dashboard statistics
func (c *Client) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Get form counts
	result, err := c.Query("SELECT COUNT(*) as count FROM forms")
	if err != nil {
		log.Error("Failed to get form count", "error", err)
	} else if len(result.Results) > 0 {
		if count, ok := result.Results[0]["count"].(float64); ok {
			stats.TotalForms = int(count)
			stats.ActiveForms = int(count) // TODO: Add active status field
		}
	}

	// Get contact count
	result, err = c.Query("SELECT COUNT(*) as count FROM contacts")
	if err != nil {
		log.Error("Failed to get contact count", "error", err)
	} else if len(result.Results) > 0 {
		if count, ok := result.Results[0]["count"].(float64); ok {
			stats.TotalContacts = int(count)
		}
	}

	// Get webhook counts for today
	today := time.Now().Format("2006-01-02")
	result, err = c.Query(
		"SELECT COUNT(*) as count FROM webhook_logs WHERE DATE(created_at) = ?",
		today,
	)
	if err != nil {
		log.Error("Failed to get webhook count", "error", err)
	} else if len(result.Results) > 0 {
		if count, ok := result.Results[0]["count"].(float64); ok {
			stats.WebhooksToday = int(count)
		}
	}

	// Get last webhook time
	result, err = c.Query("SELECT created_at FROM webhook_logs ORDER BY created_at DESC LIMIT 1")
	if err != nil {
		log.Error("Failed to get last webhook", "error", err)
	} else if len(result.Results) > 0 {
		if timestamp, ok := result.Results[0]["created_at"].(string); ok {
			stats.LastWebhook, _ = time.Parse(time.RFC3339, timestamp)
		}
	}

	// Check connection status (simple ping)
	_, err = c.Query("SELECT 1")
	if err != nil {
		stats.ConnectionStatus = "Disconnected"
	} else {
		stats.ConnectionStatus = "Connected"
	}

	return stats, nil
}

// InitSchema initializes the database schema
func (c *Client) InitSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS forms (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS form_fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			form_id TEXT NOT NULL,
			elementor_id TEXT NOT NULL,
			label TEXT NOT NULL,
			type TEXT DEFAULT 'text',
			required BOOLEAN DEFAULT 0,
			position INTEGER DEFAULT 0,
			FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone_number TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			company TEXT,
			role TEXT,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS form_numbers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			form_id TEXT NOT NULL,
			phone_number TEXT NOT NULL,
			label TEXT,
			contact_id INTEGER,
			FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE,
			FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS webhook_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			form_id TEXT,
			status TEXT,
			request TEXT,
			response TEXT,
			duration_ms INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
		)`,
	}

	for _, stmt := range statements {
		if _, err := c.Query(stmt); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	log.Info("Database schema initialized successfully")
	return nil
}