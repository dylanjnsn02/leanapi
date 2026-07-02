package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"leanapi/internal/model"
)

type paramsPane struct {
	kv kvEditor
}

func newParamsPane() paramsPane {
	return paramsPane{kv: newKVEditor("param", "param-name")}
}

func (p *paramsPane) enabledParams() []model.Header { return p.kv.enabledPairs() }
func (p *paramsPane) setParams(pairs []model.Header)  { p.kv.setRows(pairs) }
func (p *paramsPane) syncFocus()                      { p.kv.syncFocus() }
func (p *paramsPane) update(focused bool, msg tea.Msg) tea.Cmd {
	return p.kv.update(focused, msg)
}
func (p *paramsPane) handleClick(msg tea.MouseMsg) bool { return p.kv.handleClick(msg) }

func (p *paramsPane) view(width int, focused bool) string {
	lines := p.kv.rowLines(width, focused)
	help := helpStyle.Render("ctrl+n add param · ctrl+d delete param · space toggle · tab switch field")
	body := lipgloss.JoinVertical(lipgloss.Left, append(lines, "", help)...)
	return borderFor(focused).Width(width - 2).Render(body)
}
