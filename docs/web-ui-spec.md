# Bass Guitar Practice Game - Web UI Specification

## Overview

A web-based bass guitar practice game that detects real notes played through a microphone and scores the player's timing accuracy against scrolling tablature.

---

## Design Thinking

**Purpose**: Help bass players practice with real-time pitch feedback
**Audience**: Musicians who appreciate craft and have aesthetic taste
**Tone**: **Studio Industrial** - like a high-end DAW or vintage tube amp interface
**Differentiation**: The scrolling tab should feel like a **live VU meter/oscilloscope** - studio equipment, not arcade game

**Why this aesthetic**: Musicians spend hours staring at DAWs, amp panels, and pedalboards. They respond to:
- Dark backgrounds (low-light environments)
- Amber/warm accent colors (tube amp LEDs, VU meter needles)
- Monospace/technical typography (gear displays)
- High contrast for focus during performance

---

## 5-Component Game Design Evaluation

| Component | Current State | Web UI Requirements |
|-----------|---------------|---------------------|
| **Clarity** | Notes scroll R→L, fret numbers visible | Add: beat markers, countdown, approach zone highlighting |
| **Motivation** | Score + Grade | Add: persistent high scores, streak tracking per exercise |
| **Response** | ±50/100/150ms windows | CRITICAL: <16ms visual response, audio ping on detection |
| **Satisfaction** | Color change + floating text | Add: audio feedback, combo glow, screen shake at thresholds |
| **Fit** | Generic game colors | Studio-grade: amber primary, green=hit, deep warm palette |

**Conflict priority applied**: Response > Clarity > Satisfaction

---

## Visual Design System

### Color Palette (CSS Variables)

```css
:root {
  /* Background layers */
  --bg-deep: #0a0a0c;
  --bg-surface: #12121a;
  --bg-elevated: #1a1a24;

  /* Primary - Amber (tube amp LED) */
  --amber-dim: #8b6914;
  --amber: #d4a017;
  --amber-bright: #ffc933;

  /* Feedback states */
  --hit-perfect: #22c55e;  /* Bright green */
  --hit-good: #84cc16;     /* Lime */
  --hit-ok: #eab308;       /* Yellow */
  --hit-miss: #ef4444;     /* Red */

  /* UI chrome */
  --string-color: #4a4a5a;
  --play-line: #22d3ee;    /* Cyan glow */
  --text-primary: #e4e4e7;
  --text-muted: #71717a;
}
```

### Typography

- **Display/Headers**: "JetBrains Mono" or "IBM Plex Mono" (technical, precise)
- **HUD/Numbers**: "DSEG7" or similar LED-style for score display
- **Body**: "Inter" only for menu descriptions (readable, neutral)

### Key Visual Elements

1. **Tab strings**: Subtle horizontal lines with slight glow
2. **Play line**: Vertical cyan bar with pulsing glow animation
3. **Notes**: Circular with dark interior, fret number centered, colored border based on hit state
4. **Approach zone**: 200ms before play line gets highlighted (prepare cue)

---

## State Machine

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│   [MENU] ──(select)──▶ [PRE_START] ──(play note)──▶ [PLAYING]          │
│     ▲                                                    │              │
│     │                                                    │              │
│     └──────────────(play note)────── [RESULTS] ◀────────┘              │
│                                        (song ends)                      │
└─────────────────────────────────────────────────────────────────────────┘
```

### State Transitions

- MENU → PRE_START: User clicks/selects exercise
- PRE_START → PLAYING: Any valid note detected (with 3-2-1 countdown)
- PLAYING → RESULTS: Song duration exceeded
- RESULTS → MENU: Any note detected or click

---

## Component Architecture (Alpine.js)

```
web/
├── index.html              # Single page app shell
├── styles/
│   ├── main.css           # Tailwind 4 + custom properties
│   └── fonts.css          # Font imports
├── js/
│   ├── app.js             # Alpine initialization
│   ├── stores/
│   │   ├── game.js        # Alpine.store('game') - state machine, score
│   │   ├── audio.js       # Alpine.store('audio') - mic, pitch detection
│   │   └── songs.js       # Alpine.store('songs') - exercise data
│   ├── components/
│   │   ├── exercise-menu.js   # Song selection list
│   │   ├── tab-canvas.js      # Main gameplay <canvas> renderer
│   │   ├── pitch-display.js   # Shows detected note + frequency
│   │   ├── score-hud.js       # Score, combo, accuracy overlay
│   │   └── results-screen.js  # End-game stats + grade
│   └── lib/
│       ├── pitch-detector.js  # Web Audio API + pitchfinder wrapper
│       ├── hit-detector.js    # Timing window matching
│       ├── audio-feedback.js  # Sound effects for hits
│       └── note-utils.js      # Frequency ↔ note conversion
└── data/
    └── exercises.json     # Converted from YAML
```

---

## Core Gameplay Loop (60fps)

```javascript
// Pseudocode - tab-canvas.js
function gameLoop(timestamp) {
  if (!$store.game.isPlaying) return;

  // 1. Update time
  const currentTime = (timestamp - startTimestamp) / 1000;
  $store.game.currentTime = currentTime;

  // 2. Get pitch (async, but we use last known)
  const pitch = $store.audio.currentPitch;

  // 3. Check hits (if pitch valid)
  if (pitch.confidence > 0.8) {
    checkNoteHits(pitch, currentTime);
  }

  // 4. Check for missed notes (passed play line + grace period)
  markMissedNotes(currentTime);

  // 5. Render
  renderTab(currentTime);

  // 6. Check song end
  if (currentTime > song.duration) {
    $store.game.state = 'RESULTS';
    return;
  }

  requestAnimationFrame(gameLoop);
}
```

---

## Hit Detection (Response-Critical)

### Timing Windows

- **Perfect**: ±50ms (±0.05s)
- **Good**: ±100ms (±0.10s)
- **OK**: ±150ms (±0.15s)
- **Miss**: >150ms or wrong note

### Pitch Matching

- ±50 cents tolerance (≈3% frequency variance)
- Must match expected note name AND octave

```javascript
// hit-detector.js
function checkHit(detectedPitch, noteTime, currentTime) {
  const timeDelta = Math.abs(noteTime - currentTime);

  if (timeDelta <= 0.050) return 'PERFECT';
  if (timeDelta <= 0.100) return 'GOOD';
  if (timeDelta <= 0.150) return 'OK';
  return null; // Not yet in window or missed
}
```

---

## Audio Architecture (Web Audio API)

```javascript
// pitch-detector.js
async function initAudio() {
  const stream = await navigator.mediaDevices.getUserMedia({
    audio: {
      echoCancellation: false,
      noiseSuppression: false,
      autoGainControl: false,
      sampleRate: 48000
    }
  });

  const audioContext = new AudioContext({ sampleRate: 48000 });
  const source = audioContext.createMediaStreamSource(stream);
  const analyser = audioContext.createAnalyser();

  analyser.fftSize = 2048;
  analyser.smoothingTimeConstant = 0;

  source.connect(analyser);

  // Use pitchfinder library for YIN algorithm
  const detector = PitchFinder.YIN({ sampleRate: 48000 });

  // Poll at ~60fps
  setInterval(() => {
    const buffer = new Float32Array(analyser.fftSize);
    analyser.getFloatTimeDomainData(buffer);

    const frequency = detector(buffer);
    if (frequency) {
      updatePitch({
        frequency,
        note: frequencyToNote(frequency),
        confidence: calculateConfidence(buffer)
      });
    }
  }, 16);
}
```

---

## Tab Canvas Rendering

### Layout (1000x400 viewport)

```
┌────────────────────────────────────────────────────────────┐
│  E Minor Scale                     Score: 2400  Combo: 12  │  ← Header (40px)
├────────────────────────────────────────────────────────────┤
│ G ═══════════════════════════════════╪══════════════════ │
│ D ═══════════════════════════════════╪══════════════════ │  ← Tab Area
│ A ═══════════════════════════════════╪══════════════════ │    (280px)
│ E ═══════════════════════════════════╪══════════════════ │
│                                      │← Play Line (75%)   │
├────────────────────────────────────────────────────────────┤
│  Playing: E2  41.2 Hz                                      │  ← Pitch Display (40px)
└────────────────────────────────────────────────────────────┘
```

### Note Rendering

```javascript
function renderNote(ctx, note, x, y, state) {
  const radius = 20;

  // Glow effect for approaching notes
  if (state === 'approaching') {
    ctx.shadowColor = 'rgba(212, 160, 23, 0.6)';
    ctx.shadowBlur = 15;
  }

  // Circle background
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, Math.PI * 2);
  ctx.fillStyle = '#1a1a24';
  ctx.fill();

  // Border color based on hit state
  ctx.strokeStyle = getHitColor(state);
  ctx.lineWidth = 3;
  ctx.stroke();

  // Fret number
  ctx.fillStyle = '#e4e4e7';
  ctx.font = 'bold 16px "JetBrains Mono"';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(note.fret.toString(), x, y);
}
```

---

## Satisfaction Feedback (Multi-Channel)

Per game design requirement: **minimum 2 feedback channels for significant actions**.

| Event | Visual | Audio | Additional |
|-------|--------|-------|------------|
| **Perfect hit** | Green border + "Perfect!" float | High ping (C5) | +100 combo flash |
| **Good hit** | Lime border + "Good" float | Mid ping (G4) | |
| **OK hit** | Yellow border + "OK" float | Low ping (E4) | |
| **Miss** | Red border + "Miss" float | Dull thud | Combo reset shake |
| **Combo 10** | HUD glow pulse | Rising chime | "2x" indicator |
| **Combo 25** | Screen edge glow | Ascending arpeggio | "3x" indicator |
| **Combo 50** | Full border glow | Fanfare | "4x" indicator |

---

## Animation & Motion Design

### Principles

1. **Responsive**: All interactions respond within 1 frame (16ms)
2. **Purposeful**: Motion communicates state changes, not decoration
3. **Rhythmic**: Animations should feel musical - ease curves that breathe

### Key Animations

| Element | Animation | Duration | Easing |
|---------|-----------|----------|--------|
| Note approach glow | Pulse intensity | 200ms loop | ease-in-out |
| Play line | Subtle pulse | 500ms loop | ease-in-out |
| Hit feedback | Scale up + fade | 300ms | ease-out |
| Floating text | Rise + fade | 800ms | ease-out |
| Combo multiplier | Pop + settle | 200ms | spring |
| Screen shake | Horizontal oscillate | 150ms | ease-out |
| Menu item hover | Scale + glow | 150ms | ease-out |
| State transitions | Crossfade | 300ms | ease-in-out |

### CSS Animation Variables

```css
:root {
  --anim-fast: 150ms;
  --anim-normal: 300ms;
  --anim-slow: 500ms;
  --ease-out: cubic-bezier(0.33, 1, 0.68, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
}
```

---

## Song Data Format (JSON)

```json
{
  "id": "e-minor-scale",
  "title": "E Minor Scale",
  "artist": "Exercise",
  "bpm": 80,
  "tuning": "standard",
  "notes": [
    { "beat": 0, "string": 3, "fret": 0 },
    { "beat": 1, "string": 3, "fret": 2 },
    { "beat": 2, "string": 3, "fret": 3 },
    { "beat": 3, "string": 2, "fret": 0 }
  ]
}
```

**Runtime conversion**: `time = beat * (60 / bpm)`

---

## Playtest Scenarios

1. **New player test**: Can they understand "play the note when it reaches the line" without instruction?
   - Pass: 80% of users start playing correctly within 10 seconds

2. **Stress test**: Rapid note sequences, spam inputs
   - Pass: No duplicate hits, no missed detection, stable 60fps

3. **Skill test**: Can timing precision improve score?
   - Pass: Consistent player can achieve >90% Perfect rate

4. **Latency test**: Audio detection lag
   - Pass: <50ms from physical pluck to visual feedback

5. **Readability test**: Can an observer see what's happening?
   - Pass: Clear distinction between upcoming, active, hit, and missed notes

---

## Implementation Priority

### Phase 1 - Core Loop (Response + Clarity)

1. Web Audio pitch detection working
2. Canvas rendering with scrolling notes
3. Hit detection with visual feedback
4. Basic score tracking

### Phase 2 - Polish (Satisfaction)

1. Audio feedback pings
2. Floating text animations
3. Combo effects
4. Grade calculation + results screen

### Phase 3 - Content (Motivation)

1. Exercise menu with all 10 songs
2. Local storage high scores
3. Streak tracking

---

## Technology Stack

| Layer | Technology | Why |
|-------|------------|-----|
| **Framework** | Alpine.js 3.x | Lightweight reactivity, no build step needed |
| **Styling** | Tailwind CSS 4 | Utility-first, CSS variables support |
| **Rendering** | HTML5 Canvas | 60fps game rendering, pixel control |
| **Audio Input** | Web Audio API | Native browser, low latency |
| **Pitch Detection** | pitchfinder (YIN) | Proven algorithm, JS implementation |
| **Fonts** | JetBrains Mono, DSEG7 | Technical aesthetic |
| **Build** | Vite | Fast dev server, HMR |

---

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Frame rate | 60fps stable | requestAnimationFrame timing |
| Input latency | <50ms | Timestamp delta from pluck to visual |
| Pitch detection | <16ms | Web Audio callback timing |
| First paint | <1s | Lighthouse |
| Bundle size | <200KB | Compressed JS/CSS |
