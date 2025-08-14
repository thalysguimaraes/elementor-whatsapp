package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

// GetAllForms retrieves all forms with their statistics
func (c *Client) GetAllForms() ([]FormWithStats, error) {
	query := `
		SELECT 
			f.id,
			f.name,
			f.description,
			f.created_at,
			f.updated_at,
			COUNT(DISTINCT ff.id) as field_count,
			COUNT(DISTINCT fn.id) as number_count
		FROM forms f
		LEFT JOIN form_fields ff ON f.id = ff.form_id
		LEFT JOIN form_numbers fn ON f.id = fn.form_id
		GROUP BY f.id
		ORDER BY f.created_at DESC
	`

	result, err := c.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get forms: %w", err)
	}

	var forms []FormWithStats
	for _, row := range result.Results {
		form := FormWithStats{}
		
		if id, ok := row["id"].(string); ok {
			form.ID = id
		}
		if name, ok := row["name"].(string); ok {
			form.Name = name
		}
		if desc, ok := row["description"].(string); ok {
			form.Description = desc
		}
		if createdAt, ok := row["created_at"].(string); ok {
			form.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt, ok := row["updated_at"].(string); ok {
			form.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if count, ok := row["field_count"].(float64); ok {
			form.FieldCount = int(count)
		}
		if count, ok := row["number_count"].(float64); ok {
			form.NumberCount = int(count)
		}

		forms = append(forms, form)
	}

	return forms, nil
}

// GetForm retrieves a single form by ID with all its details
func (c *Client) GetForm(id string) (*Form, error) {
	// Get form basic info
	query := "SELECT * FROM forms WHERE id = ?"
	result, err := c.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("form not found")
	}

	row := result.Results[0]
	form := &Form{}

	if id, ok := row["id"].(string); ok {
		form.ID = id
	}
	if name, ok := row["name"].(string); ok {
		form.Name = name
	}
	if desc, ok := row["description"].(string); ok {
		form.Description = desc
	}
	if createdAt, ok := row["created_at"].(string); ok {
		form.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	}
	if updatedAt, ok := row["updated_at"].(string); ok {
		form.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	}

	// Get fields
	fields, err := c.getFormFields(id)
	if err != nil {
		log.Error("Failed to get form fields", "error", err)
	}
	form.Fields = fields

	// Get numbers
	numbers, err := c.getFormNumbers(id)
	if err != nil {
		log.Error("Failed to get form numbers", "error", err)
	}
	form.Numbers = numbers

	return form, nil
}

// GetFormByID is an alias for GetForm for consistency
func (c *Client) GetFormByID(id string) (*Form, error) {
	return c.GetForm(id)
}

// CreateForm creates a new form with its fields and numbers
func (c *Client) CreateForm(form *Form) error {
	// Insert form
	query := `
		INSERT INTO forms (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	
	_, err := c.Query(query, form.ID, form.Name, form.Description)
	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	// Insert fields
	for i, field := range form.Fields {
		field.FormID = form.ID
		field.Position = i
		if err := c.createFormField(&field); err != nil {
			log.Error("Failed to create field", "error", err)
		}
	}

	// Insert numbers
	for _, number := range form.Numbers {
		number.FormID = form.ID
		if err := c.createFormNumber(&number); err != nil {
			log.Error("Failed to create number", "error", err)
		}
	}

	return nil
}

// UpdateForm updates an existing form
func (c *Client) UpdateForm(form *Form) error {
	// Update form
	query := `
		UPDATE forms 
		SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	_, err := c.Query(query, form.Name, form.Description, form.ID)
	if err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	// Delete existing fields and numbers
	_, _ = c.Query("DELETE FROM form_fields WHERE form_id = ?", form.ID)
	_, _ = c.Query("DELETE FROM form_numbers WHERE form_id = ?", form.ID)

	// Re-insert fields
	for i, field := range form.Fields {
		field.FormID = form.ID
		field.Position = i
		if err := c.createFormField(&field); err != nil {
			log.Error("Failed to create field", "error", err)
		}
	}

	// Re-insert numbers
	for _, number := range form.Numbers {
		number.FormID = form.ID
		if err := c.createFormNumber(&number); err != nil {
			log.Error("Failed to create number", "error", err)
		}
	}

	return nil
}

// DeleteForm deletes a form and all its related data
func (c *Client) DeleteForm(id string) error {
	query := "DELETE FROM forms WHERE id = ?"
	_, err := c.Query(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}
	return nil
}

// Helper functions

func (c *Client) getFormFields(formID string) ([]Field, error) {
	query := `
		SELECT * FROM form_fields 
		WHERE form_id = ? 
		ORDER BY position
	`
	
	result, err := c.Query(query, formID)
	if err != nil {
		return nil, err
	}

	var fields []Field
	for _, row := range result.Results {
		field := Field{FormID: formID}
		
		if id, ok := row["id"].(float64); ok {
			field.ID = fmt.Sprintf("%d", int(id))
		}
		if elementorID, ok := row["elementor_id"].(string); ok {
			field.ElementorID = elementorID
		}
		if label, ok := row["label"].(string); ok {
			field.Label = label
		}
		if fieldType, ok := row["type"].(string); ok {
			field.Type = fieldType
		}
		if required, ok := row["required"].(float64); ok {
			field.Required = required > 0
		}
		if position, ok := row["position"].(float64); ok {
			field.Position = int(position)
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func (c *Client) getFormNumbers(formID string) ([]Number, error) {
	query := `
		SELECT * FROM form_numbers 
		WHERE form_id = ?
	`
	
	result, err := c.Query(query, formID)
	if err != nil {
		return nil, err
	}

	var numbers []Number
	for _, row := range result.Results {
		number := Number{FormID: formID}
		
		if id, ok := row["id"].(float64); ok {
			number.ID = int(id)
		}
		if phone, ok := row["phone_number"].(string); ok {
			number.PhoneNumber = phone
		}
		if label, ok := row["label"].(string); ok {
			number.Label = label
		}
		if contactID, ok := row["contact_id"].(float64); ok {
			id := int(contactID)
			number.ContactID = &id
		}

		numbers = append(numbers, number)
	}

	return numbers, nil
}

func (c *Client) createFormField(field *Field) error {
	query := `
		INSERT INTO form_fields (form_id, elementor_id, label, type, required, position)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	required := 0
	if field.Required {
		required = 1
	}
	
	_, err := c.Query(query, field.FormID, field.ElementorID, field.Label, field.Type, required, field.Position)
	return err
}

func (c *Client) createFormNumber(number *Number) error {
	query := `
		INSERT INTO form_numbers (form_id, phone_number, label, contact_id)
		VALUES (?, ?, ?, ?)
	`
	
	_, err := c.Query(query, number.FormID, number.PhoneNumber, number.Label, number.ContactID)
	return err
}

// SearchForms searches for forms by name or description
func (c *Client) SearchForms(searchTerm string) ([]FormWithStats, error) {
	query := `
		SELECT 
			f.id,
			f.name,
			f.description,
			f.created_at,
			f.updated_at,
			COUNT(DISTINCT ff.id) as field_count,
			COUNT(DISTINCT fn.id) as number_count
		FROM forms f
		LEFT JOIN form_fields ff ON f.id = ff.form_id
		LEFT JOIN form_numbers fn ON f.id = fn.form_id
		WHERE f.name LIKE ? OR f.description LIKE ?
		GROUP BY f.id
		ORDER BY f.created_at DESC
	`

	searchPattern := "%" + searchTerm + "%"
	result, err := c.Query(query, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search forms: %w", err)
	}

	var forms []FormWithStats
	for _, row := range result.Results {
		form := FormWithStats{}
		
		if id, ok := row["id"].(string); ok {
			form.ID = id
		}
		if name, ok := row["name"].(string); ok {
			form.Name = name
		}
		if desc, ok := row["description"].(string); ok {
			form.Description = desc
		}
		if createdAt, ok := row["created_at"].(string); ok {
			form.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		if updatedAt, ok := row["updated_at"].(string); ok {
			form.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		}
		if count, ok := row["field_count"].(float64); ok {
			form.FieldCount = int(count)
		}
		if count, ok := row["number_count"].(float64); ok {
			form.NumberCount = int(count)
		}

		forms = append(forms, form)
	}

	return forms, nil
}

// ExportForm exports a form configuration as JSON
func (c *Client) ExportForm(id string) ([]byte, error) {
	form, err := c.GetForm(id)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(form, "", "  ")
}

// ImportForm imports a form configuration from JSON
func (c *Client) ImportForm(data []byte) error {
	var form Form
	if err := json.Unmarshal(data, &form); err != nil {
		return fmt.Errorf("failed to unmarshal form: %w", err)
	}

	// Check if form with same ID exists
	existing, _ := c.GetForm(form.ID)
	if existing != nil {
		return fmt.Errorf("form with ID %s already exists", form.ID)
	}

	return c.CreateForm(&form)
}