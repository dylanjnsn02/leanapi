package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"

	"leanapi/internal/model"
)

type kvCol int

const (
	colCheckbox kvCol = iota
	colKey
	colVal
)

type kvRow struct {
	key     textinput.Model
	val     textinput.Model
	enabled bool
}

func newKVRow(keyPlaceholder string) kvRow {
	k := textinput.New()
	k.Placeholder = keyPlaceholder
	k.Prompt = ""
	v := textinput.New()
	v.Placeholder = "value"
	v.Prompt = ""
	return kvRow{key: k, val: v, enabled: true}
}

// kvEditor is a shared editable key/value/enabled row list, used by both
// the Headers pane and the Cookies pane (identical UX, different data and
// zone-id namespace so their click regions don't collide).
type kvEditor struct {
	zonePrefix     string
	keyPlaceholder string
	rows           []kvRow
	activeRow      int
	activeCol      kvCol
}

func newKVEditor(zonePrefix, keyPlaceholder string) kvEditor {
	return kvEditor{
		zonePrefix:     zonePrefix,
		keyPlaceholder: keyPlaceholder,
		rows:           []kvRow{newKVRow(keyPlaceholder)},
		activeCol:      colKey,
	}
}

func (k *kvEditor) zoneChk(i int) string { return fmt.Sprintf("%s-chk-%d", k.zonePrefix, i) }
func (k *kvEditor) zoneKey(i int) string { return fmt.Sprintf("%s-key-%d", k.zonePrefix, i) }
func (k *kvEditor) zoneVal(i int) string { return fmt.Sprintf("%s-val-%d", k.zonePrefix, i) }

func (k *kvEditor) enabledPairs() []model.Header {
	var out []model.Header
	for _, r := range k.rows {
		if !r.enabled || r.key.Value() == "" {
			continue
		}
		out = append(out, model.Header{Key: r.key.Value(), Value: r.val.Value(), Enabled: true})
	}
	return out
}

func (k *kvEditor) setRows(pairs []model.Header) {
	k.rows = nil
	for _, p := range pairs {
		row := newKVRow(k.keyPlaceholder)
		row.key.SetValue(p.Key)
		row.val.SetValue(p.Value)
		row.enabled = p.Enabled
		k.rows = append(k.rows, row)
	}
	if len(k.rows) == 0 {
		k.rows = []kvRow{newKVRow(k.keyPlaceholder)}
	}
	k.activeRow = 0
	k.activeCol = colKey
}

func (k *kvEditor) syncFocus() {
	for i := range k.rows {
		if i == k.activeRow && k.activeCol == colKey {
			k.rows[i].key.Focus()
		} else {
			k.rows[i].key.Blur()
		}
		if i == k.activeRow && k.activeCol == colVal {
			k.rows[i].val.Focus()
		} else {
			k.rows[i].val.Blur()
		}
	}
}

func (k *kvEditor) update(focused bool, msg tea.Msg) tea.Cmd {
	if !focused {
		return nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+n":
			k.rows = append(k.rows, newKVRow(k.keyPlaceholder))
			k.activeRow = len(k.rows) - 1
			k.activeCol = colKey
			k.syncFocus()
			return nil
		case "ctrl+d":
			if len(k.rows) > 1 {
				k.rows = append(k.rows[:k.activeRow], k.rows[k.activeRow+1:]...)
				if k.activeRow >= len(k.rows) {
					k.activeRow = len(k.rows) - 1
				}
				k.syncFocus()
			}
			return nil
		case "up":
			if k.activeRow > 0 {
				k.activeRow--
				k.syncFocus()
			}
			return nil
		case "down":
			if k.activeRow < len(k.rows)-1 {
				k.activeRow++
				k.syncFocus()
			}
			return nil
		case "tab":
			k.activeCol = (k.activeCol + 1) % 3
			k.syncFocus()
			return nil
		case " ":
			if k.activeCol == colCheckbox {
				k.rows[k.activeRow].enabled = !k.rows[k.activeRow].enabled
				return nil
			}
		}
	}

	var cmd tea.Cmd
	switch k.activeCol {
	case colKey:
		k.rows[k.activeRow].key, cmd = k.rows[k.activeRow].key.Update(msg)
	case colVal:
		k.rows[k.activeRow].val, cmd = k.rows[k.activeRow].val.Update(msg)
	}
	return cmd
}

// handleClick hit-tests a mouse event against each row's zones. Returns
// true if the click was consumed.
func (k *kvEditor) handleClick(msg tea.MouseMsg) bool {
	for i := range k.rows {
		if zone.Get(k.zoneChk(i)).InBounds(msg) {
			k.rows[i].enabled = !k.rows[i].enabled
			k.activeRow = i
			return true
		}
	}
	for i := range k.rows {
		if zone.Get(k.zoneKey(i)).InBounds(msg) {
			k.activeRow, k.activeCol = i, colKey
			k.syncFocus()
			return true
		}
	}
	for i := range k.rows {
		if zone.Get(k.zoneVal(i)).InBounds(msg) {
			k.activeRow, k.activeCol = i, colVal
			k.syncFocus()
			return true
		}
	}
	return false
}

// rowLines renders just the row lines (no border, no help text, no extra
// derived/read-only rows) so callers can compose their own final view.
func (k *kvEditor) rowLines(width int, focused bool) []string {
	keyW := width/3 - 2
	valW := width - keyW - 8

	var lines []string
	for i, r := range k.rows {
		box := "[ ]"
		if r.enabled {
			box = "[x]"
		}
		box = zone.Mark(k.zoneChk(i), box)

		r.key.Width = keyW
		r.val.Width = valW
		keyView := zone.Mark(k.zoneKey(i), r.key.View())
		valView := zone.Mark(k.zoneVal(i), r.val.View())

		marker := " "
		if focused && i == k.activeRow {
			marker = "›"
		}
		lines = append(lines, fmt.Sprintf("%s%s %s %s", marker, box, keyView, valView))
	}
	return lines
}
