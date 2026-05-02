package cmd

import "github.com/vinhphuc13/aix/internal/session"

func filterTasks(tasks []session.Task, statuses ...session.TaskStatus) []session.Task {
	set := make(map[session.TaskStatus]bool)
	for _, s := range statuses {
		set[s] = true
	}
	var out []session.Task
	for _, t := range tasks {
		if set[t.Status] {
			out = append(out, t)
		}
	}
	return out
}

func taskIcon(status session.TaskStatus) string {
	switch status {
	case session.TaskOpen:
		return "[ ]"
	case session.TaskInProgress:
		return "[~]"
	case session.TaskDone:
		return "[x]"
	case session.TaskBlocked:
		return "[!]"
	default:
		return "[ ]"
	}
}

func eventDataSummary(data map[string]string) string {
	for _, key := range []string{"title", "path", "summary", "message", "content", "name", "to"} {
		if v, ok := data[key]; ok {
			if len(v) > 60 {
				v = v[:60] + "..."
			}
			return v
		}
	}
	return ""
}
