package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

const zoneBody = "body-pane"

type bodyPane struct {
	ta textarea.Model
}

func newBodyPane() bodyPane {
	ta := textarea.New()
	ta.Placeholder = `{"key": "value"}`
	ta.ShowLineNumbers = false
	return bodyPane{ta: ta}
}

func (b *bodyPane) update(focused bool, msg tea.Msg) tea.Cmd {
	if !focused {
		return nil
	}
	var cmd tea.Cmd
	b.ta, cmd = b.ta.Update(msg)
	return cmd
}

func (b *bodyPane) setFocus(focused bool) {
	if focused {
		b.ta.Focus()
	} else {
		b.ta.Blur()
	}
}

// scroll moves the textarea's view by nudging the cursor. bubbles/textarea
// (unlike viewport) has no native mouse-wheel handling, so wheel events are
// translated into repeated cursor moves, which drags the view along with it.
func (b *bodyPane) scroll(up bool, lines int) {
	for i := 0; i < lines; i++ {
		if up {
			b.ta.CursorUp()
		} else {
			b.ta.CursorDown()
		}
	}
}

func (b *bodyPane) view(width, height int, focused bool) string {
	b.ta.SetWidth(width - 2)
	b.ta.SetHeight(height - 2)
	return zone.Mark(zoneBody, borderFor(focused).Width(width-2).Height(height-2).Render(b.ta.View()))
}
