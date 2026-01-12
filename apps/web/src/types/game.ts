/**
 * Core game state and gameplay types.
 */

// ============================================================================
// Hit Quality
// ============================================================================

export const HIT_QUALITY = {
  MISS: 0,
  OK: 1,
  GOOD: 2,
  PERFECT: 3,
} as const;

export type HitQuality = (typeof HIT_QUALITY)[keyof typeof HIT_QUALITY];

/** Get display text for a hit quality */
export function hitQualityText(quality: HitQuality): string {
  switch (quality) {
    case HIT_QUALITY.PERFECT:
      return 'Perfect!';
    case HIT_QUALITY.GOOD:
      return 'Good';
    case HIT_QUALITY.OK:
      return 'OK';
    default:
      return 'Miss';
  }
}

/** Get base score for a hit quality */
export function hitQualityScore(quality: HitQuality): number {
  switch (quality) {
    case HIT_QUALITY.PERFECT:
      return 100;
    case HIT_QUALITY.GOOD:
      return 50;
    case HIT_QUALITY.OK:
      return 25;
    default:
      return 0;
  }
}

// ============================================================================
// Game State Machine
// ============================================================================

export const GAME_STATE = {
  /** Song selection menu */
  MENU: 'MENU',
  /** Countdown before playing */
  PRE_START: 'PRE_START',
  /** Actively playing */
  PLAYING: 'PLAYING',
  /** Showing results screen */
  RESULTS: 'RESULTS',
} as const;

export type GameStateType = (typeof GAME_STATE)[keyof typeof GAME_STATE];

// ============================================================================
// Floating Score Display
// ============================================================================

export interface FloatingScore {
  /** Display text (e.g., "Perfect! x3") */
  readonly text: string;
  /** X position in canvas coordinates */
  readonly x: number;
  /** Y position in canvas coordinates */
  readonly y: number;
  /** Timestamp when this score was created (Date.now()) */
  readonly startTime: number;
  /** Hit quality determines color */
  readonly quality: HitQuality;
}

// ============================================================================
// Game State
// ============================================================================

export interface GameState {
  /** Current state in the game state machine */
  readonly state: GameStateType;
  /** Current playback time in seconds */
  readonly currentTime: number;
  /** Total accumulated score */
  readonly score: number;
  /** Current combo streak */
  readonly combo: number;
  /** Highest combo achieved this session */
  readonly maxCombo: number;
  /** Number of notes successfully hit */
  readonly notesHit: number;
  /** Number of notes missed */
  readonly notesMissed: number;
  /** Total notes in the current song */
  readonly totalNotes: number;
  /** Whether the game is actively playing */
  readonly isPlaying: boolean;
  /** Whether the game has finished */
  readonly isFinished: boolean;
  /** Active floating score displays */
  readonly floatingText: readonly FloatingScore[];
}

// ============================================================================
// Factory Functions
// ============================================================================

/** Create initial game state */
export function createInitialGameState(): GameState {
  return {
    state: GAME_STATE.MENU,
    currentTime: 0,
    score: 0,
    combo: 0,
    maxCombo: 0,
    notesHit: 0,
    notesMissed: 0,
    totalNotes: 0,
    isPlaying: false,
    isFinished: false,
    floatingText: [],
  };
}

/** Calculate accuracy as a percentage (0-100) */
export function calculateAccuracy(state: GameState): number {
  const total = state.notesHit + state.notesMissed;
  if (total === 0) {
    return 100;
  }
  return (state.notesHit / total) * 100;
}

/** Calculate combo multiplier based on current streak */
export function getComboMultiplier(combo: number): number {
  if (combo >= 50) return 4;
  if (combo >= 25) return 3;
  if (combo >= 10) return 2;
  return 1;
}
