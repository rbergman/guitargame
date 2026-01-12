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

# === Claude Autonomous Recipes ===

# Run Claude interactively
ai:
    claude

# Run Claude autonomously (sandboxed with network access)
ai-auto prompt:
    srt -s .srt.json -c 'claude --dangerously-skip-permissions \
      --no-session-persistence \
      -p "{{prompt}}"'

# Run Claude autonomously with full MCP access (Context7, Brightdata)
ai-auto-mcp prompt:
    srt -s .srt.json -c 'claude --dangerously-skip-permissions \
      --no-session-persistence \
      -p "{{prompt}}"'
