package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/views/dashboard"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/views/forms"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/views/contacts"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/views/webhook"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/views/settings"
)

type View int

const (
	ViewDashboard View = iota
	ViewForms
	ViewFormCreate
	ViewFormEdit
	ViewContacts
	ViewContactCreate
	ViewContactEdit
	ViewWebhook
	ViewSettings
)

type Model struct {
	config      *config.Config
	currentView View
	views       map[View]tea.Model
	width       int
	height      int
	styles      *styles.Styles
	breadcrumbs []string
	err         error
}

func NewModel(cfg *config.Config) *Model {
	s := styles.NewStyles(cfg.UI.Theme)
	
	m := &Model{
		config:      cfg,
		currentView: ViewDashboard,
		styles:      s,
		breadcrumbs: []string{"Dashboard"},
		views:       make(map[View]tea.Model),
	}

	// Initialize views
	m.views[ViewDashboard] = dashboard.New(cfg, s)
	m.views[ViewForms] = forms.NewListView(cfg, s)
	m.views[ViewFormCreate] = forms.NewCreateView(cfg, s)
	// Note: FormEditView is created dynamically with form ID
	m.views[ViewContacts] = contacts.NewListView(cfg, s)
	m.views[ViewContactCreate] = contacts.NewCreateView(cfg, s)
	// Note: ContactEditView is created dynamically with contact ID
	m.views[ViewWebhook] = webhook.New(cfg, s)
	m.views[ViewSettings] = settings.New(cfg, s)

	return m
}

func (m *Model) Init() tea.Cmd {
	// Initialize all views
	var cmds []tea.Cmd
	for _, view := range m.views {
		if v, ok := view.(tea.Model); ok {
			cmds = append(cmds, v.Init())
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update all views with new size
		for viewType, view := range m.views {
			if v, ok := view.(tea.Model); ok {
				updated, cmd := v.Update(msg)
				m.views[viewType] = updated
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.KeyMsg:
		// Handle ctrl+c globally
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Dashboard navigation only works from dashboard
		if m.currentView == ViewDashboard {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "1":
				cmd := m.switchView(ViewDashboard, "Dashboard")
				cmds = append(cmds, cmd)
			case "2":
				cmd := m.switchView(ViewForms, "Forms")
				cmds = append(cmds, cmd)
			case "3":
				cmd := m.switchView(ViewContacts, "Contacts")
				cmds = append(cmds, cmd)
			case "4":
				cmd := m.switchView(ViewWebhook, "Test Webhook")
				cmds = append(cmds, cmd)
			case "5":
				cmd := m.switchView(ViewSettings, "Settings")
				cmds = append(cmds, cmd)
			default:
				// Pass other keys to dashboard
				if currentView, ok := m.views[m.currentView].(tea.Model); ok {
					updated, cmd := currentView.Update(msg)
					m.views[m.currentView] = updated
					cmds = append(cmds, cmd)
				}
			}
		} else {
			// For non-dashboard views, let them handle keys first
			// Only handle ESC as a global "go back" if the view doesn't handle it
			if msg.String() == "esc" {
				// Try to pass to current view first
				if currentView, ok := m.views[m.currentView].(tea.Model); ok {
					updated, cmd := currentView.Update(msg)
					m.views[m.currentView] = updated
					
					// Check if view wants to go back (e.g., by checking a flag)
					// For now, always go back to dashboard on ESC from other views
					if m.currentView != ViewDashboard {
						m.currentView = ViewDashboard
						m.breadcrumbs = []string{"Dashboard"}
					}
					return m, cmd
				}
			} else {
				// Pass all other keys to the current view
				if currentView, ok := m.views[m.currentView].(tea.Model); ok {
					updated, cmd := currentView.Update(msg)
					m.views[m.currentView] = updated
					cmds = append(cmds, cmd)
				}
			}
		}

	case SwitchViewMsg:
		cmd := m.switchView(msg.View, msg.Title)
		cmds = append(cmds, cmd)
		if msg.Data != nil {
			// Pass data to the new view if it supports it
			if currentView, ok := m.views[m.currentView].(tea.Model); ok {
				updated, cmd := currentView.Update(msg.Data)
				m.views[m.currentView] = updated
				cmds = append(cmds, cmd)
			}
		}

	case forms.SwitchToCreateMsg:
		// Switch to form create view
		cmd := m.switchView(ViewFormCreate, "Create Form")
		cmds = append(cmds, cmd)

	case forms.SwitchToEditMsg:
		// Create and switch to form edit view
		m.views[ViewFormEdit] = forms.NewEditView(m.config, m.styles, msg.FormID)
		cmd := m.switchView(ViewFormEdit, "Edit Form")
		// Initialize the edit view
		if editView, ok := m.views[ViewFormEdit].(tea.Model); ok {
			initCmd := editView.Init()
			cmds = append(cmds, cmd, initCmd)
		} else {
			cmds = append(cmds, cmd)
		}

	case forms.GoBackToListMsg:
		// Go back to forms list
		cmd := m.switchView(ViewForms, "Forms")
		// Force reload of forms to show new/edited forms
		if formsView, ok := m.views[ViewForms].(*forms.ListView); ok {
			reloadCmd := formsView.ForceReload()
			cmds = append(cmds, cmd, reloadCmd)
		} else {
			cmds = append(cmds, cmd)
		}

	case contacts.SwitchToCreateMsg:
		// Switch to contact create view
		cmd := m.switchView(ViewContactCreate, "Create Contact")
		cmds = append(cmds, cmd)

	case contacts.SwitchToEditMsg:
		// Create and switch to contact edit view
		m.views[ViewContactEdit] = contacts.NewEditView(m.config, m.styles, msg.ContactID)
		cmd := m.switchView(ViewContactEdit, "Edit Contact")
		// Initialize the edit view
		if editView, ok := m.views[ViewContactEdit].(tea.Model); ok {
			initCmd := editView.Init()
			cmds = append(cmds, cmd, initCmd)
		} else {
			cmds = append(cmds, cmd)
		}

	case contacts.GoBackToListMsg:
		// Go back to contacts list
		cmd := m.switchView(ViewContacts, "Contacts")
		// Force reload of contacts to show new/edited contacts
		if contactsView, ok := m.views[ViewContacts].(*contacts.ListView); ok {
			reloadCmd := contactsView.ForceReload()
			cmds = append(cmds, cmd, reloadCmd)
		} else {
			cmds = append(cmds, cmd)
		}

	case error:
		m.err = msg
		return m, nil

	default:
		// Pass to current view
		if currentView, ok := m.views[m.currentView].(tea.Model); ok {
			updated, cmd := currentView.Update(msg)
			m.views[m.currentView] = updated
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	var content string
	if currentView, ok := m.views[m.currentView].(tea.Model); ok {
		content = currentView.View()
	}

	// Build the full view with header and footer
	header := m.renderHeader()
	footer := m.renderFooter()

	// Calculate available height for content
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := m.height - headerHeight - footerHeight - 2

	// Apply content area styling
	contentStyle := m.styles.Content.
		Width(m.width).
		Height(contentHeight)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		contentStyle.Render(content),
		footer,
	)
}

func (m *Model) renderHeader() string {
	// Title
	title := m.styles.Title.Render("📋 Elementor WhatsApp Manager")
	
	// Breadcrumbs
	breadcrumbStr := ""
	for i, crumb := range m.breadcrumbs {
		if i > 0 {
			breadcrumbStr += " › "
		}
		breadcrumbStr += crumb
	}
	breadcrumbs := m.styles.Breadcrumb.Render(breadcrumbStr)

	// Combine title and breadcrumbs
	left := lipgloss.JoinVertical(lipgloss.Top, title, breadcrumbs)
	
	// Status info (right side)
	status := fmt.Sprintf("Profile: %s", "default")
	right := m.styles.StatusBar.Render(status)

	// Join left and right with proper spacing
	width := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	spacer := lipgloss.NewStyle().Width(width).Render("")
	
	header := lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
	
	return m.styles.Header.Width(m.width).Render(header)
}

func (m *Model) renderFooter() string {
	var help string
	
	switch m.currentView {
	case ViewDashboard:
		help = "1-5: Navigate • ?: Help • q: Quit"
	case ViewForms:
		help = "↑↓/jk: Navigate • n: New • e: Edit • d: Delete • Enter: Select • Esc: Back"
	case ViewFormCreate, ViewFormEdit:
		help = "Tab: Next Field • Enter: Submit • Esc: Cancel"
	case ViewContacts:
		help = "↑↓/jk: Navigate • a: Add • e: Edit • d: Delete • Enter: View • Esc: Back"
	case ViewContactCreate, ViewContactEdit:
		help = "Tab: Next Field • Enter: Submit • Esc: Cancel"
	case ViewWebhook:
		help = "Tab: Next Field • Enter: Send • Esc: Back"
	case ViewSettings:
		help = "↑↓: Navigate • Enter: Edit • s: Save • Esc: Back"
	default:
		help = "?: Help • Esc: Back • q: Quit"
	}

	return m.styles.Footer.Width(m.width).Render(help)
}

func (m *Model) renderError() string {
	errorView := m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err))
	help := m.styles.Help.Render("Press any key to continue...")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		errorView,
		help,
	)
}

func (m *Model) switchView(view View, title string) tea.Cmd {
	m.currentView = view
	if len(m.breadcrumbs) > 1 {
		m.breadcrumbs = m.breadcrumbs[:1]
	}
	if view != ViewDashboard {
		m.breadcrumbs = append(m.breadcrumbs, title)
	}
	
	// Send activation message to the view when switching
	switch view {
	case ViewForms:
		if formsView, ok := m.views[ViewForms].(*forms.ListView); ok {
			return formsView.StartLoading()
		}
	case ViewContacts:
		if contactsView, ok := m.views[ViewContacts].(*contacts.ListView); ok {
			return contactsView.StartLoading()
		}
	}
	
	return nil
}

// Message types
type SwitchViewMsg struct {
	View  View
	Title string
	Data  interface{}
}

// Run starts the TUI application
func Run(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m := NewModel(cfg)
	
	p := tea.NewProgram(m, tea.WithAltScreen())
	if cfg.UI.Mouse {
		p = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	
	return nil
}

// Direct view runners for CLI commands
func RunFormsListView(cfg *config.Config) error {
	// TODO: Implement direct forms list view
	return Run(cfg)
}

func RunFormCreateView(cfg *config.Config) error {
	// TODO: Implement direct form create view
	return Run(cfg)
}

func RunContactsListView(cfg *config.Config) error {
	// TODO: Implement direct contacts list view
	return Run(cfg)
}

func RunContactCreateView(cfg *config.Config) error {
	// TODO: Implement direct contact create view
	return Run(cfg)
}

func RunWebhookTestView(cfg *config.Config, formID string) error {
	// TODO: Implement direct webhook test view
	return Run(cfg)
}

func RunConfigEditView(cfg *config.Config) error {
	// TODO: Implement direct config edit view
	return Run(cfg)
}