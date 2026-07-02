package ui

import "github.com/charmbracelet/lipgloss"

var (
	focusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("212"))
	blurredBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))

	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sendStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("28")).Padding(0, 2)
	tabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Underline(true)
	tabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// focusRing wraps a style so focused single-row widgets (method pill, send
// button) get a visibly distinct look without changing their line height --
// a Border() would make them 3 rows tall and break the row math the parent
// layout uses to place them.
func focusRing(s lipgloss.Style, focused bool) lipgloss.Style {
	if focused {
		return s.Reverse(true)
	}
	return s
}

var methodColors = map[string]lipgloss.Color{
	"GET":     lipgloss.Color("34"),  // green
	"POST":    lipgloss.Color("33"),  // blue
	"PUT":     lipgloss.Color("214"), // orange
	"PATCH":   lipgloss.Color("135"), // purple
	"DELETE":  lipgloss.Color("196"), // red
	"HEAD":    lipgloss.Color("244"), // gray
	"OPTIONS": lipgloss.Color("244"),
}

func methodPillStyle(method string, focused bool) lipgloss.Style {
	c, ok := methodColors[method]
	if !ok {
		c = lipgloss.Color("244")
	}
	s := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(c).Padding(0, 1)
	return focusRing(s, focused)
}

func borderFor(focused bool) lipgloss.Style {
	if focused {
		return focusedBorder
	}
	return blurredBorder
}
