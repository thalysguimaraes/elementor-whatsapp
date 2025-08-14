package forms

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/database"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
)

type EditView struct {
	config       *config.Config
	styles       *styles.Styles
	db           *database.Client
	form         *huh.Form
	formData     FormData
	originalForm *database.Form
	contacts     []database.Contact
	err          error
	width        int
	height       int
	done         bool
	loading      bool
}

func NewEditView(cfg *config.Config, s *styles.Styles, formID string) *EditView {
	// Create database client
	db, err := database.NewClient(cfg)
	if err != nil {
		log.Error("Failed to create database client", "error", err)
		return &EditView{
			config: cfg,
			styles: s,
			err:    err,
		}
	}

	v := &EditView{
		config:  cfg,
		styles:  s,
		db:      db,
		loading: true,
	}

	// Load the form data
	v.loadFormData(formID)
	
	return v
}

func (v *EditView) loadFormData(formID string) {
	// Load the form
	form, err := v.db.GetFormByID(formID)
	if err != nil {
		v.err = fmt.Errorf("failed to load form: %w", err)
		v.loading = false
		return
	}
	v.originalForm = form

	// Convert to FormData
	v.formData.ID = form.ID
	v.formData.Name = form.Name
	v.formData.Description = form.Description
	
	// Convert fields
	v.formData.Fields = make([]FieldData, len(form.Fields))
	for i, field := range form.Fields {
		v.formData.Fields[i] = FieldData{
			ElementorID: field.ElementorID,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
		}
	}

	// Get selected contact IDs
	v.formData.SelectedContacts = make([]string, len(form.Numbers))
	for i, number := range form.Numbers {
		v.formData.SelectedContacts[i] = fmt.Sprintf("%d", number.ContactID)
	}

	// Load all contacts for selection
	contacts, err := v.db.GetAllContacts()
	if err != nil {
		log.Error("Failed to load contacts", "error", err)
	} else {
		v.contacts = contacts
	}

	v.buildForm()
	v.loading = false
}

func (v *EditView) buildForm() {
	// Build contact options
	var contactOptions []huh.Option[string]
	for _, contact := range v.contacts {
		label := fmt.Sprintf("%s (%s)", contact.Name, contact.PhoneNumber)
		if contact.Company != "" {
			label = fmt.Sprintf("%s - %s (%s)", contact.Name, contact.Company, contact.PhoneNumber)
		}
		contactOptions = append(contactOptions, huh.NewOption(label, fmt.Sprintf("%d", contact.ID)))
	}

	// Build field strings for editing
	fieldStrings := make([]string, len(v.formData.Fields))
	for i, field := range v.formData.Fields {
		required := ""
		if field.Required {
			required = " *"
		}
		fieldStrings[i] = fmt.Sprintf("%s|%s|%s%s", field.ElementorID, field.Label, field.Type, required)
	}
	fieldsText := strings.Join(fieldStrings, "\n")

	// Create the form
	v.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Form Name").
				Description("Display name for this form").
				Value(&v.formData.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("form name is required")
					}
					return nil
				}),

			huh.NewText().
				Title("Description").
				Description("Optional description of the form").
				Value(&v.formData.Description).
				Lines(3),
		),

		huh.NewGroup(
			huh.NewText().
				Title("Form Fields").
				Description("One field per line: elementor_id|Label|type|required\nTypes: text, email, tel, textarea, select\nAdd * for required fields").
				Value(&fieldsText).
				Lines(10).
				Validate(func(s string) error {
					// Parse and validate fields
					lines := strings.Split(strings.TrimSpace(s), "\n")
					if len(lines) == 0 {
						return fmt.Errorf("at least one field is required")
					}
					
					v.formData.Fields = make([]FieldData, 0)
					for _, line := range lines {
						if strings.TrimSpace(line) == "" {
							continue
						}
						
						parts := strings.Split(line, "|")
						if len(parts) < 3 {
							return fmt.Errorf("invalid field format: %s", line)
						}
						
						field := FieldData{
							ElementorID: strings.TrimSpace(parts[0]),
							Label:       strings.TrimSpace(parts[1]),
							Type:        strings.TrimSpace(parts[2]),
							Required:    false,
						}
						
						if len(parts) > 3 || strings.HasSuffix(field.Type, "*") {
							field.Required = true
							field.Type = strings.TrimSuffix(field.Type, "*")
							field.Type = strings.TrimSpace(field.Type)
						}
						
						v.formData.Fields = append(v.formData.Fields, field)
					}
					
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Recipients").
				Description("Choose contacts who will receive notifications").
				Options(contactOptions...).
				Value(&v.formData.SelectedContacts),

			huh.NewConfirm().
				Title("Update Form?").
				Description("Apply changes to this form?").
				Value(&v.formData.Confirmed),
		),
	)

	v.form.WithTheme(huh.ThemeCharm())
	v.form.WithWidth(80)
}

func (v *EditView) Init() tea.Cmd {
	if v.form != nil {
		return v.form.Init()
	}
	return nil
}

func (v *EditView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.loading {
		return v, nil
	}

	if v.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" {
				return v, func() tea.Msg {
					return GoBackToListMsg{}
				}
			}
		}
		return v, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Go back to list view
			return v, func() tea.Msg {
				return GoBackToListMsg{}
			}
		}

	case FormUpdatedMsg:
		if msg.Error != nil {
			v.err = msg.Error
			return v, nil
		}
		v.done = true
		// Go back to list after successful update
		return v, func() tea.Msg {
			return GoBackToListMsg{}
		}
	}

	// Check if form is complete
	if v.form.State == huh.StateCompleted && v.formData.Confirmed {
		// Update the form
		return v, v.updateForm
	}

	// Update the form
	form, cmd := v.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		v.form = f
	}

	return v, cmd
}

func (v *EditView) View() string {
	if v.loading {
		return v.renderLoading()
	}

	if v.err != nil {
		return v.renderError()
	}

	if v.done {
		return v.renderSuccess()
	}

	title := v.styles.Title.Render(fmt.Sprintf("✏️ Edit Form: %s", v.formData.Name))
	
	// Form view
	formView := v.form.View()

	// Help text
	help := v.styles.Help.Render("Tab: Next Field • Shift+Tab: Previous • Enter: Submit • Esc: Cancel")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		"",
		formView,
		"",
		help,
	)
}

func (v *EditView) renderLoading() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		v.styles.Spinner.Render("Loading form data..."),
	)
}

func (v *EditView) renderError() string {
	errorView := v.styles.Error.Render(fmt.Sprintf("Error: %v", v.err))
	help := v.styles.Help.Render("Press Esc to go back")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		errorView,
		"",
		help,
	)
}

func (v *EditView) renderSuccess() string {
	success := v.styles.Success.Render("✅ Form updated successfully!")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		success,
	)
}

func (v *EditView) updateForm() tea.Msg {
	if v.db == nil {
		v.err = fmt.Errorf("database client not initialized")
		return FormUpdatedMsg{Error: v.err}
	}

	// Update the form object
	form := v.originalForm
	form.Name = v.formData.Name
	form.Description = v.formData.Description
	
	// Update fields
	form.Fields = make([]database.Field, len(v.formData.Fields))
	for i, field := range v.formData.Fields {
		form.Fields[i] = database.Field{
			FormID:      form.ID,
			ElementorID: field.ElementorID,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
			Position:    i,
		}
	}

	// Update numbers/contacts
	form.Numbers = make([]database.Number, len(v.formData.SelectedContacts))
	for i, contactIDStr := range v.formData.SelectedContacts {
		var contactID int
		fmt.Sscanf(contactIDStr, "%d", &contactID)
		
		// Get contact details
		contact, err := v.db.GetContactByID(contactID)
		if err != nil {
			log.Error("Failed to get contact", "id", contactID, "error", err)
			continue
		}
		
		form.Numbers[i] = database.Number{
			FormID:      form.ID,
			ContactID:   &contactID,
			PhoneNumber: contact.PhoneNumber,
			Label:       contact.Name,
		}
	}

	// Update in database
	if err := v.db.UpdateForm(form); err != nil {
		v.err = err
		return FormUpdatedMsg{Error: err}
	}

	return FormUpdatedMsg{FormID: form.ID}
}

// Message types
type FormUpdatedMsg struct {
	FormID string
	Error  error
}