package database

import (
	"time"
)

// Form represents a webhook form configuration
type Form struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Fields      []Field   `json:"fields"`
	Numbers     []Number  `json:"numbers"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Field represents a form field mapping
type Field struct {
	ID          string `json:"id"`
	FormID      string `json:"form_id"`
	ElementorID string `json:"elementor_id"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Position    int    `json:"position"`
}

// Number represents a WhatsApp number recipient
type Number struct {
	ID          int    `json:"id"`
	FormID      string `json:"form_id"`
	PhoneNumber string `json:"phone_number"`
	Label       string `json:"label"`
	ContactID   *int   `json:"contact_id,omitempty"`
}

// Contact represents a contact in the system
type Contact struct {
	ID          int       `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Name        string    `json:"name"`
	Company     string    `json:"company,omitempty"`
	Role        string    `json:"role,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FormWithStats includes form with additional statistics
type FormWithStats struct {
	Form
	FieldCount  int `json:"field_count"`
	NumberCount int `json:"number_count"`
}

// ContactWithStats includes contact with usage statistics
type ContactWithStats struct {
	Contact
	FormCount int      `json:"form_count"`
	FormIDs   []string `json:"form_ids"`
}

// WebhookLog represents a webhook execution log
type WebhookLog struct {
	ID         int       `json:"id"`
	FormID     string    `json:"form_id"`
	Status     string    `json:"status"`
	Request    string    `json:"request"`
	Response   string    `json:"response"`
	Duration   int       `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// Stats represents dashboard statistics
type Stats struct {
	TotalForms       int       `json:"total_forms"`
	ActiveForms      int       `json:"active_forms"`
	TotalContacts    int       `json:"total_contacts"`
	WebhooksToday    int       `json:"webhooks_today"`
	WebhooksThisWeek int       `json:"webhooks_week"`
	LastWebhook      time.Time `json:"last_webhook"`
	ConnectionStatus string    `json:"connection_status"`
}

// D1Response represents the response from Cloudflare D1 API
type D1Response struct {
	Result []D1Result `json:"result"`
	Success bool      `json:"success"`
	Errors  []D1Error `json:"errors"`
}

// D1Result represents a single query result from D1
type D1Result struct {
	Results []map[string]interface{} `json:"results"`
	Success bool                     `json:"success"`
	Meta    D1Meta                   `json:"meta"`
}

// D1Meta contains metadata about the query execution
type D1Meta struct {
	Duration     float64 `json:"duration"`
	LastRowID    int64   `json:"last_row_id"`
	RowsAffected int     `json:"rows_affected"`
	RowsRead     int     `json:"rows_read"`
	RowsWritten  int     `json:"rows_written"`
}

// D1Error represents an error from the D1 API
type D1Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// QueryRequest represents a request to the D1 API
type QueryRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}