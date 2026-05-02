# aix

Local-first AI coding session runtime. Persists engineering working state — goals, tasks, decisions, files — across AI coding sessions.

Think: **"Git for AI sessions"**. Not chat history.

## Why

AI coding tools (Claude Code, Cursor, etc.) start every session cold. `aix` gives each session memory: what you're building, what you've decided, what's left to do. When you resume, the context is injected automatically.

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

## Quick Start

```bash
# 1. Start a session in your project directory
aix start auth-refactor --goal "Replace JWT with session tokens"

# 2. Install Claude Code hooks (auto-injects context into every prompt)
aix hook install

# 3. Work — add tasks, record decisions, track files
aix add task "Update middleware to validate session tokens"
aix add decision "Use Redis for session storage" --rationale "need TTL support"
aix add file internal/auth/middleware.go --role primary

# 4. Mark tasks done as you go
aix done "Update middleware"

# 5. Save a checkpoint before stopping
aix checkpoint -m "middleware done, storage next"

# 6. Next day: resume
aix continue
```

## Commands

### Session lifecycle

| Command | Description |
|---|---|
| `aix start <name> [--goal <text>]` | Start a new session |
| `aix continue [session-id]` | Resume a session and print its context |
| `aix status` | Show current session state |
| `aix status --json` | Raw JSON output |
| `aix list` | List all sessions |

### Tracking

| Command | Description |
|---|---|
| `aix add task <title> [--note <text>]` | Add a task |
| `aix add decision <summary> [--rationale <text>]` | Record a decision |
| `aix add note <content> [--tag arch\|risk\|todo]` | Add an engineering note |
| `aix add file <path> [--role primary\|test\|config\|infra]` | Track a file |
| `aix done <task-title-or-id>` | Mark a task done |
| `aix focus <text>` | Set current focus |

### Checkpoints

```bash
aix checkpoint -m "what you just did"
aix checkpoint -m "before big refactor" --snapshot   # also copies active files
```

### Claude Code hooks

```bash
aix hook install            # project-level (.claude/settings.json)
aix hook install --global   # user-level (~/.claude/settings.json)
aix hook uninstall
```

Once installed, hooks:
- Inject session context into every prompt (`UserPromptSubmit`)
- Track file edits automatically (`PostToolUse`)
- Auto-checkpoint when a session ends (`Stop`)

For Cursor or other tools, use `aix continue --format cursor` and paste the output manually.

## How it works

`aix` stores session state in a `.aix/` directory at your project root — plain JSON files, no server, no sync.

```
.aix/
  current          # pointer to active session ID
  sessions/
    <id>.json      # session state (tasks, decisions, notes, files, checkpoints)
  events/
    <session-id>.jsonl  # append-only event log
  snapshots/       # optional file snapshots (--snapshot flag)
```

The `aix continue` command (and hooks) render this state into a context block that AI tools can read at the start of each session.

## Development

```bash
make build    # build ./aix binary
make install  # go install
make test     # run tests
make smoke    # install + end-to-end smoke test
```
