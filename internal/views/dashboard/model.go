package dashboard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/database"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
)

type Model struct {
	config   *config.Config
	styles   *styles.Styles
	spinner  spinner.Model
	loading  bool
	stats    *database.Stats
	db       *database.Client
	menuItems []MenuItem
	selected int
	width    int
	height   int
	err      error
}

type MenuItem struct {
	Title       string
	Description string
	Icon        string
	Key         string
	ViewID      int
}

func New(cfg *config.Config, s *styles.Styles) *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = s.Spinner
	
	menuItems := []MenuItem{
		{
			Title:       "Forms",
			Description: "Manage webhook forms and configurations",
			Icon:        "📝",
			Key:         "2",
			ViewID:      1, // ViewForms
		},
		{
			Title:       "Contacts",
			Description: "Organize contacts and recipients",
			Icon:        "📞",
			Key:         "3",
			ViewID:      2, // ViewContacts
		},
		{
			Title:       "Test Webhook",
			Description: "Test webhook endpoints with sample data",
			Icon:        "🧪",
			Key:         "4",
			ViewID:      3, // ViewWebhook
		},
		{
			Title:       "Settings",
			Description: "Configure application settings",
			Icon:        "⚙️",
			Key:         "5",
			ViewID:      4, // ViewSettings
		},
	}
	
	// Create database client
	db, err := database.NewClient(cfg)
	if err != nil {
		log.Error("Failed to create database client", "error", err)
	}
	
	return &Model{
		config:    cfg,
		styles:    s,
		spinner:   sp,
		loading:   true,
		menuItems: menuItems,
		selected:  0,
		db:        db,
		err:       err,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadStats,
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.menuItems)-1 {
				m.selected++
			}
		case "enter":
			// Send switch view message
			item := m.menuItems[m.selected]
			return m, m.switchView(item.ViewID, item.Title)
		case "2", "3", "4", "5":
			// Direct navigation
			for _, item := range m.menuItems {
				if item.Key == msg.String() {
					return m, m.switchView(item.ViewID, item.Title)
				}
			}
		}
		
	case StatsLoadedMsg:
		m.loading = false
		m.stats = msg.Stats
		m.err = msg.Error
		
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.err != nil {
		return m.renderError()
	}
	
	if m.loading {
		return m.renderLoading()
	}
	
	// Build dashboard layout
	statsView := m.renderStats()
	menuView := m.renderMenu()
	
	// Combine views
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		statsView,
		"\n",
		menuView,
	)
	
	return content
}

func (m *Model) renderLoading() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.spinner.View()+" Loading dashboard...",
	)
}

func (m *Model) renderStats() string {
	if m.stats == nil {
		return ""
	}
	
	// Create stat cards
	cards := []string{
		m.renderStatCard("Forms", fmt.Sprintf("%d active / %d total", m.stats.ActiveForms, m.stats.TotalForms), m.styles.Colors.Primary),
		m.renderStatCard("Contacts", fmt.Sprintf("%d total", m.stats.TotalContacts), m.styles.Colors.Secondary),
		m.renderStatCard("Webhooks Today", fmt.Sprintf("%d sent", m.stats.WebhooksToday), m.styles.Colors.Info),
		m.renderStatCard("Connection", m.stats.ConnectionStatus, m.getStatusColor()),
	}
	
	// Join cards horizontally
	statsRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	
	// Add title
	title := m.styles.Title.Render("📊 Dashboard Overview")
	
	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		statsRow,
	)
}

func (m *Model) renderStatCard(title, value string, color lipgloss.Color) string {
	cardStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 2).
		Width(25).
		Height(5)
		
	titleStyle := m.styles.Label.Copy().
		Foreground(color)
		
	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(color)
		
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		titleStyle.Render(title),
		valueStyle.Render(value),
	)
	
	return cardStyle.Render(content)
}

func (m *Model) renderMenu() string {
	title := m.styles.Title.Render("🚀 Quick Actions")
	
	var items []string
	for i, item := range m.menuItems {
		var itemView string
		
		itemTitle := fmt.Sprintf("%s %s", item.Icon, item.Title)
		itemDesc := m.styles.Muted.Render(item.Description)
		itemKey := m.styles.Badge.Render(item.Key)
		
		itemContent := lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(lipgloss.Top, itemTitle, " ", itemKey),
			itemDesc,
		)
		
		if i == m.selected {
			itemStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(m.styles.Colors.Primary).
				Padding(1, 2).
				Width(60)
			itemView = itemStyle.Render(itemContent)
		} else {
			itemStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(m.styles.Colors.Border).
				Padding(1, 2).
				Width(60)
			itemView = itemStyle.Render(itemContent)
		}
		
		items = append(items, itemView)
	}
	
	menu := lipgloss.JoinVertical(
		lipgloss.Top,
		items...,
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		menu,
	)
}

func (m *Model) getStatusColor() lipgloss.Color {
	if m.stats == nil {
		return m.styles.Colors.Warning
	}
	
	switch m.stats.ConnectionStatus {
	case "Connected":
		return m.styles.Colors.Success
	case "Disconnected":
		return m.styles.Colors.Error
	default:
		return m.styles.Colors.Warning
	}
}

func (m *Model) renderError() string {
	errorView := m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err))
	help := m.styles.Help.Render("Check your configuration and try again")
	
	return lipgloss.JoinVertical(
		lipgloss.Center,
		errorView,
		help,
	)
}

func (m *Model) loadStats() tea.Msg {
	if m.db == nil {
		return StatsLoadedMsg{
			Error: fmt.Errorf("database client not initialized"),
		}
	}
	
	stats, err := m.db.GetStats()
	if err != nil {
		log.Error("Failed to load stats", "error", err)
		// Return partial stats even on error
		if stats == nil {
			stats = &database.Stats{
				ConnectionStatus: "Disconnected",
			}
		}
	}
	
	return StatsLoadedMsg{
		Stats: stats,
		Error: nil,
	}
}

func (m *Model) switchView(viewID int, title string) tea.Cmd {
	return func() tea.Msg {
		// This will be caught by the main app model
		type SwitchViewMsg struct {
			View  int
			Title string
		}
		return SwitchViewMsg{View: viewID, Title: title}
	}
}

// Message types
type StatsLoadedMsg struct {
	Stats *database.Stats
	Error error
}