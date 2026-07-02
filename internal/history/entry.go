package history

import (
	"time"

	"leanapi/internal/model"
)

// ResponseSnapshot is the persisted (and in-memory) shape of an HTTP
// response. Body is stored raw; pretty-printing is a view-time transform.
type ResponseSnapshot struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       string
	Size       int64
}

// Entry is one logged request/response pair. Response is nil when the
// request errored before a response was received.
type Entry struct {
	ID         string
	Timestamp  time.Time
	Request    model.Request
	Response   *ResponseSnapshot
	Error      string
	DurationMS int64
}

// NewID returns a sortable, unique-enough-for-single-process id.
func NewID() string {
	return time.Now().UTC().Format("20060102T150405.000000000Z")
}
