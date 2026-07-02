package ui

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"leanapi/internal/history"
	"leanapi/internal/jsonview"
)

type responseView int

const (
	viewBody responseView = iota
	viewHeaders
	viewCookies
)

const (
	zoneRespToggleBody    = "response-toggle-body"
	zoneRespToggleHeaders = "response-toggle-headers"
	zoneRespToggleCookies = "response-toggle-cookies"
	zoneRespCopy          = "response-copy"
	zoneRespViewport      = "response-viewport"
)

type responsePane struct {
	vp viewport.Model

	hasResponse bool
	statusCode  int
	status      string
	durationMS  int64
	size        int64
	errText     string

	bodyView    string
	headersView string
	cookiesView string
	// plain* mirror the styled views above but without ANSI escapes, so
	// copying to the OS clipboard doesn't paste raw escape codes.
	plainBody    string
	plainHeaders string
	plainCookies string
	active       responseView

	copyMsg string // transient "Copied!" / "Copy failed" feedback
}

func newResponsePane() responsePane {
	return responsePane{vp: viewport.New(0, 0)}
}

func (r *responsePane) setResponse(snap *history.ResponseSnapshot, err error, durationMS int64) {
	r.durationMS = durationMS
	r.active = viewBody

	if err != nil {
		r.hasResponse = false
		r.errText = err.Error()
		r.vp.SetContent(errorStyle.Render("Error: " + err.Error()))
		return
	}

	r.hasResponse = true
	r.errText = ""
	r.statusCode = snap.StatusCode
	r.status = snap.Status
	r.size = snap.Size

	pretty, isJSON := jsonview.Pretty([]byte(snap.Body))
	r.plainBody = pretty
	if isJSON {
		r.bodyView = jsonview.Highlight(pretty)
	} else {
		r.bodyView = pretty
	}

	var hb, hbPlain strings.Builder
	keys := make([]string, 0, len(snap.Headers))
	for k := range snap.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range snap.Headers[k] {
			fmt.Fprintf(&hb, "%s: %s\n", dimStyle.Render(k), v)
			fmt.Fprintf(&hbPlain, "%s: %s\n", k, v)
		}
	}
	r.headersView = hb.String()
	r.plainHeaders = hbPlain.String()

	r.cookiesView = renderCookies(snap.Headers, true)
	r.plainCookies = renderCookies(snap.Headers, false)

	r.vp.SetContent(r.bodyView)
	r.vp.GotoTop()
}

// renderCookies parses Set-Cookie headers (reusing net/http's own cookie
// parser instead of hand-rolling one) and formats each cookie's name/value
// plus its attributes. styled controls whether the cookie name gets ANSI
// highlighting -- disabled for the clipboard-copy variant.
func renderCookies(headers map[string][]string, styled bool) string {
	resp := &http.Response{Header: http.Header(headers)}
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		msg := "No cookies set in this response."
		if styled {
			return dimStyle.Render(msg)
		}
		return msg
	}

	var b strings.Builder
	for i, c := range cookies {
		if i > 0 {
			b.WriteString("\n")
		}
		name := c.Name
		if styled {
			name = keyStyleResp.Render(name)
		}
		fmt.Fprintf(&b, "%s = %s\n", name, c.Value)
		if c.Domain != "" {
			fmt.Fprintf(&b, "  Domain: %s\n", c.Domain)
		}
		if c.Path != "" {
			fmt.Fprintf(&b, "  Path: %s\n", c.Path)
		}
		if !c.Expires.IsZero() {
			fmt.Fprintf(&b, "  Expires: %s\n", c.Expires.Format("2006-01-02 15:04:05 MST"))
		}
		if c.MaxAge != 0 {
			fmt.Fprintf(&b, "  Max-Age: %d\n", c.MaxAge)
		}
		if c.Secure {
			b.WriteString("  Secure\n")
		}
		if c.HttpOnly {
			b.WriteString("  HttpOnly\n")
		}
		if c.SameSite != http.SameSiteDefaultMode {
			fmt.Fprintf(&b, "  SameSite: %v\n", c.SameSite)
		}
	}
	return b.String()
}

var keyStyleResp = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)

func (r *responsePane) contentFor(v responseView) string {
	switch v {
	case viewHeaders:
		return r.headersView
	case viewCookies:
		return r.cookiesView
	default:
		return r.bodyView
	}
}

// activePlainText returns the unstyled (no ANSI) text for whichever view is
// currently active, suitable for writing to the OS clipboard.
func (r *responsePane) activePlainText() string {
	switch r.active {
	case viewHeaders:
		return r.plainHeaders
	case viewCookies:
		return r.plainCookies
	default:
		return r.plainBody
	}
}

func (r *responsePane) update(focused bool, msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok && focused && km.String() == "tab" {
		r.setActive((r.active + 1) % 3)
		return nil
	}
	var cmd tea.Cmd
	r.vp, cmd = r.vp.Update(msg)
	return cmd
}

func (r *responsePane) setActive(v responseView) {
	r.active = v
	r.vp.SetContent(r.contentFor(v))
	r.vp.GotoTop()
}

func (r *responsePane) setCopyFeedback(ok bool) {
	if ok {
		r.copyMsg = "Copied!"
	} else {
		r.copyMsg = "Copy failed"
	}
}

func (r *responsePane) clearCopyFeedback() {
	r.copyMsg = ""
}

// copyZoneHit reports whether msg landed on the [Copy] button. Copying
// itself is a side effect (shelling out to pbcopy/xclip/etc.), so the
// actual clipboard write is triggered by the caller as a tea.Cmd rather
// than performed here.
func (r *responsePane) copyZoneHit(msg tea.MouseMsg) bool {
	return zone.Get(zoneRespCopy).InBounds(msg)
}

func (r *responsePane) handleClick(msg tea.MouseMsg) bool {
	switch {
	case zone.Get(zoneRespToggleBody).InBounds(msg):
		r.setActive(viewBody)
		return true
	case zone.Get(zoneRespToggleHeaders).InBounds(msg):
		r.setActive(viewHeaders)
		return true
	case zone.Get(zoneRespToggleCookies).InBounds(msg):
		r.setActive(viewCookies)
		return true
	case zone.Get(zoneRespViewport).InBounds(msg):
		return true
	}
	return false
}

func (r *responsePane) view(width, height int, focused bool) string {
	statusLine := dimStyle.Render("Response — no request sent yet")
	if r.errText != "" {
		statusLine = errorStyle.Render("Response — error — " + fmt.Sprintf("%dms", r.durationMS))
	} else if r.hasResponse {
		statusStyle := dimStyle
		if r.statusCode >= 200 && r.statusCode < 300 {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
		} else if r.statusCode >= 400 {
			statusStyle = errorStyle
		}

		toggle := func(label string, v responseView, zoneID string) string {
			s := tabInactive
			if r.active == v {
				s = tabActive
			}
			return zone.Mark(zoneID, s.Render(label))
		}
		toggles := lipgloss.JoinHorizontal(lipgloss.Top,
			toggle("Body", viewBody, zoneRespToggleBody), "  ",
			toggle("Headers", viewHeaders, zoneRespToggleHeaders), "  ",
			toggle("Cookies", viewCookies, zoneRespToggleCookies),
		)

		copyBtn := zone.Mark(zoneRespCopy, tabInactive.Render("[Copy]"))
		feedback := ""
		if r.copyMsg != "" {
			style := dimStyle
			if r.copyMsg == "Copied!" {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
			} else {
				style = errorStyle
			}
			feedback = "  " + style.Render(r.copyMsg)
		}

		statusLine = lipgloss.JoinHorizontal(lipgloss.Top,
			fmt.Sprintf("%s  •  %dms  •  %d bytes  •  ", statusStyle.Render(r.status), r.durationMS, r.size),
			toggles, "  ", copyBtn, feedback,
		)
	}

	r.vp.Width = width - 2
	r.vp.Height = height - 3

	viewportBox := zone.Mark(zoneRespViewport, borderFor(focused).Width(width-2).Height(height-3).Render(r.vp.View()))
	return lipgloss.JoinVertical(lipgloss.Left, statusLine, viewportBox)
}
