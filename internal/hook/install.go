package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

type hookGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []hookEntry `json:"hooks"`
}

var aixHooks = []struct {
	eventType string
	matcher   string
	command   string
}{
	{"UserPromptSubmit", "", "aix hook prompt"},
	{"PostToolUse", "Edit|Write|MultiEdit", "aix hook posttooluse"},
	{"Stop", "", "aix hook stop"},
}

func Install(settingsPath string) error {
	settings := loadSettings(settingsPath)

	var hooksMap map[string][]json.RawMessage
	if raw, ok := settings["hooks"]; ok {
		_ = json.Unmarshal(raw, &hooksMap)
	}
	if hooksMap == nil {
		hooksMap = make(map[string][]json.RawMessage)
	}

	for _, h := range aixHooks {
		if isInstalled(hooksMap[h.eventType], h.command) {
			continue
		}
		g := hookGroup{
			Matcher: h.matcher,
			Hooks:   []hookEntry{{Type: "command", Command: h.command, Timeout: 10}},
		}
		data, _ := json.Marshal(g)
		hooksMap[h.eventType] = append(hooksMap[h.eventType], json.RawMessage(data))
	}

	hooksData, err := json.Marshal(hooksMap)
	if err != nil {
		return err
	}
	settings["hooks"] = json.RawMessage(hooksData)

	return writeSettings(settingsPath, settings)
}

func Uninstall(settingsPath string) error {
	settings := loadSettings(settingsPath)
	if settings["hooks"] == nil {
		return nil
	}

	var hooksMap map[string][]json.RawMessage
	if err := json.Unmarshal(settings["hooks"], &hooksMap); err != nil {
		return nil
	}

	remove := map[string]bool{
		"aix hook prompt":      true,
		"aix hook posttooluse": true,
		"aix hook stop":        true,
	}

	for eventType, groups := range hooksMap {
		var kept []json.RawMessage
		for _, raw := range groups {
			var g hookGroup
			if err := json.Unmarshal(raw, &g); err != nil {
				kept = append(kept, raw)
				continue
			}
			var newHooks []hookEntry
			for _, h := range g.Hooks {
				if !remove[h.Command] {
					newHooks = append(newHooks, h)
				}
			}
			if len(newHooks) > 0 {
				g.Hooks = newHooks
				data, _ := json.Marshal(g)
				kept = append(kept, json.RawMessage(data))
			}
		}
		if len(kept) > 0 {
			hooksMap[eventType] = kept
		} else {
			delete(hooksMap, eventType)
		}
	}

	hooksData, err := json.Marshal(hooksMap)
	if err != nil {
		return err
	}
	settings["hooks"] = json.RawMessage(hooksData)
	return writeSettings(settingsPath, settings)
}

func FindSettingsPath(cwd string, global bool) string {
	if global {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude", "settings.json")
	}
	return filepath.Join(cwd, ".claude", "settings.json")
}

func PrintWarningIfNotOnPath() {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(dir, "aix")); err == nil {
			return
		}
	}
	fmt.Fprintln(os.Stderr, "warning: 'aix' not found on PATH — hooks won't work until it is (try: go install .)")
}

func loadSettings(path string) map[string]json.RawMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]json.RawMessage)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return make(map[string]json.RawMessage)
	}
	return m
}

func writeSettings(path string, settings map[string]json.RawMessage) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func isInstalled(groups []json.RawMessage, command string) bool {
	for _, raw := range groups {
		var g hookGroup
		if err := json.Unmarshal(raw, &g); err != nil {
			continue
		}
		for _, h := range g.Hooks {
			if h.Command == command {
				return true
			}
		}
	}
	return false
}
