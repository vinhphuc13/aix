package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/vinhphuc13/aix/internal/mcpserver"
	mcpserver_pkg "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server for bidirectional context sync with Claude Code and Cursor",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the aix MCP server (stdio transport)",
	Long: `Start the aix MCP server over stdio.

Configure in Claude Code (.claude/settings.json):
  "mcpServers": {
    "aix": { "command": "aix", "args": ["mcp", "serve"] }
  }

Configure in Cursor (settings.json):
  "mcpServers": {
    "aix": { "command": "aix", "args": ["mcp", "serve"] }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		aixDir := mustFindAIXDir()
		s := mcpserver.New(aixDir)
		return mcpserver_pkg.ServeStdio(s)
	},
}

var mcpConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Print MCP config snippet for Claude Code and Cursor",
	RunE: func(cmd *cobra.Command, args []string) error {
		binary, err := os.Executable()
		if err != nil {
			binary = "aix"
		}
		// On Windows use the full path; on Unix just the name if on PATH
		if runtime.GOOS != "windows" {
			if base := filepath.Base(binary); base == "aix" {
				binary = "aix"
			}
		}

		cfg := map[string]any{
			"mcpServers": map[string]any{
				"aix": map[string]any{
					"command": binary,
					"args":    []string{"mcp", "serve"},
				},
			},
		}
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println("Add to .claude/settings.json or Cursor's MCP settings:")
		fmt.Println(string(data))
		fmt.Println()
		fmt.Println("Tools exposed:")
		fmt.Println("  aix_status        — get current session context")
		fmt.Println("  aix_add_task      — add a task")
		fmt.Println("  aix_done          — mark a task done")
		fmt.Println("  aix_add_decision  — record a decision")
		fmt.Println("  aix_add_note      — add a note")
		fmt.Println("  aix_checkpoint    — save a checkpoint")
		fmt.Println("  aix_focus         — set current focus")
		return nil
	},
}

func init() {
	mcpCmd.AddCommand(mcpServeCmd, mcpConfigCmd)
}
