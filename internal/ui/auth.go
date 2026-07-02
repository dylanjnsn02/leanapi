package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"leanapi/internal/model"
)

var authTypes = []model.AuthType{model.AuthNone, model.AuthBasic, model.AuthBearer, model.AuthAPIKey}

var authTypeLabels = map[model.AuthType]string{
	model.AuthNone:   "No Auth",
	model.AuthBasic:  "Basic",
	model.AuthBearer: "Bearer Token",
	model.AuthAPIKey: "API Key",
}

type authField int

const (
	fieldType authField = iota
	fieldUsername
	fieldPassword
	fieldToken
	fieldAPIKeyName
	fieldAPIKeyValue
	fieldAPIKeyPlacement
)

func zoneAuthField(f authField) string { return fmt.Sprintf("auth-field-%d", f) }

type authPane struct {
	typeIdx int

	username textinput.Model
	password textinput.Model
	token    textinput.Model
	keyName  textinput.Model
	keyValue textinput.Model
	inQuery  bool

	active authField
}

func newAuthPane() authPane {
	mk := func(placeholder string, mask bool) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Prompt = ""
		if mask {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		return ti
	}
	return authPane{
		username: mk("username", false),
		password: mk("password", true),
		token:    mk("token", false),
		keyName:  mk("X-API-Key", false),
		keyValue: mk("value", false),
		active:   fieldType,
	}
}

func (a *authPane) authType() model.AuthType { return authTypes[a.typeIdx] }

func (a *authPane) config() model.AuthConfig {
	placement := model.APIKeyInHeader
	if a.inQuery {
		placement = model.APIKeyInQuery
	}
	return model.AuthConfig{
		Type:            a.authType(),
		Username:        a.username.Value(),
		Password:        a.password.Value(),
		Token:           a.token.Value(),
		APIKeyName:      a.keyName.Value(),
		APIKeyValue:     a.keyValue.Value(),
		APIKeyPlacement: placement,
	}
}

// fieldsForType lists the focusable fields (after fieldType) for the
// current auth type, in tab order.
func (a *authPane) fieldsForType() []authField {
	switch a.authType() {
	case model.AuthBasic:
		return []authField{fieldUsername, fieldPassword}
	case model.AuthBearer:
		return []authField{fieldToken}
	case model.AuthAPIKey:
		return []authField{fieldAPIKeyName, fieldAPIKeyValue, fieldAPIKeyPlacement}
	default:
		return nil
	}
}

func (a *authPane) syncFocus() {
	all := []*textinput.Model{&a.username, &a.password, &a.token, &a.keyName, &a.keyValue}
	for _, f := range all {
		f.Blur()
	}
	switch a.active {
	case fieldUsername:
		a.username.Focus()
	case fieldPassword:
		a.password.Focus()
	case fieldToken:
		a.token.Focus()
	case fieldAPIKeyName:
		a.keyName.Focus()
	case fieldAPIKeyValue:
		a.keyValue.Focus()
	}
}

func (a *authPane) update(focused bool, msg tea.Msg) tea.Cmd {
	if !focused {
		return nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "tab":
			fields := append([]authField{fieldType}, a.fieldsForType()...)
			idx := 0
			for i, f := range fields {
				if f == a.active {
					idx = i
					break
				}
			}
			a.active = fields[(idx+1)%len(fields)]
			a.syncFocus()
			return nil
		case "left", "right":
			if a.active == fieldType {
				n := len(authTypes)
				if km.String() == "right" {
					a.typeIdx = (a.typeIdx + 1) % n
				} else {
					a.typeIdx = (a.typeIdx - 1 + n) % n
				}
				a.active = fieldType
				a.syncFocus()
				return nil
			}
			if a.active == fieldAPIKeyPlacement {
				a.inQuery = !a.inQuery
				return nil
			}
		case " ":
			if a.active == fieldAPIKeyPlacement {
				a.inQuery = !a.inQuery
				return nil
			}
		}
	}

	var cmd tea.Cmd
	switch a.active {
	case fieldUsername:
		a.username, cmd = a.username.Update(msg)
	case fieldPassword:
		a.password, cmd = a.password.Update(msg)
	case fieldToken:
		a.token, cmd = a.token.Update(msg)
	case fieldAPIKeyName:
		a.keyName, cmd = a.keyName.Update(msg)
	case fieldAPIKeyValue:
		a.keyValue, cmd = a.keyValue.Update(msg)
	}
	return cmd
}

func (a *authPane) handleClick(msg tea.MouseMsg) bool {
	if zone.Get(zoneAuthField(fieldType)).InBounds(msg) {
		a.active = fieldType
		a.syncFocus()
		return true
	}
	for _, f := range a.fieldsForType() {
		if zone.Get(zoneAuthField(f)).InBounds(msg) {
			a.active = f
			a.syncFocus()
			return true
		}
	}
	return false
}

func (a *authPane) view(width int, focused bool) string {
	label := authTypeLabels[a.authType()]
	pillStyle := tabInactive
	if focused && a.active == fieldType {
		pillStyle = tabActive
	}
	pill := zone.Mark(zoneAuthField(fieldType), pillStyle.Render("‹ "+label+" ›"))
	lines := []string{pill, ""}

	row := func(label string, ti textinput.Model, f authField) {
		ti.Width = width - len(label) - 6
		lines = append(lines, zone.Mark(zoneAuthField(f), label+": "+ti.View()))
	}

	switch a.authType() {
	case model.AuthBasic:
		row("Username", a.username, fieldUsername)
		row("Password", a.password, fieldPassword)
	case model.AuthBearer:
		row("Token   ", a.token, fieldToken)
	case model.AuthAPIKey:
		row("Key Name ", a.keyName, fieldAPIKeyName)
		row("Key Value", a.keyValue, fieldAPIKeyValue)
		placement := "Header"
		if a.inQuery {
			placement = "Query Param"
		}
		placementStyle := tabInactive
		if focused && a.active == fieldAPIKeyPlacement {
			placementStyle = tabActive
		}
		lines = append(lines, zone.Mark(zoneAuthField(fieldAPIKeyPlacement), "Send in  : "+placementStyle.Render("‹ "+placement+" ›")))
	default:
		lines = append(lines, dimStyle.Render("No authentication will be sent."))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return borderFor(focused).Width(width - 2).Render(body)
}
