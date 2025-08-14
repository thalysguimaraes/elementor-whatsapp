package database

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// GetAllContacts retrieves all contacts
func (c *Client) GetAllContacts() ([]Contact, error) {
	query := `
		SELECT * FROM contacts
		ORDER BY name ASC
	`

	result, err := c.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	var contacts []Contact
	for _, row := range result.Results {
		contact := Contact{}
		
		if id, ok := row["id"].(float64); ok {
			contact.ID = int(id)
		}
		if phone, ok := row["phone_number"].(string); ok {
			contact.PhoneNumber = phone
		}
		if name, ok := row["name"].(string); ok {
			contact.Name = name
		}
		if company, ok := row["company"].(string); ok {
			contact.Company = company
		}
		if role, ok := row["role"].(string); ok {
			contact.Role = role
		}
		if notes, ok := row["notes"].(string); ok {
			contact.Notes = notes
		}
		if createdAt, ok := row["created_at"].(string); ok {
			contact.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt, ok := row["updated_at"].(string); ok {
			contact.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContactByID retrieves a single contact by ID
func (c *Client) GetContactByID(id int) (*Contact, error) {
	query := "SELECT * FROM contacts WHERE id = ?"
	result, err := c.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	row := result.Results[0]
	contact := &Contact{}
	
	if id, ok := row["id"].(float64); ok {
		contact.ID = int(id)
	}
	if phone, ok := row["phone_number"].(string); ok {
		contact.PhoneNumber = phone
	}
	if name, ok := row["name"].(string); ok {
		contact.Name = name
	}
	if company, ok := row["company"].(string); ok {
		contact.Company = company
	}
	if role, ok := row["role"].(string); ok {
		contact.Role = role
	}
	if notes, ok := row["notes"].(string); ok {
		contact.Notes = notes
	}
	if createdAt, ok := row["created_at"].(string); ok {
		contact.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	}
	if updatedAt, ok := row["updated_at"].(string); ok {
		contact.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	}

	return contact, nil
}

// GetContactsWithStats retrieves all contacts with usage statistics
func (c *Client) GetContactsWithStats() ([]ContactWithStats, error) {
	query := `
		SELECT 
			c.*,
			COUNT(DISTINCT fn.form_id) as form_count,
			GROUP_CONCAT(DISTINCT fn.form_id) as form_ids
		FROM contacts c
		LEFT JOIN form_numbers fn ON c.id = fn.contact_id
		GROUP BY c.id
		ORDER BY c.name ASC
	`

	result, err := c.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts with stats: %w", err)
	}

	var contacts []ContactWithStats
	for _, row := range result.Results {
		contact := ContactWithStats{}
		
		if id, ok := row["id"].(float64); ok {
			contact.ID = int(id)
		}
		if phone, ok := row["phone_number"].(string); ok {
			contact.PhoneNumber = phone
		}
		if name, ok := row["name"].(string); ok {
			contact.Name = name
		}
		if company, ok := row["company"].(string); ok {
			contact.Company = company
		}
		if role, ok := row["role"].(string); ok {
			contact.Role = role
		}
		if notes, ok := row["notes"].(string); ok {
			contact.Notes = notes
		}
		if createdAt, ok := row["created_at"].(string); ok {
			contact.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt, ok := row["updated_at"].(string); ok {
			contact.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if count, ok := row["form_count"].(float64); ok {
			contact.FormCount = int(count)
		}
		if formIDs, ok := row["form_ids"].(string); ok && formIDs != "" {
			contact.FormIDs = strings.Split(formIDs, ",")
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContact retrieves a single contact by ID
func (c *Client) GetContact(id int) (*Contact, error) {
	query := "SELECT * FROM contacts WHERE id = ?"
	result, err := c.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	row := result.Results[0]
	contact := &Contact{}

	if id, ok := row["id"].(float64); ok {
		contact.ID = int(id)
	}
	if phone, ok := row["phone_number"].(string); ok {
		contact.PhoneNumber = phone
	}
	if name, ok := row["name"].(string); ok {
		contact.Name = name
	}
	if company, ok := row["company"].(string); ok {
		contact.Company = company
	}
	if role, ok := row["role"].(string); ok {
		contact.Role = role
	}
	if notes, ok := row["notes"].(string); ok {
		contact.Notes = notes
	}
	if createdAt, ok := row["created_at"].(string); ok {
		contact.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	}
	if updatedAt, ok := row["updated_at"].(string); ok {
		contact.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	}

	return contact, nil
}

// CreateContact creates a new contact
func (c *Client) CreateContact(contact *Contact) (int, error) {
	query := `
		INSERT INTO contacts (phone_number, name, company, role, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	
	result, err := c.Query(query, contact.PhoneNumber, contact.Name, contact.Company, contact.Role, contact.Notes)
	if err != nil {
		return 0, fmt.Errorf("failed to create contact: %w", err)
	}

	return int(result.Meta.LastRowID), nil
}

// UpdateContact updates an existing contact
func (c *Client) UpdateContact(contact *Contact) error {
	query := `
		UPDATE contacts 
		SET phone_number = ?, name = ?, company = ?, role = ?, notes = ?, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	_, err := c.Query(query, contact.PhoneNumber, contact.Name, contact.Company, contact.Role, contact.Notes, contact.ID)
	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	return nil
}

// DeleteContact deletes a contact
func (c *Client) DeleteContact(id int) error {
	// First, remove references from form_numbers
	_, err := c.Query("UPDATE form_numbers SET contact_id = NULL WHERE contact_id = ?", id)
	if err != nil {
		log.Warn("Failed to remove contact references", "error", err)
	}

	// Delete the contact
	query := "DELETE FROM contacts WHERE id = ?"
	_, err = c.Query(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}

// SearchContacts searches for contacts by name, company, or phone
func (c *Client) SearchContacts(searchTerm string) ([]ContactWithStats, error) {
	query := `
		SELECT 
			c.*,
			COUNT(DISTINCT fn.form_id) as form_count,
			GROUP_CONCAT(DISTINCT fn.form_id) as form_ids
		FROM contacts c
		LEFT JOIN form_numbers fn ON c.id = fn.contact_id
		WHERE c.name LIKE ? OR c.company LIKE ? OR c.phone_number LIKE ?
		GROUP BY c.id
		ORDER BY c.name ASC
	`

	searchPattern := "%" + searchTerm + "%"
	result, err := c.Query(query, searchPattern, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}

	var contacts []ContactWithStats
	for _, row := range result.Results {
		contact := ContactWithStats{}
		
		if id, ok := row["id"].(float64); ok {
			contact.ID = int(id)
		}
		if phone, ok := row["phone_number"].(string); ok {
			contact.PhoneNumber = phone
		}
		if name, ok := row["name"].(string); ok {
			contact.Name = name
		}
		if company, ok := row["company"].(string); ok {
			contact.Company = company
		}
		if role, ok := row["role"].(string); ok {
			contact.Role = role
		}
		if notes, ok := row["notes"].(string); ok {
			contact.Notes = notes
		}
		if createdAt, ok := row["created_at"].(string); ok {
			contact.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt, ok := row["updated_at"].(string); ok {
			contact.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if count, ok := row["form_count"].(float64); ok {
			contact.FormCount = int(count)
		}
		if formIDs, ok := row["form_ids"].(string); ok && formIDs != "" {
			contact.FormIDs = strings.Split(formIDs, ",")
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContactsByForm retrieves all contacts associated with a form
func (c *Client) GetContactsByForm(formID string) ([]Contact, error) {
	query := `
		SELECT DISTINCT c.* 
		FROM contacts c
		JOIN form_numbers fn ON c.id = fn.contact_id
		WHERE fn.form_id = ?
		ORDER BY c.name ASC
	`

	result, err := c.Query(query, formID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts by form: %w", err)
	}

	var contacts []Contact
	for _, row := range result.Results {
		contact := Contact{}
		
		if id, ok := row["id"].(float64); ok {
			contact.ID = int(id)
		}
		if phone, ok := row["phone_number"].(string); ok {
			contact.PhoneNumber = phone
		}
		if name, ok := row["name"].(string); ok {
			contact.Name = name
		}
		if company, ok := row["company"].(string); ok {
			contact.Company = company
		}
		if role, ok := row["role"].(string); ok {
			contact.Role = role
		}
		if notes, ok := row["notes"].(string); ok {
			contact.Notes = notes
		}
		if createdAt, ok := row["created_at"].(string); ok {
			contact.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt, ok := row["updated_at"].(string); ok {
			contact.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// ExportContactsCSV exports all contacts as CSV
func (c *Client) ExportContactsCSV() ([]byte, error) {
	contacts, err := c.GetContactsWithStats()
	if err != nil {
		return nil, err
	}

	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Name", "Phone Number", "Company", "Role", "Notes", "Form Count", "Created At"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, contact := range contacts {
		record := []string{
			fmt.Sprintf("%d", contact.ID),
			contact.Name,
			contact.PhoneNumber,
			contact.Company,
			contact.Role,
			contact.Notes,
			fmt.Sprintf("%d", contact.FormCount),
			contact.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return []byte(buf.String()), nil
}

// ImportContactsCSV imports contacts from CSV data
func (c *Client) ImportContactsCSV(data []byte) (int, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create a map of header positions
	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Read and import records
	imported := 0
	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file
		}

		contact := Contact{}

		// Map CSV fields to contact fields
		if idx, ok := headerMap["name"]; ok && idx < len(record) {
			contact.Name = record[idx]
		}
		if idx, ok := headerMap["phone number"]; ok && idx < len(record) {
			contact.PhoneNumber = record[idx]
		} else if idx, ok := headerMap["phone"]; ok && idx < len(record) {
			contact.PhoneNumber = record[idx]
		}
		if idx, ok := headerMap["company"]; ok && idx < len(record) {
			contact.Company = record[idx]
		}
		if idx, ok := headerMap["role"]; ok && idx < len(record) {
			contact.Role = record[idx]
		}
		if idx, ok := headerMap["notes"]; ok && idx < len(record) {
			contact.Notes = record[idx]
		}

		// Skip if required fields are missing
		if contact.Name == "" || contact.PhoneNumber == "" {
			continue
		}

		// Try to create the contact
		if _, err := c.CreateContact(&contact); err != nil {
			log.Warn("Failed to import contact", "name", contact.Name, "error", err)
			continue
		}

		imported++
	}

	return imported, nil
}