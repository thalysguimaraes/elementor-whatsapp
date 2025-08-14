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

type CreateView struct {
	config   *config.Config
	styles   *styles.Styles
	db       *database.Client
	form     *huh.Form
	formData FormData
	contacts []database.Contact
	err      error
	width    int
	height   int
	done     bool
}

type FormData struct {
	ID               string
	Name             string
	Description      string
	Fields           []FieldData
	SelectedContacts []string
	Confirmed        bool
}

type FieldData struct {
	ElementorID string
	Label       string
	Type        string
	Required    bool
}

func NewCreateView(cfg *config.Config, s *styles.Styles) *CreateView {
	// Create database client
	db, err := database.NewClient(cfg)
	if err != nil {
		log.Error("Failed to create database client", "error", err)
	}

	v := &CreateView{
		config: cfg,
		styles: s,
		db:     db,
		err:    err,
	}

	// Initialize with some default fields
	v.formData.Fields = []FieldData{
		{ElementorID: "name", Label: "Name", Type: "text", Required: true},
		{ElementorID: "email", Label: "Email", Type: "email", Required: true},
		{ElementorID: "phone", Label: "Phone", Type: "tel", Required: false},
		{ElementorID: "message", Label: "Message", Type: "textarea", Required: false},
	}

	v.buildForm()
	return v
}

func (v *CreateView) buildForm() {
	// Load contacts for selection
	if v.db != nil {
		contacts, err := v.db.GetAllContacts()
		if err != nil {
			log.Error("Failed to load contacts", "error", err)
		} else {
			v.contacts = contacts
		}
	}

	// Build contact options
	var contactOptions []huh.Option[string]
	for _, contact := range v.contacts {
		label := fmt.Sprintf("%s (%s)", contact.Name, contact.PhoneNumber)
		if contact.Company != "" {
			label = fmt.Sprintf("%s - %s (%s)", contact.Name, contact.Company, contact.PhoneNumber)
		}
		contactOptions = append(contactOptions, huh.NewOption(label, fmt.Sprintf("%d", contact.ID)))
	}

	// Create the form
	v.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Form ID").
				Description("Unique identifier for the form (leave blank to auto-generate)").
				Value(&v.formData.ID).
				Placeholder("contact-form").
				Validate(func(s string) error {
					if s != "" && strings.Contains(s, " ") {
						return fmt.Errorf("ID cannot contain spaces")
					}
					return nil
				}),

			huh.NewInput().
				Title("Form Name").
				Description("Display name for the form").
				Value(&v.formData.Name).
				Placeholder("Contact Form").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),

			huh.NewText().
				Title("Description").
				Description("Optional description of the form").
				Value(&v.formData.Description).
				Placeholder("Form for website contact submissions"),
		),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Recipients").
				Description("Choose contacts who will receive notifications").
				Options(contactOptions...).
				Value(&v.formData.SelectedContacts),

			huh.NewConfirm().
				Title("Create Form?").
				Description("Are you ready to create this form?").
				Value(&v.formData.Confirmed).
				Affirmative("Yes, create it!").
				Negative("No, go back"),
		),
	).WithTheme(huh.ThemeCharm())
}

func (v *CreateView) Init() tea.Cmd {
	return v.form.Init()
}

func (v *CreateView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case FormCreatedMsg:
		if msg.Error != nil {
			v.err = msg.Error
			return v, nil
		}
		v.done = true
		// Go back to list after successful creation
		return v, func() tea.Msg {
			return GoBackToListMsg{}
		}
	}

	// Check if form is complete
	if v.form.State == huh.StateCompleted && v.formData.Confirmed {
		// Create the form
		return v, v.createForm
	}

	// Update the form
	form, cmd := v.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		v.form = f
	}

	return v, cmd
}

func (v *CreateView) View() string {
	if v.err != nil {
		return v.renderError()
	}

	if v.done {
		return v.renderSuccess()
	}

	title := v.styles.Title.Render("📝 Create New Form")
	
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

func (v *CreateView) renderError() string {
	errorView := v.styles.Error.Render(fmt.Sprintf("Error: %v", v.err))
	help := v.styles.Help.Render("Press Esc to go back")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		errorView,
		help,
	)
}

func (v *CreateView) renderSuccess() string {
	successView := v.styles.Success.Render("✓ Form created successfully!")
	webhookURL := fmt.Sprintf("%s/webhook/%s", v.config.Cloudflare.WorkerURL, v.formData.ID)
	urlView := v.styles.Info.Render(fmt.Sprintf("Webhook URL: %s", webhookURL))
	help := v.styles.Help.Render("Press Enter to continue")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		successView,
		urlView,
		help,
	)
}

func (v *CreateView) createForm() tea.Msg {
	if v.db == nil {
		v.err = fmt.Errorf("database client not initialized")
		return FormCreatedMsg{Error: v.err}
	}

	// Generate ID if not provided
	if v.formData.ID == "" {
		v.formData.ID = generateFormID(v.formData.Name)
	}

	// Build the form object
	form := &database.Form{
		ID:          v.formData.ID,
		Name:        v.formData.Name,
		Description: v.formData.Description,
	}

	// Add fields
	for i, field := range v.formData.Fields {
		form.Fields = append(form.Fields, database.Field{
			FormID:      v.formData.ID,
			ElementorID: field.ElementorID,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
			Position:    i,
		})
	}

	// Add selected contacts as numbers
	for _, contactID := range v.formData.SelectedContacts {
		// Find the contact
		for _, contact := range v.contacts {
			if fmt.Sprintf("%d", contact.ID) == contactID {
				cID := contact.ID
				form.Numbers = append(form.Numbers, database.Number{
					FormID:      v.formData.ID,
					PhoneNumber: contact.PhoneNumber,
					Label:       contact.Name,
					ContactID:   &cID,
				})
				break
			}
		}
	}

	// Create the form in the database
	if err := v.db.CreateForm(form); err != nil {
		v.err = err
		return FormCreatedMsg{Error: err}
	}

	return FormCreatedMsg{FormID: form.ID}
}

func generateFormID(name string) string {
	// Simple ID generation from name
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	// Remove non-alphanumeric characters
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Message types
type FormCreatedMsg struct {
	FormID string
	Error  error
}

type GoBackToListMsg struct{}