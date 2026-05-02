package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add items to the current session (task, decision, note, file)",
}

// add task
var addTaskNote string

var addTaskCmd = &cobra.Command{
	Use:   "task <title>",
	Short: "Add a pending task",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		aixDir := mustFindAIXDir()
		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}
		t := session.Task{
			ID:        session.NewID(),
			Title:     title,
			Status:    session.TaskOpen,
			Note:      addTaskNote,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		s.Tasks = append(s.Tasks, t)
		if err := session.Save(aixDir, s); err != nil {
			return err
		}
		_ = event.Append(aixDir, s.ID, event.EventTaskAdded, map[string]string{
			"id": t.ID, "title": t.Title,
		})
		fmt.Printf("[ ] %s\n", t.Title)
		return nil
	},
}

// add decision
var addDecisionRationale string

var addDecisionCmd = &cobra.Command{
	Use:   "decision <summary>",
	Short: "Record a decision",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		summary := strings.Join(args, " ")
		aixDir := mustFindAIXDir()
		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}
		d := session.Decision{
			ID:        session.NewID(),
			Summary:   summary,
			Rationale: addDecisionRationale,
			CreatedAt: time.Now(),
		}
		s.Decisions = append(s.Decisions, d)
		if err := session.Save(aixDir, s); err != nil {
			return err
		}
		_ = event.Append(aixDir, s.ID, event.EventDecisionAdded, map[string]string{"summary": d.Summary})
		if d.Rationale != "" {
			fmt.Printf("• %s [%s]\n", d.Summary, d.Rationale)
		} else {
			fmt.Printf("• %s\n", d.Summary)
		}
		return nil
	},
}

// add note
var addNoteTag string

var addNoteCmd = &cobra.Command{
	Use:   "note <content>",
	Short: "Add an engineering note",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content := strings.Join(args, " ")
		aixDir := mustFindAIXDir()
		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}
		n := session.Note{
			ID:        session.NewID(),
			Content:   content,
			Tag:       addNoteTag,
			CreatedAt: time.Now(),
		}
		s.Notes = append(s.Notes, n)
		if err := session.Save(aixDir, s); err != nil {
			return err
		}
		_ = event.Append(aixDir, s.ID, event.EventNoteAdded, map[string]string{
			"content": n.Content, "tag": n.Tag,
		})
		tag := ""
		if n.Tag != "" {
			tag = "[" + n.Tag + "] "
		}
		fmt.Printf("Note: %s%s\n", tag, n.Content)
		return nil
	},
}

// add file
var addFileRole string

var addFileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Track a file in the active session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		aixDir := mustFindAIXDir()
		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}

		projectRoot := filepath.Dir(aixDir)
		if !filepath.IsAbs(path) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, path)
		}
		relPath := path
		if rel, err := filepath.Rel(projectRoot, path); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = rel
		}

		for _, f := range s.ActiveFiles {
			if f.Path == relPath {
				fmt.Printf("Already tracking: %s\n", relPath)
				return nil
			}
		}

		role := addFileRole
		if role == "" {
			role = "primary"
		}
		s.ActiveFiles = append(s.ActiveFiles, session.ActiveFile{
			Path:    relPath,
			Role:    role,
			AddedAt: time.Now(),
		})
		if err := session.Save(aixDir, s); err != nil {
			return err
		}
		_ = event.Append(aixDir, s.ID, event.EventFileAdded, map[string]string{
			"path": relPath, "role": role,
		})
		fmt.Printf("[%s] %s\n", role, relPath)
		return nil
	},
}

func init() {
	addTaskCmd.Flags().StringVar(&addTaskNote, "note", "", "additional note for the task")
	addDecisionCmd.Flags().StringVar(&addDecisionRationale, "rationale", "", "rationale for the decision")
	addNoteCmd.Flags().StringVar(&addNoteTag, "tag", "", "tag: arch, risk, todo, etc.")
	addFileCmd.Flags().StringVar(&addFileRole, "role", "primary", "role: primary, test, config, infra")

	addCmd.AddCommand(addTaskCmd, addDecisionCmd, addNoteCmd, addFileCmd)
}
