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

## Testing Convention

Every exported function must have at least one test. Test files live alongside the code they test (foo.go -> foo_test.go).

**Naming pattern:** `Test<Operation>_<Condition>_<ExpectedOutcome>`

- Single obvious case: `TestAdd`
- Specific scenario: `TestMove_OutOfBounds_ReturnsError`
- Table-driven: top-level function is `TestMove`, each row passed to `t.Run("condition returns outcome", ...)`

**Rules:**
- Prefer table-driven tests for 3+ input variations
- Use `github.com/stretchr/testify/assert` for assertions
- Each test is fully self-contained — no shared state between tests
- Test public behaviour only — do not reach into unexported internals
- Benchmarks named `Benchmark<Operation>` when performance matters

## No Non-ASCII Characters

Never use non-ASCII characters anywhere in this repo — no emoji, no Unicode symbols, no curly quotes. This applies to:
- Source code (comments, strings, identifiers)
- Config files
- Commit messages
- Log files
- Any file Claude writes or edits

If any non-ASCII characters are found in existing files, remove them immediately.

## Project Context

**Game:** Net Zero - a government simulation game
**Setting:** A UK-like country, present day politics and policy environment
**Player role:** Civil servant pushing for net zero initiative
**Genre:** Map-based strategy with resource management as the central mechanic
**Timeline:** In-game clock runs from 2010 to 2050; player must reach net zero before the deadline
**Platform:** Desktop (Windows primary), rendered via Ebitengine v2 (Apache 2.0)

**Engine stack (all permissive licences, no copyleft):**
| Role | Library | Licence |
|---|---|---|
| Rendering + game loop | github.com/hajimehoshi/ebiten/v2 | Apache 2.0 |
| HUD / UI widgets | github.com/ebitenui/ebitenui | MIT |
| Tile map loading | github.com/lafriks/go-tiled | MIT |
| Camera pan/zoom | github.com/mazznoer/kamera | MIT |
| Entity management | github.com/yohamta/donburi | MIT |
| Assertions in tests | github.com/stretchr/testify | MIT |

**Numerical basis:** All carbon values, costs, and targets are sourced from the UK Green Book
and DESNZ supplementary guidance. Full reference data in docs/green_book_reference.md.
Key anchors: 2010 start = 590 MtCO2e/yr; SPC central 2030 = GBP 280/tCO2e; 2050 target = 0 net.

**Core design pillars:**
- Resource management: budget, political capital, public opinion, carbon output
- Map interaction: regions of the country with different energy profiles, industries, and electorates
- Policy decisions: draft, lobby, and pass legislation with trade-offs
- Time pressure: each in-game year advances the clock; events occur that the player must react to

**Package structure target:**
```
app/
    cmd/app/        <- entry point
    internal/
        game/       <- game loop, state machine
        world/      <- map, regions, tiles
        policy/     <- policy drafting and effects
        economy/    <- budget, resources, carbon model
        ui/         <- rendering, HUD, menus
        clock/      <- in-game time progression
        event/      <- random and scripted events
```
