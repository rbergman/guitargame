# TypeScript Migration & Monorepo Setup Plan

## Overview

Convert the guitargame repository from a single Go desktop app to a monorepo with:
- **apps/desktop/** - Original Go Gioui app (preserved as reference)
- **apps/web/** - New TypeScript web app (primary development focus)
- **songs/** - Shared song data (YAML source files)

## Rationale

The Go code is a complete desktop GUI application, not a backend API. The web UI is a separate implementation of the same game concept. Keeping both:
- Preserves working reference implementation
- Allows future desktop development if desired
- Shares song content between implementations

---

## Phase 1: Repository Restructure

### 1.1 Create Monorepo Directory Structure

```
guitargame/
├── .mise.toml              # Tool versions: node@22, just
├── .envrc                  # Direnv (optional, for env vars)
├── justfile                # Root router (monorepo)
├── CLAUDE.md               # Agent instructions (update)
├── README.md               # Project overview
├── docs/
│   ├── web-ui-spec.md      # Existing spec
│   └── typescript-migration-plan.md  # This document
├── .beads/                 # Issue tracking (existing)
├── apps/
│   ├── desktop/            # ← Move existing Go code here
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Makefile        # Keep for Go builds
│   │   ├── .goreleaser.yml
│   │   └── internal/
│   │       ├── audio/
│   │       ├── game/
│   │       ├── render/
│   │       └── song/
│   └── web/                # ← New TypeScript app
│       ├── package.json
│       ├── tsconfig.json
│       ├── eslint.config.js
│       ├── justfile        # Module justfile
│       ├── index.html      # Vite entry
│       ├── src/
│       │   ├── main.ts
│       │   ├── app.ts      # Alpine.js init
│       │   ├── types/      # Shared type definitions
│       │   ├── stores/     # Alpine stores
│       │   ├── components/ # UI components
│       │   └── lib/        # Utilities
│       ├── styles/
│       │   └── main.css    # Tailwind + custom
│       └── public/
│           └── fonts/
└── songs/                  # Shared YAML source (stays at root)
    ├── 01-e-minor-scale.yaml
    └── ...
```

### 1.2 Move Go Code

```bash
# Create apps directory
mkdir -p apps/desktop

# Move Go files to apps/desktop/
git mv main.go apps/desktop/
git mv go.mod apps/desktop/
git mv go.sum apps/desktop/
git mv Makefile apps/desktop/
git mv .goreleaser.yml apps/desktop/
git mv internal apps/desktop/
```

### 1.3 Update Go Module Path

In `apps/desktop/go.mod`:
```go
module guitargame/apps/desktop

// Update internal imports in main.go:
// "guitargame/internal/audio" → "guitargame/apps/desktop/internal/audio"
```

---

## Phase 2: Tooling Setup

### 2.1 Mise Configuration

Create `.mise.toml` at repo root:

```toml
[tools]
node = "22"
just = "latest"

[settings]
auto_install = true
```

### 2.2 Root Justfile (Monorepo Router)

```just
# Guitar Game Monorepo
# Usage: just --list
# Usage: just web <recipe>

set shell := ["mise", "exec", "--", "bash", "-c"]

# TypeScript web app
mod web "apps/web"

# Go desktop app (reference)
mod desktop "apps/desktop"

default:
    @just --list

# === Umbrella Recipes ===

# Run all quality checks
check: web::check
    @echo "All checks passed"

# Initial setup for new contributors
setup:
    mise trust
    mise install
    just web::setup
    @echo "Ready to develop"

# Clean all build artifacts
clean: web::clean
    @echo "Cleaned"
```

### 2.3 Web App Justfile (`apps/web/justfile`)

```just
# Web App Build System
# Usage: just web <recipe>

default:
    @just --list

# === Setup ===

# Install dependencies
setup:
    npm ci

# === Quality Gates ===

# Full quality check (pre-commit)
check: typecheck lint test
    @echo "Web checks passed"

# Quick check (dev iteration)
check-quick: typecheck lint
    @echo "Quick checks passed"

# TypeScript type checking
typecheck:
    npx tsc --noEmit

# ESLint with auto-fix
lint:
    npx eslint src/ --fix

# Run tests
test:
    npx vitest run

# === Development ===

# Start dev server
dev:
    npx vite

# Production build
build:
    npx vite build

# Preview production build
preview:
    npx vite preview

# === Maintenance ===

# Remove build artifacts
clean:
    rm -rf dist node_modules/.cache
```

### 2.4 Desktop App Justfile (`apps/desktop/justfile`)

```just
# Desktop App Build System (Go/Gioui)
# Usage: just desktop <recipe>
# Note: This is the reference implementation

default:
    @just --list

# Build for current platform
build:
    go build -o bin/guitargame .

# Run the app
run: build
    ./bin/guitargame

# Run tests
test:
    go test -race ./...

# Clean build artifacts
clean:
    rm -rf bin/
```

---

## Phase 3: TypeScript Project Setup

### 3.1 Initialize Vite + TypeScript

```bash
cd apps/web
npm create vite@latest . -- --template vanilla-ts
```

### 3.2 Install Dependencies

```bash
# Core
npm install alpinejs

# Dev dependencies
npm install -D \
  typescript \
  typescript-eslint \
  @eslint-community/eslint-plugin-eslint-comments \
  vitest \
  tailwindcss@latest \
  @tailwindcss/vite \
  pitchfinder \
  husky
```

### 3.3 TypeScript Configuration (`tsconfig.json`)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "noPropertyAccessFromIndexSignature": true,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true,
    "isolatedModules": true,
    "verbatimModuleSyntax": true,
    "skipLibCheck": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "dist",
    "rootDir": "src",
    "types": ["vite/client"]
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

### 3.4 ESLint Configuration (`eslint.config.js`)

```javascript
import tseslint from 'typescript-eslint';
import eslintComments from '@eslint-community/eslint-plugin-eslint-comments/configs';

export default tseslint.config(
  eslintComments.recommended,
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      // Zero tolerance for any
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unsafe-argument': 'error',
      '@typescript-eslint/no-unsafe-assignment': 'error',
      '@typescript-eslint/no-unsafe-call': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'error',
      '@typescript-eslint/no-unsafe-return': 'error',

      // Promise handling
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-misused-promises': 'error',
      '@typescript-eslint/require-await': 'error',

      // Type assertions
      '@typescript-eslint/consistent-type-assertions': ['error', {
        assertionStyle: 'never'
      }],

      // Explicit return types for exports
      '@typescript-eslint/explicit-function-return-type': ['error', {
        allowExpressions: true,
        allowTypedFunctionExpressions: true,
      }],

      // Code quality
      'complexity': ['error', 10],
      'max-depth': ['error', 4],
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_'
      }],
    },
  },
  {
    ignores: ['dist/**', 'node_modules/**', '*.config.js'],
  }
);
```

### 3.5 Package.json Scripts

```json
{
  "name": "@guitargame/web",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "typecheck": "tsc --noEmit",
    "lint": "eslint src/",
    "lint:fix": "eslint src/ --fix",
    "test": "vitest run",
    "test:watch": "vitest",
    "check": "npm run typecheck && npm run lint && npm run test",
    "prepare": "cd ../.. && husky apps/web/.husky"
  }
}
```

### 3.6 Vite Configuration (`vite.config.ts`)

```typescript
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [tailwindcss()],
  build: {
    target: 'ES2022',
    sourcemap: true,
  },
});
```

### 3.7 Tailwind CSS Setup (`styles/main.css`)

```css
@import 'tailwindcss';

/* Design System - Studio Industrial */
@theme {
  /* Background layers */
  --color-bg-deep: #0a0a0c;
  --color-bg-surface: #12121a;
  --color-bg-elevated: #1a1a24;

  /* Primary - Amber (tube amp LED) */
  --color-amber-dim: #8b6914;
  --color-amber: #d4a017;
  --color-amber-bright: #ffc933;

  /* Feedback states */
  --color-hit-perfect: #22c55e;
  --color-hit-good: #84cc16;
  --color-hit-ok: #eab308;
  --color-hit-miss: #ef4444;

  /* UI chrome */
  --color-string: #4a4a5a;
  --color-play-line: #22d3ee;
  --color-text-primary: #e4e4e7;
  --color-text-muted: #71717a;

  /* Animation timing */
  --ease-out: cubic-bezier(0.33, 1, 0.68, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);

  /* Font families */
  --font-mono: 'JetBrains Mono', monospace;
  --font-display: 'DSEG7 Classic', monospace;
}
```

---

## Phase 4: TypeScript Architecture

### 4.1 Directory Structure

```
apps/web/src/
├── main.ts                 # Entry point
├── app.ts                  # Alpine.js initialization
├── types/
│   ├── index.ts            # Re-exports
│   ├── game.ts             # GameState, HitQuality, etc.
│   ├── song.ts             # Song, TabNote, Tuning
│   └── audio.ts            # PitchResult, AudioState
├── stores/
│   ├── game.ts             # Alpine.store('game')
│   ├── audio.ts            # Alpine.store('audio')
│   └── songs.ts            # Alpine.store('songs')
├── components/
│   ├── exercise-menu.ts    # Song selection
│   ├── tab-canvas.ts       # Main gameplay canvas
│   ├── pitch-display.ts    # Detected note indicator
│   ├── score-hud.ts        # Score/combo overlay
│   └── results-screen.ts   # End game stats
└── lib/
    ├── pitch-detector.ts   # Web Audio + pitchfinder
    ├── hit-detector.ts     # Timing window logic
    ├── audio-feedback.ts   # Sound effects
    └── note-utils.ts       # Frequency ↔ note conversion
```

### 4.2 Core Type Definitions (`types/game.ts`)

```typescript
export const HIT_QUALITY = {
  MISS: 0,
  OK: 1,
  GOOD: 2,
  PERFECT: 3,
} as const;

export type HitQuality = typeof HIT_QUALITY[keyof typeof HIT_QUALITY];

export const GAME_STATE = {
  MENU: 'MENU',
  PRE_START: 'PRE_START',
  PLAYING: 'PLAYING',
  RESULTS: 'RESULTS',
} as const;

export type GameStateType = typeof GAME_STATE[keyof typeof GAME_STATE];

export interface GameState {
  readonly state: GameStateType;
  readonly currentTime: number;
  readonly score: number;
  readonly combo: number;
  readonly maxCombo: number;
  readonly notesHit: number;
  readonly notesMissed: number;
  readonly totalNotes: number;
  readonly isPlaying: boolean;
  readonly isFinished: boolean;
  readonly floatingText: readonly FloatingScore[];
}

export interface FloatingScore {
  readonly text: string;
  readonly x: number;
  readonly y: number;
  readonly startTime: number;
  readonly quality: HitQuality;
}
```

### 4.3 Branded Types for Domain Safety

```typescript
// types/branded.ts
declare const brand: unique symbol;

type Brand<T, B> = T & { readonly [brand]: B };

export type Frequency = Brand<number, 'Frequency'>;
export type Cents = Brand<number, 'Cents'>;
export type BPM = Brand<number, 'BPM'>;
export type Milliseconds = Brand<number, 'Milliseconds'>;

// Constructor functions with validation
export function frequency(hz: number): Frequency {
  if (hz < 20 || hz > 20000) {
    throw new RangeError(`Frequency ${hz}Hz out of audible range`);
  }
  return hz as Frequency;
}

export function bpm(value: number): BPM {
  if (value < 20 || value > 300) {
    throw new RangeError(`BPM ${value} out of reasonable range`);
  }
  return value as BPM;
}
```

---

## Phase 5: Migration Checklist

### 5.1 Beads Updates Required

**New Epic**: "Epic: Repository Setup & TypeScript Migration" (P0)

**New Tasks** (frontloaded):
1. Create monorepo directory structure
2. Move Go code to apps/desktop/
3. Update Go module imports
4. Create .mise.toml with node@22, just
5. Create root justfile (monorepo router)
6. Initialize Vite TypeScript project
7. Configure strict tsconfig.json
8. Configure ESLint strict-type-checked
9. Create apps/web/justfile
10. Set up Tailwind 4 with design system
11. Create core type definitions
12. Set up husky pre-commit hooks

**Update Existing Tasks**:
- "Initialize Vite project" → now includes TypeScript template
- "Install Tailwind CSS 4" → moves to apps/web/
- "Create design system CSS" → becomes Tailwind @theme config
- "Set up project directory structure" → reflects new monorepo layout

### 5.2 Dependency Order

```
[Repository Setup] (new epic - P0)
    ↓
[Project Setup & Foundation] (existing - becomes web-specific)
    ↓
[Audio Pipeline] → [Core Game Logic] → [Canvas] → [UI] → [Feedback] → [Content]
```

---

## Phase 6: Verification

### 6.1 Quality Gates Pass

```bash
# From repo root
just setup           # Install tooling
just check           # All quality gates

# Or specifically
just web check       # TypeScript checks only
just web dev         # Start dev server
```

### 6.2 Type Safety Verification

- `npm run typecheck` passes with zero errors
- `npm run lint` passes with strict rules
- No `any` types in codebase
- All exported functions have explicit return types
- All promises handled explicitly

### 6.3 Dev Experience

```bash
# New contributor workflow
git clone <repo>
just setup           # One command setup
just web dev         # Start coding
```

---

## Summary

| Phase | Deliverable | Priority |
|-------|-------------|----------|
| 1 | Monorepo structure with Go code moved | P0 |
| 2 | mise + just tooling configured | P0 |
| 3 | Vite + TypeScript + Tailwind project | P0 |
| 4 | Type definitions and architecture | P1 |
| 5 | Beads updated with new tasks | P0 |
| 6 | All quality gates passing | P0 |

**Estimated new beads**: 12 tasks for repo setup (frontloaded before existing work)
