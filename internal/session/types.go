package session

import "time"

type Status string

const (
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
)

type TaskStatus string

const (
	TaskOpen       TaskStatus = "open"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskBlocked    TaskStatus = "blocked"
)

type Session struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Goal         string       `json:"goal"`
	Status       Status       `json:"status"`
	CurrentFocus string       `json:"current_focus,omitempty"`
	LastAgent    string       `json:"last_agent,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Tasks        []Task       `json:"tasks"`
	Decisions    []Decision   `json:"decisions"`
	Notes        []Note       `json:"notes"`
	ActiveFiles  []ActiveFile `json:"active_files"`
	Checkpoints  []Checkpoint `json:"checkpoints"`
	ArchNotes    string       `json:"arch_notes,omitempty"`
}

type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	Note      string     `json:"note,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Decision struct {
	ID        string    `json:"id"`
	Summary   string    `json:"summary"`
	Rationale string    `json:"rationale,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Note struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Tag       string    `json:"tag,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ActiveFile struct {
	Path       string    `json:"path"`
	Role       string    `json:"role"`
	AddedAt    time.Time `json:"added_at"`
	LastEditAt time.Time `json:"last_edit_at,omitempty"`
}

type Checkpoint struct {
	ID         string    `json:"id"`
	Message    string    `json:"message"`
	SnapshotID string    `json:"snapshot_id,omitempty"`
	OpenTasks  int       `json:"open_tasks"`
	DoneTasks  int       `json:"done_tasks"`
	CreatedAt  time.Time `json:"created_at"`
}
