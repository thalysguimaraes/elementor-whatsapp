package contacts

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/database"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
)

type ListView struct {
	config   *config.Config
	styles   *styles.Styles
	table    table.Model
	spinner  spinner.Model
	db       *database.Client
	contacts []database.ContactWithStats
	loading  bool
	err      error
	width    int
	height   int
}

func NewListView(cfg *config.Config, s *styles.Styles) *ListView {
	// Create table with empty data initially
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Phone", Width: 20},
		{Title: "Company", Width: 25},
		{Title: "Role", Width: 20},
		{Title: "Forms", Width: 10},
	}
	
	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(s.Colors.Border).
		BorderBottom(true).
		Bold(false)
	tableStyle.Selected = tableStyle.Selected.
		Foreground(s.Colors.Secondary).
		Background(s.Colors.BgSecondary).
		Bold(false)
	t.SetStyles(tableStyle)
	
	// Create spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = s.Spinner
	
	// Create database client
	db, err := database.NewClient(cfg)
	if err != nil {
		log.Error("Failed to create database client", "error", err)
	}
	
	return &ListView{
		config:  cfg,
		styles:  s,
		table:   t,
		spinner: sp,
		db:      db,
		loading: false,  // Don't start loading immediately
		err:     err,
	}
}

func (m *ListView) Init() tea.Cmd {
	// Only start the spinner, don't load contacts immediately
	return m.spinner.Tick
}

// StartLoading triggers the initial data load when the view becomes active
func (m *ListView) StartLoading() tea.Cmd {
	if !m.loading && len(m.contacts) == 0 {
		m.loading = true
		return m.loadContacts
	}
	return nil
}

func (m *ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case ViewActivatedMsg:
		// Start loading when view becomes active
		if cmd := m.StartLoading(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.height - 10)
		
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		
		switch msg.String() {
		case "a":
			// Add new contact
			return m, func() tea.Msg {
				return SwitchToCreateMsg{}
			}
		case "e":
			// Edit selected contact
			if len(m.contacts) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.contacts) {
					return m, func() tea.Msg {
						return SwitchToEditMsg{ContactID: m.contacts[selectedIdx].ID}
					}
				}
			}
		case "d":
			// Delete selected contact
			if len(m.contacts) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.contacts) {
					// TODO: Show confirmation dialog
					// For now, just delete directly
					return m, m.deleteContact(m.contacts[selectedIdx].ID)
				}
			}
		case "enter":
			// View contact details (for now, same as edit)
			if len(m.contacts) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.contacts) {
					return m, func() tea.Msg {
						return SwitchToEditMsg{ContactID: m.contacts[selectedIdx].ID}
					}
				}
			}
		case "r":
			// Refresh
			m.loading = true
			return m, m.loadContacts
		}
		
	case ContactsLoadedMsg:
		m.loading = false
		m.contacts = msg.Contacts
		m.err = msg.Error
		m.updateTable()
		
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	
	return m, tea.Batch(cmds...)
}

func (m *ListView) View() string {
	if m.err != nil {
		return m.renderError()
	}
	
	if m.loading {
		return m.renderLoading()
	}
	
	title := m.styles.Title.Render("📞 Contacts Management")
	
	// Stats bar
	totalForms := 0
	for _, contact := range m.contacts {
		totalForms += contact.FormCount
	}
	stats := m.styles.Muted.Render(fmt.Sprintf("%d contacts • %d form associations", len(m.contacts), totalForms))
	
	// Table
	tableView := m.table.View()
	
	// Actions hint
	actions := m.styles.Help.Render("a: Add • e: Edit • d: Delete • Enter: View • r: Refresh")
	
	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		stats,
		"",
		tableView,
		"",
		actions,
	)
}

func (m *ListView) renderLoading() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.spinner.View()+" Loading contacts...",
	)
}

func (m *ListView) renderError() string {
	errorView := m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err))
	help := m.styles.Help.Render("Press 'r' to retry")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		errorView,
		help,
	)
}

func (m *ListView) loadContacts() tea.Msg {
	if m.db == nil {
		return ContactsLoadedMsg{
			Error: fmt.Errorf("database client not initialized"),
		}
	}
	
	contacts, err := m.db.GetContactsWithStats()
	if err != nil {
		log.Error("Failed to load contacts", "error", err)
		return ContactsLoadedMsg{
			Error: err,
		}
	}
	
	return ContactsLoadedMsg{
		Contacts: contacts,
	}
}

func (m *ListView) updateTable() {
	var rows []table.Row
	for _, contact := range m.contacts {
		company := contact.Company
		if company == "" {
			company = "-"
		}
		role := contact.Role
		if role == "" {
			role = "-"
		}
		
		rows = append(rows, table.Row{
			contact.Name,
			contact.PhoneNumber,
			company,
			role,
			fmt.Sprintf("%d", contact.FormCount),
		})
	}
	
	m.table.SetRows(rows)
}

// Message types
type ContactsLoadedMsg struct {
	Contacts []database.ContactWithStats
	Error    error
}

type ViewActivatedMsg struct{}

type SwitchToCreateMsg struct{}

type SwitchToEditMsg struct {
	ContactID int
}

type ContactDeletedMsg struct {
	ContactID int
	Error     error
}

// ForceReload forces a reload of contacts regardless of current state
func (m *ListView) ForceReload() tea.Cmd {
	m.loading = true
	return m.loadContacts
}

func (m *ListView) deleteContact(contactID int) tea.Cmd {
	return func() tea.Msg {
		err := m.db.DeleteContact(contactID)
		if err != nil {
			return ContactDeletedMsg{ContactID: contactID, Error: err}
		}
		// Reload contacts after deletion
		contacts, err := m.db.GetContactsWithStats()
		return ContactsLoadedMsg{Contacts: contacts, Error: err}
	}
}