package inject

import (
	"os"
	"path/filepath"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
)

// WriteContextFile writes the rendered context block to <aixDir>/context.md.
func WriteContextFile(aixDir string, s *session.Session, events []event.Event) error {
	content := RenderContext(s, events)
	path := filepath.Join(aixDir, "context.md")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
