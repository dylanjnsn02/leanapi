package model

// AuthType identifies which authentication scheme a request uses.
type AuthType int

const (
	AuthNone AuthType = iota
	AuthBasic
	AuthBearer
	AuthAPIKey
)

// APIKeyPlacement controls where an API key auth value is injected.
type APIKeyPlacement int

const (
	APIKeyInHeader APIKeyPlacement = iota
	APIKeyInQuery
)

// Header is a single user-entered request header.
type Header struct {
	Key     string
	Value   string
	Enabled bool
}

// AuthConfig holds the fields for every supported auth type. Only the
// fields relevant to Type are used when building a request.
type AuthConfig struct {
	Type AuthType

	Username string
	Password string

	Token string

	APIKeyName      string
	APIKeyValue     string
	APIKeyPlacement APIKeyPlacement
}

var Methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

// Request is the full description of an HTTP call a user builds in the UI.
// Auth-derived headers are intentionally not part of Headers: they are
// computed at send time so they can never silently collide with or be
// duplicated alongside user-entered headers.
type Request struct {
	Method  string
	URL     string
	Params  []Header // query params merged onto the URL at send time
	Headers []Header
	Cookies []Header // name/value pairs sent as a single Cookie header at send time
	Auth    AuthConfig
	Body    string
}

// NewRequest returns a Request with sane defaults for a fresh session.
func NewRequest() Request {
	return Request{
		Method: "GET",
	}
}
