package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NewID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func InitDir(aixDir string) error {
	for _, sub := range []string{"sessions", "events", "snapshots"} {
		if err := os.MkdirAll(filepath.Join(aixDir, sub), 0755); err != nil {
			return err
		}
	}
	return nil
}

func NewSession(aixDir, name, goal string) (*Session, error) {
	s := &Session{
		ID:          NewID(),
		Name:        name,
		Goal:        goal,
		Status:      StatusActive,
		LastAgent:   "claude",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tasks:       []Task{},
		Decisions:   []Decision{},
		Notes:       []Note{},
		ActiveFiles: []ActiveFile{},
		Checkpoints: []Checkpoint{},
	}
	if err := Save(aixDir, s); err != nil {
		return nil, err
	}
	if err := SetCurrent(aixDir, s.ID); err != nil {
		return nil, err
	}
	return s, nil
}

func Save(aixDir string, s *Session) error {
	s.UpdatedAt = time.Now()
	dir := filepath.Join(aixDir, "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func Load(aixDir, id string) (*Session, error) {
	path := filepath.Join(aixDir, "sessions", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("session %s not found", id)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("corrupt session %s: %w", id, err)
	}
	return &s, nil
}

func SetCurrent(aixDir, id string) error {
	path := filepath.Join(aixDir, "current")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(id), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func GetCurrent(aixDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(aixDir, "current"))
	if err != nil {
		return "", fmt.Errorf("no active session (run 'aix start <name>')")
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", fmt.Errorf("no active session")
	}
	return id, nil
}

func LoadCurrent(aixDir string) (*Session, error) {
	id, err := GetCurrent(aixDir)
	if err != nil {
		return nil, err
	}
	return Load(aixDir, id)
}

func List(aixDir string) ([]*Session, error) {
	dir := filepath.Join(aixDir, "sessions")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		s, err := Load(aixDir, id)
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
