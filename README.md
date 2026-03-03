# CCSS — Claude Code Session Manager

A terminal UI for browsing, searching, and analyzing [Claude Code](https://docs.anthropic.com/en/docs/claude-code) session data.

CCSS reads the session files stored in `~/.claude/` and provides:

- **Session browser** — sessions grouped by project, sorted by time
- **Conversation viewer** — full message history with collapsible tool/thinking blocks
- **Full-text search** — search across all sessions with concurrent JSONL scanning
- **Cost dashboard** — token usage, model breakdown, estimated costs, activity charts
- **Markdown export** — export any session as a `.md` file

## Install

```bash
go install github.com/ccss@latest
```

Or build from source:

```bash
git clone https://github.com/anthropics/ccss.git
cd ccss
go build -o ccss .
```

## Usage

```bash
ccss
```

That's it. CCSS automatically finds and loads session data from `~/.claude/`.

## Keybindings

| Key | Action |
|-----|--------|
| `1` `2` `3` | Switch tabs (Sessions / Search / Stats) |
| `j` `k` / `↑` `↓` | Navigate / scroll |
| `enter` | Open session / execute search |
| `/` | Filter sessions / focus search |
| `tab` | Toggle tool/thinking block |
| `e` | Export current session |
| `?` | Help |
| `esc` | Back |
| `q` | Quit |

## Screenshots

```
CCSS  1 Sessions  2 Search  3 Stats
──────────────────────────────────────────────
D:\claudecode
  ▸ 15:34  "Fix authentication bug in..."   42 msgs  main
    12:01  "Add export feature to..."         8 msgs  feat/export
D:\other-project
    09:22  "Setup CI pipeline..."            15 msgs  develop
```

## How it works

CCSS reads three data sources from `~/.claude/`:

- `projects/*/sessions-index.json` — session metadata (project path, timestamps, prompts)
- `projects/*/sessions/*.jsonl` — full conversation logs (streamed with 16MB line buffer)
- `statsCache/stats-cache.json` — aggregated usage statistics

Search uses a two-phase approach: first scanning session metadata in memory, then concurrently scanning JSONL files (4 goroutines, `bytes.Contains` pre-filter to skip non-matching lines before JSON parsing).

Cost estimation uses hardcoded per-model pricing:

| Model | Input | Output | Cache Read | Cache Write |
|-------|-------|--------|------------|-------------|
| claude-opus-4 | $5/M | $25/M | $0.50/M | $6.25/M |
| claude-sonnet-4 | $3/M | $15/M | $0.30/M | $3.75/M |
| claude-haiku-4 | $1/M | $5/M | $0.10/M | $1.25/M |

## Requirements

- Go 1.22+
- A terminal with Unicode support
- Claude Code session data in `~/.claude/`

## License

MIT
