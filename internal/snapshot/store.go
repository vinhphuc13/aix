package snapshot

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Meta struct {
	ID           string    `json:"id"`
	SessionID    string    `json:"session_id"`
	CheckpointID string    `json:"checkpoint_id"`
	CreatedAt    time.Time `json:"created_at"`
	Files        []File    `json:"files"`
}

type File struct {
	OriginalPath string `json:"original_path"`
	SnapshotPath string `json:"snapshot_path"`
	SizeBytes    int64  `json:"size_bytes"`
}

func newID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Create(aixDir, sessionID, checkpointID, projectRoot string, files []string) (*Meta, error) {
	id := newID()
	snapDir := filepath.Join(aixDir, "snapshots", id)
	if err := os.MkdirAll(snapDir, 0755); err != nil {
		return nil, err
	}

	meta := &Meta{
		ID:           id,
		SessionID:    sessionID,
		CheckpointID: checkpointID,
		CreatedAt:    time.Now(),
	}

	for _, relPath := range files {
		srcPath := filepath.Join(projectRoot, relPath)
		data, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}
		snapName := strings.ReplaceAll(relPath, string(filepath.Separator), "__")
		dstPath := filepath.Join(snapDir, snapName)
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			continue
		}
		meta.Files = append(meta.Files, File{
			OriginalPath: relPath,
			SnapshotPath: snapName,
			SizeBytes:    int64(len(data)),
		})
	}

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(snapDir, "meta.json"), metaData, 0644); err != nil {
		return nil, err
	}
	return meta, nil
}

func Load(aixDir, id string) (*Meta, error) {
	data, err := os.ReadFile(filepath.Join(aixDir, "snapshots", id, "meta.json"))
	if err != nil {
		return nil, err
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
