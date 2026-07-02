package httpclient

import (
	"net/http"
	"net/url"
	"strings"

	"leanapi/internal/model"
)

// DerivedHeaders computes the header(s) implied by an auth config. It never
// mutates or reads Request.Headers -- callers apply these first so that a
// user-entered header with the same name can override them.
func DerivedHeaders(auth model.AuthConfig) []model.Header {
	switch auth.Type {
	case model.AuthBasic:
		return nil // net/http.Request.SetBasicAuth is applied directly in BuildHTTPRequest
	case model.AuthBearer:
		if auth.Token == "" {
			return nil
		}
		return []model.Header{{Key: "Authorization", Value: "Bearer " + auth.Token, Enabled: true}}
	case model.AuthAPIKey:
		if auth.APIKeyPlacement == model.APIKeyInHeader && auth.APIKeyName != "" {
			return []model.Header{{Key: auth.APIKeyName, Value: auth.APIKeyValue, Enabled: true}}
		}
		return nil
	default:
		return nil
	}
}

// BuildHTTPRequest turns a model.Request into a ready-to-send *http.Request,
// applying auth (as headers/query/basic-auth) before user headers so that a
// user header of the same name always wins (last write wins via Set).
func BuildHTTPRequest(r model.Request) (*http.Request, error) {
	rawURL := r.URL

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if r.Auth.Type == model.AuthAPIKey && r.Auth.APIKeyPlacement == model.APIKeyInQuery && r.Auth.APIKeyName != "" {
		q.Set(r.Auth.APIKeyName, r.Auth.APIKeyValue)
	}
	for _, p := range r.Params {
		if !p.Enabled || p.Key == "" {
			continue
		}
		q.Set(p.Key, p.Value)
	}
	u.RawQuery = q.Encode()
	rawURL = u.String()

	var body strings.Reader
	if r.Body != "" {
		body = *strings.NewReader(r.Body)
	}

	method := r.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, rawURL, &body)
	if err != nil {
		return nil, err
	}

	if r.Auth.Type == model.AuthBasic {
		req.SetBasicAuth(r.Auth.Username, r.Auth.Password)
	}

	for _, h := range DerivedHeaders(r.Auth) {
		req.Header.Set(h.Key, h.Value)
	}

	for _, c := range r.Cookies {
		if !c.Enabled || c.Key == "" {
			continue
		}
		req.AddCookie(&http.Cookie{Name: c.Key, Value: c.Value})
	}

	hasContentType := false
	for _, h := range r.Headers {
		if !h.Enabled || h.Key == "" {
			continue
		}
		req.Header.Set(h.Key, h.Value)
		if strings.EqualFold(h.Key, "Content-Type") {
			hasContentType = true
		}
	}

	if r.Body != "" && !hasContentType {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
