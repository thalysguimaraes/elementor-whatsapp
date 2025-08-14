package styles

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	// Layout
	Header     lipgloss.Style
	Footer     lipgloss.Style
	Content    lipgloss.Style
	Sidebar    lipgloss.Style
	
	// Typography
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Label      lipgloss.Style
	Text       lipgloss.Style
	Muted      lipgloss.Style
	
	// Navigation
	Breadcrumb lipgloss.Style
	MenuItem   lipgloss.Style
	ActiveItem lipgloss.Style
	
	// Status
	StatusBar  lipgloss.Style
	Success    lipgloss.Style
	Warning    lipgloss.Style
	Error      lipgloss.Style
	Info       lipgloss.Style
	
	// Interactive
	Button     lipgloss.Style
	Input      lipgloss.Style
	Select     lipgloss.Style
	Checkbox   lipgloss.Style
	
	// Table
	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableCell   lipgloss.Style
	
	// Misc
	Help       lipgloss.Style
	Spinner    lipgloss.Style
	Progress   lipgloss.Style
	Badge      lipgloss.Style
	
	// Colors (for reference)
	Colors     ColorScheme
}

type ColorScheme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Info      lipgloss.Color
	Muted     lipgloss.Color
	BgPrimary lipgloss.Color
	BgSecondary lipgloss.Color
	Border    lipgloss.Color
}

var (
	// Charm theme - signature pink/purple
	CharmColors = ColorScheme{
		Primary:     lipgloss.Color("#FF79C6"),
		Secondary:   lipgloss.Color("#BD93F9"),
		Success:     lipgloss.Color("#50FA7B"),
		Warning:     lipgloss.Color("#F1FA8C"),
		Error:       lipgloss.Color("#FF5555"),
		Info:        lipgloss.Color("#8BE9FD"),
		Muted:       lipgloss.Color("#6272A4"),
		BgPrimary:   lipgloss.Color("#282A36"),
		BgSecondary: lipgloss.Color("#44475A"),
		Border:      lipgloss.Color("#6272A4"),
	}
	
	// Default/Light theme
	DefaultColors = ColorScheme{
		Primary:     lipgloss.Color("#5A56E0"),
		Secondary:   lipgloss.Color("#7A77E6"),
		Success:     lipgloss.Color("#52C41A"),
		Warning:     lipgloss.Color("#FAAD14"),
		Error:       lipgloss.Color("#F5222D"),
		Info:        lipgloss.Color("#1890FF"),
		Muted:       lipgloss.Color("#8C8C8C"),
		BgPrimary:   lipgloss.Color("#FFFFFF"),
		BgSecondary: lipgloss.Color("#F5F5F5"),
		Border:      lipgloss.Color("#D9D9D9"),
	}
	
	// Dark theme
	DarkColors = ColorScheme{
		Primary:     lipgloss.Color("#61AFEF"),
		Secondary:   lipgloss.Color("#C678DD"),
		Success:     lipgloss.Color("#98C379"),
		Warning:     lipgloss.Color("#E5C07B"),
		Error:       lipgloss.Color("#E06C75"),
		Info:        lipgloss.Color("#61AFEF"),
		Muted:       lipgloss.Color("#5C6370"),
		BgPrimary:   lipgloss.Color("#282C34"),
		BgSecondary: lipgloss.Color("#3E4451"),
		Border:      lipgloss.Color("#4B5263"),
	}
)

func NewStyles(theme string) *Styles {
	var colors ColorScheme
	
	switch theme {
	case "charm":
		colors = CharmColors
	case "dark":
		colors = DarkColors
	case "light", "default":
		colors = DefaultColors
	default:
		colors = CharmColors
	}
	
	s := &Styles{
		Colors: colors,
	}
	
	// Layout styles
	s.Header = lipgloss.NewStyle().
		Background(colors.BgSecondary).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colors.Border)
		
	s.Footer = lipgloss.NewStyle().
		Background(colors.BgSecondary).
		Padding(0, 2).
		Height(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(colors.Border)
		
	s.Content = lipgloss.NewStyle().
		Padding(1, 2)
		
	s.Sidebar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true).
		BorderForeground(colors.Border).
		Padding(1, 2)
	
	// Typography styles
	s.Title = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		MarginBottom(1)
		
	s.Subtitle = lipgloss.NewStyle().
		Foreground(colors.Secondary).
		MarginBottom(1)
		
	s.Label = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Bold(true)
		
	s.Text = lipgloss.NewStyle().
		Foreground(lipgloss.Color(""))
		
	s.Muted = lipgloss.NewStyle().
		Foreground(colors.Muted)
	
	// Navigation styles
	s.Breadcrumb = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)
		
	s.MenuItem = lipgloss.NewStyle().
		Padding(0, 2)
		
	s.ActiveItem = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Background(colors.BgSecondary).
		Padding(0, 2).
		Bold(true)
	
	// Status styles
	s.StatusBar = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)
		
	s.Success = lipgloss.NewStyle().
		Foreground(colors.Success).
		Bold(true)
		
	s.Warning = lipgloss.NewStyle().
		Foreground(colors.Warning).
		Bold(true)
		
	s.Error = lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)
		
	s.Info = lipgloss.NewStyle().
		Foreground(colors.Info)
	
	// Interactive styles
	s.Button = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF")).
		Background(colors.Primary).
		Padding(0, 3).
		MarginRight(1)
		
	s.Input = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.Border).
		Padding(0, 1)
		
	s.Select = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.Border)
		
	s.Checkbox = lipgloss.NewStyle().
		Foreground(colors.Primary)
	
	// Table styles
	s.TableHeader = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colors.Border)
		
	s.TableRow = lipgloss.NewStyle().
		Padding(0, 1)
		
	s.TableCell = lipgloss.NewStyle().
		Padding(0, 1)
	
	// Misc styles
	s.Help = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Italic(true)
		
	s.Spinner = lipgloss.NewStyle().
		Foreground(colors.Primary)
		
	s.Progress = lipgloss.NewStyle().
		Foreground(colors.Primary)
		
	s.Badge = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF")).
		Background(colors.Primary).
		Padding(0, 1).
		MarginRight(1)
	
	return s
}

// Helper functions for consistent styling

func (s *Styles) RenderSuccess(text string) string {
	return s.Success.Render("✓ " + text)
}

func (s *Styles) RenderError(text string) string {
	return s.Error.Render("✗ " + text)
}

func (s *Styles) RenderWarning(text string) string {
	return s.Warning.Render("⚠ " + text)
}

func (s *Styles) RenderInfo(text string) string {
	return s.Info.Render("ℹ " + text)
}

func (s *Styles) RenderBadge(text string, color lipgloss.Color) string {
	style := s.Badge.Copy().Background(color)
	return style.Render(text)
}