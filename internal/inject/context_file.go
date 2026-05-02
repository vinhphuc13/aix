package inject

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
)

// WriteContextFile writes the rendered context block to <aixDir>/context.md.
// If <projectRoot>/.cursorrules already contains an aix block (from a prior
// aix continue --format cursor), it is updated in place automatically.
func WriteContextFile(aixDir string, s *session.Session, events []event.Event) error {
	content := RenderContext(s, events)
	path := filepath.Join(aixDir, "context.md")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	// Auto-sync .cursorrules if the user previously opted in via
	// aix continue --format cursor (sentinel block already present).
	projectRoot := filepath.Dir(aixDir)
	cursorRules := filepath.Join(projectRoot, ".cursorrules")
	existing, err := os.ReadFile(cursorRules)
	if err == nil && strings.Contains(string(existing), "=== AIX SESSION CONTEXT ===") {
		_ = UpsertCursorRules(projectRoot, s, events)
	}
	return nil
}

// UpsertCursorRules inserts or replaces the aix context block in
// <projectRoot>/.cursorrules. Content outside the block is preserved.
func UpsertCursorRules(projectRoot string, s *session.Session, events []event.Event) error {
	block := RenderContext(s, events)
	path := filepath.Join(projectRoot, ".cursorrules")
	existing, _ := os.ReadFile(path)
	updated := upsertBlock(string(existing), block)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(updated), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func upsertBlock(existing, block string) string {
	const start = "=== AIX SESSION CONTEXT ==="
	const end = "=== END AIX CONTEXT ==="
	si := strings.Index(existing, start)
	ei := strings.Index(existing, end)
	if si >= 0 && ei > si {
		tail := ei + len(end)
		if tail < len(existing) && existing[tail] == '\n' {
			tail++
		}
		return existing[:si] + block + existing[tail:]
	}
	if len(existing) == 0 {
		return block
	}
	if !strings.HasSuffix(existing, "\n") {
		return existing + "\n\n" + block
	}
	return existing + "\n" + block
}
