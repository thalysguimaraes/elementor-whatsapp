package settings

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
)

type Model struct {
	config   *config.Config
	styles   *styles.Styles
	sections []Section
	selected int
	width    int
	height   int
}

type Section struct {
	Title  string
	Items  []ConfigItem
}

type ConfigItem struct {
	Label       string
	Value       string
	Key         string
	Sensitive   bool
}

func New(cfg *config.Config, s *styles.Styles) *Model {
	sections := []Section{
		{
			Title: "Cloudflare Configuration",
			Items: []ConfigItem{
				{Label: "Account ID", Value: cfg.Cloudflare.AccountID, Key: "cloudflare.account_id"},
				{Label: "API Token", Value: maskToken(cfg.Cloudflare.APIToken), Key: "cloudflare.api_token", Sensitive: true},
				{Label: "Database ID", Value: cfg.Cloudflare.DatabaseID, Key: "cloudflare.database_id"},
				{Label: "Worker URL", Value: cfg.Cloudflare.WorkerURL, Key: "cloudflare.worker_url"},
			},
		},
		{
			Title: "Z-API Configuration",
			Items: []ConfigItem{
				{Label: "Instance ID", Value: cfg.ZAPI.InstanceID, Key: "zapi.instance_id"},
				{Label: "Instance Token", Value: maskToken(cfg.ZAPI.InstanceToken), Key: "zapi.instance_token", Sensitive: true},
				{Label: "Client Token", Value: maskToken(cfg.ZAPI.ClientToken), Key: "zapi.client_token", Sensitive: true},
			},
		},
		{
			Title: "UI Preferences",
			Items: []ConfigItem{
				{Label: "Theme", Value: cfg.UI.Theme, Key: "ui.theme"},
				{Label: "Mouse Support", Value: boolToString(cfg.UI.Mouse), Key: "ui.mouse"},
				{Label: "Animations", Value: boolToString(cfg.UI.Animations), Key: "ui.animations"},
				{Label: "Vim Bindings", Value: boolToString(cfg.UI.VimBindings), Key: "ui.vim_bindings"},
				{Label: "Auto Refresh", Value: cfg.UI.AutoRefresh.String(), Key: "ui.auto_refresh"},
			},
		},
	}
	
	return &Model{
		config:   cfg,
		styles:   s,
		sections: sections,
		selected: 0,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < m.getTotalItems()-1 {
				m.selected++
			}
		case "enter":
			// Edit selected item
			// TODO: Open edit dialog
		case "s":
			// Save configuration
			// TODO: Save to file
		case "esc", "q":
			// Navigate back to dashboard
			return m, func() tea.Msg {
				type SwitchViewMsg struct {
					View  int
					Title string
				}
				return SwitchViewMsg{View: 0, Title: "Dashboard"}
			}
		}
	}
	
	return m, nil
}

func (m *Model) View() string {
	title := m.styles.Title.Render("⚙️ Settings")
	
	description := m.styles.Muted.Render("Configure application settings and preferences")
	
	// Render sections
	var sectionViews []string
	itemIndex := 0
	
	for _, section := range m.sections {
		sectionTitle := m.styles.Subtitle.Render(section.Title)
		
		var items []string
		for _, item := range section.Items {
			var itemView string
			
			label := m.styles.Label.Render(item.Label + ":")
			value := item.Value
			if value == "" {
				value = m.styles.Muted.Render("(not set)")
			} else if !item.Sensitive {
				value = m.styles.Text.Render(value)
			} else {
				value = m.styles.Warning.Render(value)
			}
			
			itemContent := lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().Width(25).Render(label),
				value,
			)
			
			if itemIndex == m.selected {
				itemView = m.styles.ActiveItem.Render("▶ " + itemContent)
			} else {
				itemView = "  " + itemContent
			}
			
			items = append(items, itemView)
			itemIndex++
		}
		
		sectionContent := lipgloss.JoinVertical(
			lipgloss.Top,
			sectionTitle,
			lipgloss.JoinVertical(lipgloss.Top, items...),
		)
		
		sectionBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.styles.Colors.Border).
			Padding(1, 2).
			Width(70).
			MarginBottom(1).
			Render(sectionContent)
		
		sectionViews = append(sectionViews, sectionBox)
	}
	
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		sectionViews...,
	)
	
	// Actions
	actions := m.styles.Help.Render("Press Enter to edit • 's' to save changes")
	
	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		description,
		"",
		content,
		"",
		actions,
	)
}

func (m *Model) getTotalItems() int {
	total := 0
	for _, section := range m.sections {
		total += len(section.Items)
	}
	return total
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func boolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}