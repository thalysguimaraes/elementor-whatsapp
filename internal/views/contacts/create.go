package contacts

import (
	"fmt"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/database"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
)

type CreateView struct {
	config      *config.Config
	styles      *styles.Styles
	db          *database.Client
	form        *huh.Form
	contactData ContactData
	err         error
	width       int
	height      int
	done        bool
}

type ContactData struct {
	Name        string
	PhoneNumber string
	Company     string
	Role        string
	Notes       string
	Confirmed   bool
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

	v.buildForm()
	return v
}

func (v *CreateView) buildForm() {
	// Create the form
	v.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Contact Name").
				Description("Full name of the contact").
				Value(&v.contactData.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("WhatsApp Number").
				Description("Phone number with country code (e.g., 5511999999999)").
				Value(&v.contactData.PhoneNumber).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("phone number is required")
					}
					// Basic validation for phone number
					matched, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, s)
					if !matched {
						return fmt.Errorf("invalid phone number format")
					}
					return nil
				}),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Company").
				Description("Company or organization (optional)").
				Value(&v.contactData.Company),

			huh.NewInput().
				Title("Role/Position").
				Description("Job title or position (optional)").
				Value(&v.contactData.Role),
		),

		huh.NewGroup(
			huh.NewText().
				Title("Notes").
				Description("Additional notes about this contact (optional)").
				Value(&v.contactData.Notes).
				Lines(4),

			huh.NewConfirm().
				Title("Create Contact?").
				Description("Add this contact to the system?").
				Value(&v.contactData.Confirmed),
		),
	)

	v.form.WithTheme(huh.ThemeCharm())
	v.form.WithWidth(60)
}

func (v *CreateView) Init() tea.Cmd {
	if v.form != nil {
		return v.form.Init()
	}
	return nil
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

	case ContactCreatedMsg:
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
	if v.form.State == huh.StateCompleted && v.contactData.Confirmed {
		// Create the contact
		return v, v.createContact
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

	title := v.styles.Title.Render("📞 Add New Contact")
	
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
		"",
		help,
	)
}

func (v *CreateView) renderSuccess() string {
	success := v.styles.Success.Render("✅ Contact created successfully!")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		success,
	)
}

func (v *CreateView) createContact() tea.Msg {
	if v.db == nil {
		v.err = fmt.Errorf("database client not initialized")
		return ContactCreatedMsg{Error: v.err}
	}

	// Create the contact
	contact := &database.Contact{
		Name:        v.contactData.Name,
		PhoneNumber: v.contactData.PhoneNumber,
		Company:     v.contactData.Company,
		Role:        v.contactData.Role,
		Notes:       v.contactData.Notes,
	}

	// Save to database
	id, err := v.db.CreateContact(contact)
	if err != nil {
		v.err = err
		return ContactCreatedMsg{Error: err}
	}
	contact.ID = id

	return ContactCreatedMsg{ContactID: contact.ID}
}

// Message types
type ContactCreatedMsg struct {
	ContactID int
	Error     error
}

type GoBackToListMsg struct{}