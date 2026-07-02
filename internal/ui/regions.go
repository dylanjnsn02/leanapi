package ui

// Pane identifies a focusable region of the app.
type Pane int

const (
	PaneMethod Pane = iota
	PaneURL
	PaneSend
	PaneTabs
	PaneParams
	PaneBody
	PaneHeaders
	PaneAuth
	PaneCookies
	PaneResponse
	PaneHistory
)
