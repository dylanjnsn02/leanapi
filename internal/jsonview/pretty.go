package jsonview

import (
	"bytes"
	"encoding/json"
)

// Pretty indents raw JSON bytes without unmarshaling through a map (which
// would lose key order). Returns the original string and false if raw is
// not valid JSON.
func Pretty(raw []byte) (string, bool) {
	if !json.Valid(raw) {
		return string(raw), false
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw), false
	}
	return buf.String(), true
}
