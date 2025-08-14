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

type EditView struct {
	config          *config.Config
	styles          *styles.Styles
	db              *database.Client
	form            *huh.Form
	contactData     ContactData
	originalContact *database.Contact
	err             error
	width           int
	height          int
	done            bool
	loading         bool
}

func NewEditView(cfg *config.Config, s *styles.Styles, contactID int) *EditView {
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

	// Load the contact data
	v.loadContactData(contactID)
	
	return v
}

func (v *EditView) loadContactData(contactID int) {
	// Load the contact
	contact, err := v.db.GetContactByID(contactID)
	if err != nil {
		v.err = fmt.Errorf("failed to load contact: %w", err)
		v.loading = false
		return
	}
	v.originalContact = contact

	// Convert to ContactData
	v.contactData.Name = contact.Name
	v.contactData.PhoneNumber = contact.PhoneNumber
	v.contactData.Company = contact.Company
	v.contactData.Role = contact.Role
	v.contactData.Notes = contact.Notes

	v.buildForm()
	v.loading = false
}

func (v *EditView) buildForm() {
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
				Title("Update Contact?").
				Description("Apply changes to this contact?").
				Value(&v.contactData.Confirmed),
		),
	)

	v.form.WithTheme(huh.ThemeCharm())
	v.form.WithWidth(60)
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

	case ContactUpdatedMsg:
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
	if v.form.State == huh.StateCompleted && v.contactData.Confirmed {
		// Update the contact
		return v, v.updateContact
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

	title := v.styles.Title.Render(fmt.Sprintf("✏️ Edit Contact: %s", v.contactData.Name))
	
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
		v.styles.Spinner.Render("Loading contact data..."),
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
	success := v.styles.Success.Render("✅ Contact updated successfully!")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		success,
	)
}

func (v *EditView) updateContact() tea.Msg {
	if v.db == nil {
		v.err = fmt.Errorf("database client not initialized")
		return ContactUpdatedMsg{Error: v.err}
	}

	// Update the contact object
	contact := v.originalContact
	contact.Name = v.contactData.Name
	contact.PhoneNumber = v.contactData.PhoneNumber
	contact.Company = v.contactData.Company
	contact.Role = v.contactData.Role
	contact.Notes = v.contactData.Notes

	// Update in database
	if err := v.db.UpdateContact(contact); err != nil {
		v.err = err
		return ContactUpdatedMsg{Error: err}
	}

	return ContactUpdatedMsg{ContactID: contact.ID}
}

// Message types
type ContactUpdatedMsg struct {
	ContactID int
	Error     error
}