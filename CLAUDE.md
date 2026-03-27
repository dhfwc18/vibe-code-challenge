# Claude Code Instructions

## Interaction Logging

Every interaction must be logged to `vibe_code_log/`. Rules:

- Each session gets its own log file named `session_YYYY-MM-DD_NNN.md` (NNN = 001, 002, etc. for multiple sessions per day)
- Each log entry records:
  - **User instruction:** exactly what the user said
  - **Response:** a summary of any verbal (non-code) response given
- Log the interaction at the end of each response, before finishing
- Do not log tool call details, only the human-readable exchange

## Git Commit Convention

Format: `<type>(<scope>): <short summary>`
Optional body: explains *why*, not *what*.

**Types:** `feat` | `fix` | `refactor` | `test` | `docs` | `chore` | `style`

**Rules:**
- One logical change per commit
- Summary line ≤ 72 characters, imperative mood ("add", not "added")
- Scope = area changed (e.g. `auth`, `api`, `ui`); omit if global
- Body only when the why isn't obvious from the diff
- Never use `--no-verify`
- Never amend published commits — create a new one instead
- Never force-push to main/master

**Examples:**
```
feat(auth): add JWT refresh token rotation
fix(api): return 404 when resource not found
refactor(ui): extract Button into shared component
chore: upgrade eslint to v9
```

**Branching strategy:**
- `main` is the stable branch — all merged code must build and pass lint
- Work happens on short-lived branches named `<type>/<short-description>` (e.g. `feat/player-movement`, `fix/collision-detection`)
- Branches are deleted after merging
- Never commit directly to `main`

## Repository Layout

```
vibe-code-challenge/        <- parent repo (logs, docs, config)
    app/                    <- Go project root (go.mod lives here)
        cmd/app/main.go
        internal/
```

All Go code lives under `app/`. The parent directory holds session logs, CLAUDE.md, and README only.

## No Non-ASCII Characters

Never use non-ASCII characters anywhere in this repo — no emoji, no Unicode symbols, no curly quotes. This applies to:
- Source code (comments, strings, identifiers)
- Config files
- Commit messages
- Log files
- Any file Claude writes or edits

If any non-ASCII characters are found in existing files, remove them immediately.
