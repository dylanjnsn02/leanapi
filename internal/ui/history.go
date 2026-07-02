package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"leanapi/internal/history"
)

func zoneHistoryRow(i int) string { return fmt.Sprintf("history-row-%d", i) }

type historyPane struct {
	entries []history.Entry // newest first
	cursor  int
	loadErr string
}

func newHistoryPane() historyPane {
	return historyPane{}
}

func (h *historyPane) reload() {
	entries, err := history.LoadAll()
	if err != nil {
		h.loadErr = err.Error()
		h.entries = nil
		return
	}
	h.loadErr = ""
	// LoadAll returns oldest-first; reverse for newest-first display.
	h.entries = make([]history.Entry, len(entries))
	for i, e := range entries {
		h.entries[len(entries)-1-i] = e
	}
	if h.cursor >= len(h.entries) {
		h.cursor = 0
	}
}

// update returns the selected entry (non-nil) when the user presses enter.
func (h *historyPane) update(msg tea.Msg) *history.Entry {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch km.String() {
	case "up", "k":
		if h.cursor > 0 {
			h.cursor--
		}
	case "down", "j":
		if h.cursor < len(h.entries)-1 {
			h.cursor++
		}
	case "enter":
		if h.cursor < len(h.entries) {
			return &h.entries[h.cursor]
		}
	}
	return nil
}

// handleClick returns the selected entry when a row is clicked.
func (h *historyPane) handleClick(msg tea.MouseMsg) *history.Entry {
	for i := range h.entries {
		if zone.Get(zoneHistoryRow(i)).InBounds(msg) {
			h.cursor = i
			return &h.entries[i]
		}
	}
	return nil
}

func (h *historyPane) view(width int) string {
	if h.loadErr != "" {
		return errorStyle.Render("Failed to load history: " + h.loadErr)
	}
	if len(h.entries) == 0 {
		return dimStyle.Render("No history yet — send a request to see it here.")
	}

	var lines []string
	for i, e := range h.entries {
		status := "—"
		if e.Response != nil {
			status = fmt.Sprintf("%d", e.Response.StatusCode)
		} else if e.Error != "" {
			status = "ERR"
		}
		line := fmt.Sprintf("%-6s %-4s %-40s %s", e.Timestamp.Local().Format("01-02 15:04:05"), e.Request.Method, truncate(e.Request.URL, 40), status)
		style := lipgloss.NewStyle()
		if i == h.cursor {
			style = style.Reverse(true)
		}
		lines = append(lines, zone.Mark(zoneHistoryRow(i), style.Render(line)))
	}

	help := helpStyle.Render("↑/↓ select · enter load into request builder · ctrl+h/esc close")
	body := lipgloss.JoinVertical(lipgloss.Left, append(lines, "", help)...)
	return body
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
