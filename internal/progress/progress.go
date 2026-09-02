// Package progress tracks reading position in $XDG_STATE_HOME/pgbook.
package progress

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/pgrundev/pgbook/internal/topic"
)

type state struct {
	LastRead string `json:"last_read"`
}

func stateFile() (string, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "pgbook", "progress.json"), nil
}

// SaveLastRead records the slug of the topic just read.
func SaveLastRead(slug string) error {
	path, err := stateFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(state{LastRead: slug})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LastRead returns the saved slug, or "" when there is no progress yet.
func LastRead() (string, error) {
	path, err := stateFile()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return "", nil // corrupt state: start over rather than fail
	}
	return s.LastRead, nil
}

// Next returns the topic after lastRead in book order. done is true when
// lastRead was the final topic.
func Next(ts []topic.Topic, lastRead string) (next topic.Topic, done bool) {
	ordered := make([]topic.Topic, len(ts))
	copy(ordered, ts)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Order < ordered[j].Order })
	if len(ordered) == 0 {
		return topic.Topic{}, true
	}
	if lastRead == "" {
		return ordered[0], false
	}
	for i, t := range ordered {
		if t.Slug == lastRead {
			if i == len(ordered)-1 {
				return topic.Topic{}, true
			}
			return ordered[i+1], false
		}
	}
	return ordered[0], false
}
