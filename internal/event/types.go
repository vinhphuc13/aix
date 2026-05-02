package event

import "time"

type EventType string

const (
	EventSessionStarted EventType = "session_started"
	EventSessionResumed EventType = "session_resumed"
	EventSessionStopped EventType = "session_stopped"
	EventCheckpoint     EventType = "checkpoint"
	EventTaskAdded      EventType = "task_added"
	EventTaskDone       EventType = "task_done"
	EventFileEdited     EventType = "file_edited"
	EventFileAdded      EventType = "file_added"
	EventDecisionAdded  EventType = "decision_added"
	EventNoteAdded      EventType = "note_added"
	EventFocusChanged   EventType = "focus_changed"
)

type Event struct {
	ID        string            `json:"id"`
	SessionID string            `json:"session_id"`
	Type      EventType         `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]string `json:"data,omitempty"`
}
