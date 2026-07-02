package ui

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"leanapi/internal/history"
	"leanapi/internal/httpclient"
	"leanapi/internal/model"
)

const copyFeedbackDuration = 1500 * time.Millisecond

// copyResultMsg reports whether a clipboard write (a blocking shell-out to
// pbcopy/xclip/etc.) succeeded; clearCopyFeedbackMsg clears the transient
// "Copied!" indicator a moment later.
type copyResultMsg struct{ ok bool }
type clearCopyFeedbackMsg struct{}

func copyToClipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		return copyResultMsg{ok: clipboard.WriteAll(text) == nil}
	}
}

func clearCopyFeedbackCmd() tea.Cmd {
	return tea.Tick(copyFeedbackDuration, func(time.Time) tea.Msg { return clearCopyFeedbackMsg{} })
}

type viewMode int

const (
	viewRequest viewMode = iota
	viewHistory
)

// contentPanes are the panes reachable only while their tab is active.
var contentPanes = map[reqTab]Pane{
	TabParams:  PaneParams,
	TabBody:    PaneBody,
	TabHeaders: PaneHeaders,
	TabAuth:    PaneAuth,
	TabCookies: PaneCookies,
}

type RootModel struct {
	topbar   topbar
	tabs     tabStrip
	params   paramsPane
	body     bodyPane
	headers  headersPane
	auth     authPane
	cookies  cookiesPane
	response responsePane
	history  historyPane

	focus Pane
	mode  viewMode

	width, height int
}

func NewRootModel() RootModel {
	m := RootModel{
		topbar:   newTopbar(),
		tabs:     newTabStrip(),
		params:   newParamsPane(),
		body:     newBodyPane(),
		headers:  newHeadersPane(),
		auth:     newAuthPane(),
		cookies:  newCookiesPane(),
		response: newResponsePane(),
		history:  newHistoryPane(),
		focus:    PaneURL,
	}
	m.topbar.url.Focus()
	return m
}

func (m RootModel) Init() tea.Cmd {
	return textinput.Blink
}

func appendHistoryCmd(e history.Entry) tea.Cmd {
	return func() tea.Msg {
		_ = history.Append(e) // best-effort: history persistence must never block or crash the UI
		return nil
	}
}

func (m *RootModel) currentContentPane() Pane {
	return contentPanes[m.tabs.active]
}

func (m *RootModel) buildRequest() model.Request {
	return model.Request{
		Method:  m.topbar.method(),
		URL:     m.topbar.url.Value(),
		Params:  m.params.enabledParams(),
		Headers: m.headers.enabledUserHeaders(),
		Cookies: m.cookies.enabledCookies(),
		Auth:    m.auth.config(),
		Body:    m.body.ta.Value(),
	}
}

func (m *RootModel) applyRequest(r model.Request) {
	for i, meth := range model.Methods {
		if meth == r.Method {
			m.topbar.methodIdx = i
			break
		}
	}
	m.topbar.url.SetValue(r.URL)

	m.params = newParamsPane()
	m.params.setParams(r.Params)

	m.headers = newHeadersPane()
	m.headers.setHeaders(r.Headers)

	m.cookies = newCookiesPane()
	m.cookies.setCookies(r.Cookies)

	m.auth = newAuthPane()
	for i, t := range authTypes {
		if t == r.Auth.Type {
			m.auth.typeIdx = i
			break
		}
	}
	m.auth.username.SetValue(r.Auth.Username)
	m.auth.password.SetValue(r.Auth.Password)
	m.auth.token.SetValue(r.Auth.Token)
	m.auth.keyName.SetValue(r.Auth.APIKeyName)
	m.auth.keyValue.SetValue(r.Auth.APIKeyValue)
	m.auth.inQuery = r.Auth.APIKeyPlacement == model.APIKeyInQuery

	m.body = newBodyPane()
	m.body.ta.SetValue(r.Body)
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case httpclient.ResponseMsg:
		m.topbar.sending = false
		m.response.setResponse(msg.Snapshot, msg.Err, msg.DurationMS)
		entry := history.Entry{
			ID:         history.NewID(),
			Request:    m.buildRequest(),
			Response:   msg.Snapshot,
			DurationMS: msg.DurationMS,
		}
		if msg.Err != nil {
			entry.Error = msg.Err.Error()
		}
		entry.Timestamp = time.Now()
		return m, appendHistoryCmd(entry)

	case copyResultMsg:
		m.response.setCopyFeedback(msg.ok)
		return m, clearCopyFeedbackCmd()

	case clearCopyFeedbackMsg:
		m.response.clearCopyFeedback()
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	cmd := m.dispatch(msg)
	return m, cmd
}

func (m RootModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+h":
		if m.mode == viewHistory {
			m.mode = viewRequest
		} else {
			m.mode = viewHistory
			m.history.reload()
		}
		return m, nil
	}

	if m.mode == viewHistory {
		switch msg.String() {
		case "esc":
			m.mode = viewRequest
			return m, nil
		}
		if entry := m.history.update(msg); entry != nil {
			m.applyRequest(entry.Request)
			m.mode = viewRequest
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		if m.focus == PaneParams || m.focus == PaneBody || m.focus == PaneHeaders || m.focus == PaneAuth || m.focus == PaneCookies {
			m.focus = PaneTabs
			m.syncFocus()
			return m, nil
		}
	case "enter":
		if m.focus == PaneURL || m.focus == PaneSend {
			return m.startSend()
		}
	case "y":
		if m.focus == PaneResponse {
			return m.triggerCopy()
		}
	case "tab":
		if m.focus == PaneMethod {
			m.focus = PaneURL
		} else if m.focus == PaneURL {
			m.focus = PaneSend
		} else if m.focus == PaneSend {
			m.focus = PaneTabs
		} else if m.focus == PaneTabs {
			m.focus = m.currentContentPane()
		} else {
			// inside a content/response pane: let it consume tab locally
			break
		}
		m.syncFocus()
		return m, nil
	case "shift+tab":
		switch m.focus {
		case PaneURL:
			m.focus = PaneMethod
		case PaneSend:
			m.focus = PaneURL
		case PaneTabs:
			m.focus = PaneSend
		case PaneParams, PaneBody, PaneHeaders, PaneAuth, PaneCookies:
			m.focus = PaneTabs
		}
		m.syncFocus()
		return m, nil
	case "left", "right":
		if m.focus == PaneMethod {
			m.topbar.cycleMethod(msg.String() == "right")
			return m, nil
		}
		if m.focus == PaneTabs {
			m.tabs.update(msg)
			return m, nil
		}
	}

	cmd := m.dispatch(msg)
	return m, cmd
}

func (m *RootModel) syncFocus() {
	m.body.setFocus(m.focus == PaneBody)
	if m.focus == PaneParams {
		m.params.syncFocus()
	}
	if m.focus == PaneHeaders {
		m.headers.syncFocus()
	}
	if m.focus == PaneAuth {
		m.auth.syncFocus()
	}
	if m.focus == PaneCookies {
		m.cookies.syncFocus()
	}
}

func (m RootModel) startSend() (tea.Model, tea.Cmd) {
	if m.topbar.url.Value() == "" || m.topbar.sending {
		return m, nil
	}
	m.topbar.sending = true
	req := m.buildRequest()
	return m, tea.Batch(httpclient.SendRequestCmd(req), m.topbar.spin.Tick)
}

// triggerCopy copies whichever response sub-view (Body/Headers/Cookies) is
// currently active to the OS clipboard.
func (m RootModel) triggerCopy() (tea.Model, tea.Cmd) {
	text := m.response.activePlainText()
	if text == "" {
		return m, nil
	}
	return m, copyToClipboardCmd(text)
}

func (m *RootModel) dispatch(msg tea.Msg) tea.Cmd {
	switch m.focus {
	case PaneParams:
		return m.params.update(true, msg)
	case PaneBody:
		return m.body.update(true, msg)
	case PaneHeaders:
		return m.headers.update(true, msg)
	case PaneAuth:
		return m.auth.update(true, msg)
	case PaneCookies:
		return m.cookies.update(true, msg)
	case PaneResponse:
		return m.response.update(true, msg)
	default:
		return m.topbar.update(m.focus, msg)
	}
}

func (m RootModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if tea.MouseEvent(msg).IsWheel() {
		return m.handleWheel(msg)
	}

	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	if m.mode == viewHistory {
		if entry := m.history.handleClick(msg); entry != nil {
			m.applyRequest(entry.Request)
			m.mode = viewRequest
		}
		return m, nil
	}

	if zone.Get(zoneMethod).InBounds(msg) {
		m.topbar.cycleMethod(true)
		m.focus = PaneMethod
		return m, nil
	}
	if zone.Get(zoneURL).InBounds(msg) {
		m.focus = PaneURL
		m.syncFocus()
		return m, nil
	}
	if zone.Get(zoneSend).InBounds(msg) {
		m.focus = PaneSend
		return m.startSend()
	}
	for _, tb := range tabOrderList {
		if zone.Get(zoneForTab(tb)).InBounds(msg) {
			m.tabs.active = tb
			m.focus = contentPanes[tb]
			m.syncFocus()
			return m, nil
		}
	}
	if zone.Get(zoneBody).InBounds(msg) {
		m.focus = PaneBody
		m.syncFocus()
		return m, nil
	}
	if m.params.handleClick(msg) {
		m.focus = PaneParams
		return m, nil
	}
	if m.headers.handleClick(msg) {
		m.focus = PaneHeaders
		return m, nil
	}
	if m.auth.handleClick(msg) {
		m.focus = PaneAuth
		return m, nil
	}
	if m.cookies.handleClick(msg) {
		m.focus = PaneCookies
		return m, nil
	}
	if m.response.copyZoneHit(msg) {
		m.focus = PaneResponse
		return m.triggerCopy()
	}
	if m.response.handleClick(msg) {
		m.focus = PaneResponse
		return m, nil
	}

	return m, nil
}

// handleWheel scrolls whichever scrollable pane the mouse is currently
// hovering over, independent of keyboard focus -- matching how scrolling
// works in normal apps (you don't have to click into a pane first).
func (m RootModel) handleWheel(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.mode == viewHistory {
		return m, nil
	}

	if zone.Get(zoneRespViewport).InBounds(msg) {
		cmd := m.response.update(true, msg)
		return m, cmd
	}

	if m.tabs.active == TabBody && zone.Get(zoneBody).InBounds(msg) {
		m.body.scroll(msg.Button == tea.MouseButtonWheelUp, 3)
		return m, nil
	}

	return m, nil
}

func (m RootModel) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var out string
	if m.mode == viewHistory {
		title := lipgloss.NewStyle().Bold(true).Render("History")
		out = lipgloss.JoinVertical(lipgloss.Left, title, "", m.history.view(m.width))
	} else {
		top := m.topbar.view(m.width, m.focus)
		tabsLine := m.tabs.view(m.focus == PaneTabs)

		contentHeight := 8
		var content string
		switch m.tabs.active {
		case TabParams:
			content = m.params.view(m.width, m.focus == PaneParams)
		case TabBody:
			content = m.body.view(m.width, contentHeight, m.focus == PaneBody)
		case TabHeaders:
			derived := httpclient.DerivedHeaders(m.auth.config())
			content = m.headers.view(m.width, m.focus == PaneHeaders, derived)
		case TabAuth:
			content = m.auth.view(m.width, m.focus == PaneAuth)
		case TabCookies:
			content = m.cookies.view(m.width, m.focus == PaneCookies)
		}

		usedHeight := 3 + 1 + contentHeight + 1 // topbar + tabstrip + content + help line
		responseHeight := m.height - usedHeight
		if responseHeight < 4 {
			responseHeight = 4
		}
		resp := m.response.view(m.width, responseHeight, m.focus == PaneResponse)

		help := helpStyle.Render("tab/shift+tab move · esc back to tabs · y copy response · ctrl+h history · ctrl+c quit")

		out = lipgloss.JoinVertical(lipgloss.Left, top, tabsLine, content, resp, help)
	}

	return zone.Scan(out)
}
