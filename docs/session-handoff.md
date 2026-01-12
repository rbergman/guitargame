# Session Handoff: Guitar Game Web UI

## Working Context

**Repository**: `/Users/bob/Projects/github/guitargame`
**Branch**: `main` (no worktree - working directly on main)
**Status**: Planning complete, ready to begin implementation

## Current Epic Status

**Epic**: `guitargame-svk` - Repository Setup & TypeScript Migration (P0)

This epic converts the repo from a single Go desktop app to a monorepo with:
- `apps/desktop/` - Original Go Gioui app (preserved as reference)
- `apps/web/` - New TypeScript web app (primary development focus)
- `songs/` - Shared song data (YAML source files)

| Order | Task ID | Title | Status |
|-------|---------|-------|--------|
| 1 | guitargame-aop | Create monorepo directory structure | ready |
| 2 | guitargame-yh8 | Move Go code to apps/desktop/ | blocked by aop |
| 3 | guitargame-pij | Update Go module imports | blocked by yh8 |
| 4 | guitargame-ive | Create .mise.toml with node@22 and just | blocked by aop |
| 5 | guitargame-2cc | Create root justfile (monorepo router) | blocked by ive |
| 6 | guitargame-bo4 | Initialize Vite TypeScript project in apps/web/ | blocked by aop, ive |
| 7 | guitargame-myv | Configure strict tsconfig.json | blocked by bo4 |
| 8 | guitargame-51l | Configure ESLint strict-type-checked | blocked by myv |
| 9 | guitargame-wu9 | Create apps/web/justfile module | blocked by 2cc, bo4 |
| 10 | guitargame-l6l | Set up Tailwind 4 with @theme design system | blocked by bo4 |
| 11 | guitargame-4no | Create core TypeScript type definitions | blocked by myv |
| 12 | guitargame-8tu | Set up husky pre-commit hooks | blocked by 51l |

**Progress**: 0/12 subtasks complete

## Design Documents

| Document | Path | Purpose |
|----------|------|---------|
| Web UI Spec | `docs/web-ui-spec.md` | Full spec for bass guitar practice game web UI |
| Migration Plan | `docs/typescript-migration-plan.md` | Monorepo structure, TypeScript setup, tooling |

## Git State

```
Branch: main (up to date with origin/main)
Untracked files: .beads/, AGENTS.md, CLAUDE.md, docs/
```

Recent commits:
- `39c237d` bd sync: 2026-01-12 13:50:44
- `4609f44` bd sync: 2026-01-12 13:36:38
- `0fe2e28` Initial version

## Ready Tasks

Only these can be started (no blockers):
```
bd ready
1. guitargame-svk [epic] - Repository Setup & TypeScript Migration
2. guitargame-aop [task] - Create monorepo directory structure
```

Start with `guitargame-aop` - it unblocks everything else.

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | Go desktop app entry (will move to apps/desktop/) |
| `internal/` | Go modules: audio, game, render, song |
| `songs/` | 10 YAML exercise files (stays at root) |
| `Makefile` | Go build (will move to apps/desktop/) |
| `docs/web-ui-spec.md` | Complete web UI specification |
| `docs/typescript-migration-plan.md` | Migration plan with code examples |

## Commands

### Beads workflow
```bash
bd prime                    # Get workflow context (run after compaction)
bd ready                    # Show available work
bd update <id> --status in_progress  # Claim task
bd close <id>               # Complete task
bd sync                     # Sync with git (run at session end)
```

### Quality gates (once web app exists)
```bash
just check                  # Root: all quality gates
just web check              # Web only: typecheck + lint + test
just web dev                # Start dev server
```

## Additional Task: SRT Recipe

Add a justfile recipe for running Claude Code with SRT (sandbox runtime) configured for network access:
- Context7 access (documentation lookup)
- Brightdata access (web scraping)

Example pattern:
```just
# Run Claude Code with network-enabled sandbox
claude-srt:
    srt --allow-net=context7.io,brightdata.com -- claude
```

## Next Steps

1. **Run `bd prime`** to restore beads workflow context
2. **Claim task**: `bd update guitargame-aop --status in_progress`
3. **Create monorepo structure**:
   ```bash
   mkdir -p apps/desktop apps/web
   ```
4. **Move Go code**: `git mv main.go go.mod go.sum Makefile .goreleaser.yml internal apps/desktop/`
5. **Create `.mise.toml`**: `mise use node@22 just`
6. **Create root `justfile`** with monorepo router pattern
7. **Add SRT recipe** to justfile for Claude Code with network access
8. **Initialize Vite TS project** in apps/web/
9. **Configure strict TypeScript and ESLint** per typescript-pro skill
10. **Close tasks** as completed, sync beads

## Skills to Use

- **dm-lang:typescript-pro** - Strict TypeScript setup, zero-any tolerance
- **dm-lang:just-pro** - Monorepo justfile patterns
- **dm-work:mise** - Tool version management
- **frontend-design:frontend-design** - Distinctive UI (when building components)
- **dm-game:game-design** - 5-component evaluation (when implementing gameplay)

## Project Context

This is a **Bass Guitar Practice Game** - a rhythm game where:
- Players play real bass notes detected through microphone
- Notes scroll toward a play line (like Guitar Hero)
- Scoring: Perfect (±50ms), Good (±100ms), OK (±150ms), Miss
- Design aesthetic: "Studio Industrial" - dark theme, amber accents, technical typography

The Go desktop app is a working reference implementation. The web app will be a complete reimplementation using:
- Alpine.js 3.x for reactivity
- Tailwind CSS 4 for styling
- HTML5 Canvas for game rendering
- Web Audio API + pitchfinder for pitch detection
