# aix

Local-first AI coding session runtime. Persists engineering working state — goals, tasks, decisions, files — across AI coding sessions.

Think: **"Git for AI sessions"**. Not chat history.

## Why

You hit Claude Pro's limit mid-feature. You switch to Cursor. The context is gone — Cursor doesn't know what you were building, what you decided, what's left to do. You spend the first few minutes re-explaining everything.

`aix` solves this. It keeps a shared `.aix/` state file in your project. Every AI tool reads from and writes to the same session. Switch between Claude Code and Cursor mid-task — context is continuous, no manual copy-paste.

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

## Usage

### 1. Start a session

```bash
cd your-project
aix start auth-refactor --goal "Replace JWT with session tokens"
```

This creates `.aix/` in your project and auto-saves an initial checkpoint. You're ready.

### 2. Connect your AI tools

**Claude Code** — hooks are installed automatically by `aix start`. Nothing extra to do.

**Cursor + Claude Code (bidirectional)** — set up the MCP server so both tools share the same session state:

```bash
aix mcp config   # prints the JSON snippet to paste into settings
```

Paste the output into `.claude/settings.json` (Claude Code) and Cursor's MCP settings (`Cursor Settings → MCP`). Both tools now read and write the same session.

### 3. Work normally

aix tracks your session automatically. You can also update it manually from the terminal at any time:

```bash
aix add task "Write integration tests for token refresh"
aix add decision "Use Redis for session storage" --rationale "need TTL support"
aix add file internal/auth/middleware.go --role primary
aix focus "token refresh flow"
```

### 4. Mark progress

```bash
aix done "Write integration tests"
aix checkpoint -m "middleware done, redis storage next"
```

### 5. Switch tools, resume context

**Switch from Claude Code to Cursor:**

If you set up MCP (step 2), Cursor already has the current context — just open it and keep working.

If you didn't set up MCP, run once:
```bash
aix continue --format cursor
```
This writes your session context into `.cursorrules`. Cursor picks it up automatically on the next prompt.

**Switch from Cursor back to Claude Code:**

If you used MCP, Claude Code already sees the updated state.

If not, just open Claude Code — the hook injects the latest context from `.aix/context.md` automatically.

### 6. Next day

```bash
aix continue     # prints current context; confirms you're resuming the right session
aix status       # see tasks, decisions, last checkpoint at a glance
```

---

## How context flows

```
┌─────────────────────────────────────────────────────────┐
│  Claude Code                         Cursor              │
│                                                          │
│  hook: auto-inject          MCP: aix_status             │
│  context per prompt    ◀──▶  aix_add_task, aix_done...  │
│                                                          │
│              .aix/ (shared state)                        │
│         sessions/<id>.json   context.md                  │
└─────────────────────────────────────────────────────────┘
```

- `.aix/context.md` — always kept up-to-date after every change. Claude Code hooks inject it automatically; Cursor reads it via MCP or `.cursorrules`.
- MCP server — both tools can call `aix_add_task`, `aix_done`, etc. directly. Changes from either side are immediately visible to the other.

---

## Commands

### Session

| Command | Description |
|---|---|
| `aix start <name> [--goal <text>]` | Start a new session (auto-creates initial checkpoint) |
| `aix continue [session-id] [--format cursor]` | Resume session; `--format cursor` writes to `.cursorrules` |
| `aix status [--json]` | Show current session state |
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

### Claude Code hooks

Hooks are installed automatically when you run `aix start`. To manage them manually:

```bash
aix hook install            # project-level (.claude/settings.json)
aix hook install --global   # user-level (~/.claude/settings.json)
aix hook uninstall
```

Hooks handle three events:
- `UserPromptSubmit` — injects session context into every prompt
- `PostToolUse` — tracks file edits automatically
- `Stop` — auto-checkpoints when the session ends

### MCP server

```bash
aix mcp serve    # start MCP server over stdio (called by AI tools)
aix mcp config   # print config snippet for Claude Code and Cursor
```

MCP tools exposed to AI agents:

| Tool | Description |
|---|---|
| `aix_status` | Get current session context |
| `aix_add_task` | Add a task |
| `aix_done` | Mark a task done |
| `aix_add_decision` | Record a decision |
| `aix_add_note` | Add a note |
| `aix_checkpoint` | Save a checkpoint |
| `aix_focus` | Set current focus |

---

## File layout

```
.aix/
  current              # active session ID
  context.md           # always-current context block (updated on every change)
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
