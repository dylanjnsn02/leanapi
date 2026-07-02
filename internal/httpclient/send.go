package httpclient

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"leanapi/internal/history"
	"leanapi/internal/model"
)

// ResponseMsg is delivered to the Bubble Tea Update loop once a request
// finishes (successfully or not).
type ResponseMsg struct {
	Snapshot   *history.ResponseSnapshot
	Err        error
	DurationMS int64
}

// SendRequestCmd builds and executes req on Bubble Tea's own command
// goroutine, never blocking the UI thread.
func SendRequestCmd(req model.Request) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		httpReq, err := BuildHTTPRequest(req)
		if err != nil {
			return ResponseMsg{Err: err, DurationMS: time.Since(start).Milliseconds()}
		}

		resp, err := Client.Do(httpReq)
		duration := time.Since(start).Milliseconds()
		if err != nil {
			return ResponseMsg{Err: err, DurationMS: duration}
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ResponseMsg{Err: err, DurationMS: duration}
		}

		snap := &history.ResponseSnapshot{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    map[string][]string(resp.Header),
			Body:       string(bodyBytes),
			Size:       int64(len(bodyBytes)),
		}

		return ResponseMsg{Snapshot: snap, DurationMS: duration}
	}
}
