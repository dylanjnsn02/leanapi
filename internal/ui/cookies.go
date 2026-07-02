package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"leanapi/internal/model"
)

type cookiesPane struct {
	kv kvEditor
}

func newCookiesPane() cookiesPane {
	return cookiesPane{kv: newKVEditor("cookie", "cookie-name")}
}

func (c *cookiesPane) enabledCookies() []model.Header { return c.kv.enabledPairs() }
func (c *cookiesPane) setCookies(pairs []model.Header)  { c.kv.setRows(pairs) }
func (c *cookiesPane) syncFocus()                       { c.kv.syncFocus() }
func (c *cookiesPane) update(focused bool, msg tea.Msg) tea.Cmd {
	return c.kv.update(focused, msg)
}
func (c *cookiesPane) handleClick(msg tea.MouseMsg) bool { return c.kv.handleClick(msg) }

func (c *cookiesPane) view(width int, focused bool) string {
	lines := c.kv.rowLines(width, focused)
	help := helpStyle.Render("ctrl+n add cookie · ctrl+d delete cookie · space toggle · tab switch field")
	body := lipgloss.JoinVertical(lipgloss.Left, append(lines, "", help)...)
	return borderFor(focused).Width(width - 2).Render(body)
}
