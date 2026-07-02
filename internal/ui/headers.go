package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"leanapi/internal/model"
)

type headersPane struct {
	kv kvEditor
}

func newHeadersPane() headersPane {
	return headersPane{kv: newKVEditor("header", "Header-Name")}
}

func (h *headersPane) enabledUserHeaders() []model.Header { return h.kv.enabledPairs() }
func (h *headersPane) setHeaders(pairs []model.Header)     { h.kv.setRows(pairs) }
func (h *headersPane) syncFocus()                          { h.kv.syncFocus() }
func (h *headersPane) update(focused bool, msg tea.Msg) tea.Cmd {
	return h.kv.update(focused, msg)
}
func (h *headersPane) handleClick(msg tea.MouseMsg) bool { return h.kv.handleClick(msg) }

func (h *headersPane) view(width int, focused bool, derived []model.Header) string {
	keyW := width/3 - 2
	lines := h.kv.rowLines(width, focused)

	for _, d := range derived {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  [x] %-*s %s  (auth)", keyW, d.Key, d.Value)))
	}

	help := helpStyle.Render("ctrl+n add row · ctrl+d delete row · space toggle · tab switch field")
	body := lipgloss.JoinVertical(lipgloss.Left, append(lines, "", help)...)
	return borderFor(focused).Width(width - 2).Render(body)
}
