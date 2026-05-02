# aix

Local-first AI coding session runtime. Persists engineering working state — goals, tasks, decisions, files — across AI coding sessions.

Think: **"Git for AI sessions"**. Not chat history.

## Why

You hit Claude Pro's limit mid-feature. You switch to Cursor. The context is gone — Cursor doesn't know what you were building, what you decided, what's left to do. You spend the first few minutes re-explaining everything.

`aix` solves this. Both Claude Code and Cursor connect to the same MCP server. Every AI tool reads from and writes to the same session. Switch mid-task — context is continuous, zero manual steps.

## Install

```bash
go install github.com/vinhphuc13/aix@latest
```

Or build from source:

```bash
git clone https://github.com/vinhphuc13/aix
cd aix
make install
```

Requires Go 1.23+.

---

## Setup (one time per project)

**1. Start a session**

```bash
cd your-project
aix start auth-refactor --goal "Replace JWT with session tokens"
```

Claude Code hooks are installed automatically. Session is ready.

**2. Connect both tools via MCP**

```bash
aix mcp config
```

This prints a JSON snippet. Paste it into:
- `.claude/settings.json` — Claude Code
- `Cursor Settings → MCP` — Cursor

Both tools now share the same session over MCP.

---

## Usage

### Work

Both Claude Code and Cursor call `aix_status` to read the current session and `aix_add_task`, `aix_done`, etc. to update it. You can also run these from the terminal at any time:

```bash
aix add task "Write integration tests for token refresh"
aix add decision "Use Redis for session storage" --rationale "need TTL support"
aix done "Write integration tests"
aix checkpoint -m "middleware done, redis storage next"
aix focus "token refresh endpoint"
```

### Switch between Claude Code and Cursor

Just switch. Both tools connect to the same MCP server and see the same session state. No resume command, no copy-paste, no re-explaining.

### Check where you are

```bash
aix status       # tasks, decisions, last checkpoint
aix continue     # print full context block
aix list         # list all sessions
```

---

## How it works

```
  Claude Code  ◀──── MCP (aix mcp serve) ────▶  Cursor
                              │
                         .aix/  (local)
                    sessions/<id>.json
                    context.md
                    events/<id>.jsonl
```

Both tools connect to `aix mcp serve` (stdio MCP server). Every tool call — `aix_done`, `aix_add_task`, etc. — updates the shared `.aix/` state and rewrites `.aix/context.md`. Claude Code also gets context injected per-prompt via its hook.

---

## Commands

### Session

| Command | Description |
|---|---|
| `aix start <name> [--goal <text>]` | Start a session (auto-installs hooks, auto-checkpoints) |
| `aix continue [session-id]` | Print current context block |
| `aix status [--json]` | Show tasks, decisions, last checkpoint |
| `aix list` | List all sessions |

### Tracking

| Command | Description |
|---|---|
| `aix add task <title> [--note <text>]` | Add a task |
| `aix add decision <summary> [--rationale <text>]` | Record a decision |
| `aix add note <content> [--tag arch\|risk\|todo]` | Add a note |
| `aix add file <path> [--role primary\|test\|config\|infra]` | Track a file |
| `aix done <task-title-or-id>` | Mark a task done |
| `aix focus <text>` | Set current focus |
| `aix checkpoint -m <message> [--snapshot]` | Save a checkpoint |

### MCP server

```bash
aix mcp serve    # start MCP server over stdio (called by AI tools, not humans)
aix mcp config   # print config snippet for Claude Code and Cursor
```

MCP tools available to AI agents:

| Tool | Description |
|---|---|
| `aix_status` | Get current session context |
| `aix_add_task` | Add a task |
| `aix_done` | Mark a task done |
| `aix_add_decision` | Record a decision |
| `aix_add_note` | Add a note |
| `aix_checkpoint` | Save a checkpoint |
| `aix_focus` | Set current focus |

### Claude Code hooks

Installed automatically by `aix start`. Managed manually with:

```bash
aix hook install [--global]
aix hook uninstall
```

Hooks: `UserPromptSubmit` (inject context), `PostToolUse` (track file edits), `Stop` (auto-checkpoint).

---

## File layout

```
.aix/
  current              # active session ID
  context.md           # always-current context block
  sessions/<id>.json   # full session state
  events/<id>.jsonl    # append-only event log
  snapshots/           # optional file snapshots (--snapshot flag)
```

---

## Development

```bash
make build    # build ./aix binary
make install  # go install
make test     # run tests
make smoke    # install + end-to-end smoke test
```
