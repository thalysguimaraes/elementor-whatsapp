package forms

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
	config  *config.Config
	styles  *styles.Styles
	table   table.Model
	spinner spinner.Model
	db      *database.Client
	forms   []database.FormWithStats
	loading bool
	err     error
	width   int
	height  int
}

func NewListView(cfg *config.Config, s *styles.Styles) *ListView {
	// Create table with empty data initially
	columns := []table.Column{
		{Title: "ID", Width: 20},
		{Title: "Name", Width: 30},
		{Title: "Fields", Width: 10},
		{Title: "Recipients", Width: 12},
		{Title: "Created", Width: 20},
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
		Foreground(s.Colors.Primary).
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
	// Only start the spinner, don't load forms immediately
	return m.spinner.Tick
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
		case "n":
			// Create new form
			return m, func() tea.Msg {
				return SwitchToCreateMsg{}
			}
		case "e":
			// Edit selected form
			if len(m.forms) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.forms) {
					return m, func() tea.Msg {
						return SwitchToEditMsg{FormID: m.forms[selectedIdx].ID}
					}
				}
			}
		case "d":
			// Delete selected form
			if len(m.forms) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.forms) {
					// TODO: Show confirmation dialog
					// For now, just delete directly
					return m, m.deleteForm(m.forms[selectedIdx].ID)
				}
			}
		case "enter":
			// View form details (for now, same as edit)
			if len(m.forms) > 0 {
				selectedIdx := m.table.Cursor()
				if selectedIdx < len(m.forms) {
					return m, func() tea.Msg {
						return SwitchToEditMsg{FormID: m.forms[selectedIdx].ID}
					}
				}
			}
		case "r":
			// Refresh
			m.loading = true
			return m, m.loadForms
		}
		
	case FormsLoadedMsg:
		m.loading = false
		m.forms = msg.Forms
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
	
	title := m.styles.Title.Render("📝 Forms Management")
	
	// Stats bar
	totalFields := 0
	totalRecipients := 0
	for _, form := range m.forms {
		totalFields += form.FieldCount
		totalRecipients += form.NumberCount
	}
	stats := m.styles.Muted.Render(fmt.Sprintf("%d forms • %d total fields • %d recipients", len(m.forms), totalFields, totalRecipients))
	
	// Table
	tableView := m.table.View()
	
	// Actions hint
	actions := m.styles.Help.Render("n: New • e: Edit • d: Delete • Enter: View • r: Refresh")
	
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
		m.spinner.View()+" Loading forms...",
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

func (m *ListView) loadForms() tea.Msg {
	if m.db == nil {
		return FormsLoadedMsg{
			Error: fmt.Errorf("database client not initialized"),
		}
	}
	
	forms, err := m.db.GetAllForms()
	if err != nil {
		log.Error("Failed to load forms", "error", err)
		return FormsLoadedMsg{
			Error: err,
		}
	}
	
	return FormsLoadedMsg{
		Forms: forms,
	}
}

func (m *ListView) updateTable() {
	var rows []table.Row
	for _, form := range m.forms {
		rows = append(rows, table.Row{
			form.ID,
			form.Name,
			fmt.Sprintf("%d", form.FieldCount),
			fmt.Sprintf("%d", form.NumberCount),
			form.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	
	m.table.SetRows(rows)
}

// Message types
type FormsLoadedMsg struct {
	Forms []database.FormWithStats
	Error error
}

type ViewActivatedMsg struct{}

type SwitchToCreateMsg struct{}

type SwitchToEditMsg struct {
	FormID string
}

type FormDeletedMsg struct {
	FormID string
	Error  error
}

// StartLoading triggers the initial data load when the view becomes active
func (m *ListView) StartLoading() tea.Cmd {
	if !m.loading {
		m.loading = true
		return m.loadForms
	}
	return nil
}

// ForceReload forces a reload of forms regardless of current state
func (m *ListView) ForceReload() tea.Cmd {
	m.loading = true
	return m.loadForms
}

func (m *ListView) deleteForm(formID string) tea.Cmd {
	return func() tea.Msg {
		err := m.db.DeleteForm(formID)
		if err != nil {
			return FormDeletedMsg{FormID: formID, Error: err}
		}
		// Reload forms after deletion
		forms, err := m.db.GetAllForms()
		return FormsLoadedMsg{Forms: forms, Error: err}
	}
}