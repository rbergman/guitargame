# Guitar Game Monorepo
# Usage: just --list
# Usage: just web <recipe>

set shell := ["mise", "exec", "--", "bash", "-c"]

# TypeScript web app
mod web "apps/web"

# Go desktop app (reference implementation)
mod desktop "apps/desktop"

default:
    @just --list

# === Umbrella Recipes ===

# Run all quality checks
check: web::check
    @echo "All checks passed"

# Quick check (dev iteration)
check-quick: web::check-quick
    @echo "Quick checks passed"

# Initial setup for new contributors
setup:
    mise trust
    mise install
    just web::setup
    @echo "Ready to develop"

# Clean all build artifacts
clean: web::clean
    @echo "Cleaned"

# === Claude Recipes ===

# Run Claude interactively (no sandbox)
ai:
    claude

# Run Claude interactively with sandbox (dangerous mode, but filesystem/network restricted)
ai-sandboxed:
    srt -s .srt.json -c 'claude --dangerously-skip-permissions'

# Run Claude autonomously (sandboxed, non-interactive)
ai-auto prompt:
    srt -s .srt.json -c 'claude --dangerously-skip-permissions \
      --no-session-persistence \
      -p "{{prompt}}"'
