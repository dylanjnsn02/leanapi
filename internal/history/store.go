package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultPath returns the path to the history JSONL file, creating its
// parent directory if needed.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "leanapi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.jsonl"), nil
}

// Append writes one entry as a single JSON line to the history file.
func Append(e Entry) error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(e)
}

// LoadAll reads every entry from the history file, oldest first.
func LoadAll() ([]Entry, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip malformed lines rather than fail the whole browser
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
