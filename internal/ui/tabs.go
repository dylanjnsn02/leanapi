package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type reqTab int

const (
	TabParams reqTab = iota
	TabBody
	TabHeaders
	TabAuth
	TabCookies
)

var tabLabels = map[reqTab]string{
	TabParams:  "Params",
	TabBody:    "Body",
	TabHeaders: "Headers",
	TabAuth:    "Auth",
	TabCookies: "Cookies",
}

var tabOrderList = []reqTab{TabParams, TabBody, TabHeaders, TabAuth, TabCookies}

func zoneForTab(t reqTab) string {
	return "tab-" + tabLabels[t]
}

type tabStrip struct {
	active reqTab
}

func newTabStrip() tabStrip {
	return tabStrip{active: TabBody}
}

func (t *tabStrip) update(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "right", "l":
			t.cycle(true)
		case "left", "h":
			t.cycle(false)
		}
	}
}

func (t *tabStrip) cycle(forward bool) {
	idx := 0
	for i, x := range tabOrderList {
		if x == t.active {
			idx = i
			break
		}
	}
	if forward {
		idx = (idx + 1) % len(tabOrderList)
	} else {
		idx = (idx - 1 + len(tabOrderList)) % len(tabOrderList)
	}
	t.active = tabOrderList[idx]
}

// view renders the Body/Headers/Auth tab strip. focused indicates the strip
// itself has keyboard focus (distinct from which tab is active): the
// focused tab gets a surrounding marker so the user can tell tab vs. arrow
// navigation is currently live here.
func (t *tabStrip) view(focused bool) string {
	var parts []string
	for _, tb := range tabOrderList {
		label := tabLabels[tb]
		if focused && tb == t.active {
			label = "‹ " + label + " ›"
		} else {
			label = "  " + label + "  "
		}
		style := tabInactive
		if tb == t.active {
			style = tabActive
		}
		rendered := zone.Mark(zoneForTab(tb), style.Render(label))
		parts = append(parts, rendered)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
