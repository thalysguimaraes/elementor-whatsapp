package webhook

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/styles"
)

type Model struct {
	config   *config.Config
	styles   *styles.Styles
	inputs   []textinput.Model
	focused  int
	width    int
	height   int
}

func New(cfg *config.Config, s *styles.Styles) *Model {
	// Create input fields
	inputs := make([]textinput.Model, 5)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Select or enter form ID"
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 40
	inputs[0].Prompt = "Form ID: "
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "John Doe"
	inputs[1].CharLimit = 100
	inputs[1].Width = 40
	inputs[1].Prompt = "Name: "
	
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "john@example.com"
	inputs[2].CharLimit = 100
	inputs[2].Width = 40
	inputs[2].Prompt = "Email: "
	
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "+1 234 567 8900"
	inputs[3].CharLimit = 20
	inputs[3].Width = 40
	inputs[3].Prompt = "Phone: "
	
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Test message..."
	inputs[4].CharLimit = 500
	inputs[4].Width = 40
	inputs[4].Prompt = "Message: "
	
	return &Model{
		config:  cfg,
		styles:  s,
		inputs:  inputs,
		focused: 0,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.nextInput()
		case "shift+tab", "up":
			m.prevInput()
		case "enter":
			if m.focused == len(m.inputs)-1 {
				// Submit form
				// TODO: Send test webhook
			} else {
				m.nextInput()
			}
		}
	}
	
	// Update the focused input
	for i := range m.inputs {
		if i == m.focused {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	title := m.styles.Title.Render("🧪 Test Webhook")
	
	description := m.styles.Muted.Render("Send a test webhook to verify your configuration")
	
	// Form inputs
	var fields []string
	for i, input := range m.inputs {
		style := lipgloss.NewStyle().MarginBottom(1)
		if i == m.focused {
			style = style.Foreground(m.styles.Colors.Primary)
		}
		fields = append(fields, style.Render(input.View()))
	}
	
	form := lipgloss.JoinVertical(
		lipgloss.Top,
		fields...,
	)
	
	// Submit button
	submitStyle := m.styles.Button
	if m.focused == len(m.inputs)-1 {
		submitStyle = submitStyle.Background(m.styles.Colors.Success)
	}
	submit := submitStyle.Render("Send Test Webhook")
	
	// Result area (placeholder)
	resultTitle := m.styles.Subtitle.Render("Results")
	resultBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Colors.Border).
		Padding(1, 2).
		Width(60).
		Height(8).
		Render("Test results will appear here...")
	
	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		description,
		"",
		form,
		"",
		submit,
		"",
		resultTitle,
		resultBox,
	)
}

func (m *Model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

func (m *Model) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}