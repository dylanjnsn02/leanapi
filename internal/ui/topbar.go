package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"leanapi/internal/model"
)

const (
	zoneMethod = "method"
	zoneURL    = "url"
	zoneSend   = "send"
)

type topbar struct {
	methodIdx int
	url       textinput.Model
	spin      spinner.Model
	sending   bool
}

func newTopbar() topbar {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/users"
	ti.Prompt = ""

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return topbar{url: ti, spin: sp}
}

func (t *topbar) method() string {
	return model.Methods[t.methodIdx]
}

func (t *topbar) cycleMethod(forward bool) {
	n := len(model.Methods)
	if forward {
		t.methodIdx = (t.methodIdx + 1) % n
	} else {
		t.methodIdx = (t.methodIdx - 1 + n) % n
	}
}

func (t *topbar) update(focus Pane, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	if t.sending {
		t.spin, cmd = t.spin.Update(msg)
	}
	if focus == PaneURL {
		var c2 tea.Cmd
		t.url, c2 = t.url.Update(msg)
		cmd = tea.Batch(cmd, c2)
	}
	return cmd
}

func (t *topbar) view(width int, focus Pane) string {
	pill := zone.Mark(zoneMethod, methodPillStyle(t.method(), focus == PaneMethod).Render(t.method()))

	sendLabel := "Send"
	if t.sending {
		sendLabel = t.spin.View() + " Sending"
	}
	sendBtn := zone.Mark(zoneSend, focusRing(sendStyle, focus == PaneSend).Render(sendLabel))

	urlWidth := width - lipgloss.Width(pill) - lipgloss.Width(sendBtn) - 4
	if urlWidth < 10 {
		urlWidth = 10
	}
	t.url.Width = urlWidth - 2

	urlBox := zone.Mark(zoneURL, borderFor(focus == PaneURL).Width(urlWidth).Render(t.url.View()))

	return lipgloss.JoinHorizontal(lipgloss.Center, pill, " ", urlBox, " ", sendBtn)
}
