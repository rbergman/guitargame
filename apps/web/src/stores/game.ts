/**
 * Alpine.js game store - state machine and scoring logic.
 */
import Alpine from 'alpinejs';
import {
  type GameStateType,
  type HitQuality,
  type FloatingScore,
  GAME_STATE,
  createInitialGameState,
  hitQualityScore,
  hitQualityText,
  getComboMultiplier,
} from '../types/game';

// ============================================================================
// Store Interface
// ============================================================================

/**
 * Mutable game store interface for Alpine.js.
 * Based on GameState but with mutable properties for store updates.
 */
export interface GameStore {
  // State (mutable for Alpine)
  state: GameStateType;
  currentTime: number;
  score: number;
  combo: number;
  maxCombo: number;
  notesHit: number;
  notesMissed: number;
  totalNotes: number;
  isPlaying: boolean;
  isFinished: boolean;
  floatingText: FloatingScore[];

  // State transitions
  selectSong(totalNotes: number): void;
  startPlaying(): void;
  endGame(): void;
  returnToMenu(): void;

  // Gameplay
  recordHit(quality: HitQuality, x: number, y: number): void;
  recordMiss(): void;
  updateTime(time: number): void;
  clearFloatingText(olderThan: number): void;
}

// ============================================================================
// Store Implementation
// ============================================================================

function createGameStore(): GameStore {
  const initial = createInitialGameState();

  return {
    // State from createInitialGameState
    state: initial.state,
    currentTime: initial.currentTime,
    score: initial.score,
    combo: initial.combo,
    maxCombo: initial.maxCombo,
    notesHit: initial.notesHit,
    notesMissed: initial.notesMissed,
    totalNotes: initial.totalNotes,
    isPlaying: initial.isPlaying,
    isFinished: initial.isFinished,
    floatingText: [...initial.floatingText],

    // ========================================================================
    // State Transitions
    // ========================================================================

    /**
     * Transition from MENU to PRE_START.
     * Sets total notes for the selected song.
     */
    selectSong(totalNotes: number): void {
      if (this.state !== GAME_STATE.MENU) return;

      this.state = GAME_STATE.PRE_START;
      this.totalNotes = totalNotes;
      this.isPlaying = false;
      this.isFinished = false;
    },

    /**
     * Transition from PRE_START to PLAYING.
     * Called when countdown completes or first note detected.
     */
    startPlaying(): void {
      if (this.state !== GAME_STATE.PRE_START) return;

      this.state = GAME_STATE.PLAYING;
      this.isPlaying = true;
      this.currentTime = 0;
    },

    /**
     * Transition from PLAYING to RESULTS.
     * Called when song ends.
     */
    endGame(): void {
      if (this.state !== GAME_STATE.PLAYING) return;

      this.state = GAME_STATE.RESULTS;
      this.isPlaying = false;
      this.isFinished = true;
    },

    /**
     * Transition from any state back to MENU.
     * Resets all gameplay state.
     */
    returnToMenu(): void {
      const reset = createInitialGameState();
      this.state = reset.state;
      this.currentTime = reset.currentTime;
      this.score = reset.score;
      this.combo = reset.combo;
      this.maxCombo = reset.maxCombo;
      this.notesHit = reset.notesHit;
      this.notesMissed = reset.notesMissed;
      this.totalNotes = reset.totalNotes;
      this.isPlaying = reset.isPlaying;
      this.isFinished = reset.isFinished;
      this.floatingText = [];
    },

    // ========================================================================
    // Gameplay Actions
    // ========================================================================

    /**
     * Record a successful hit.
     * Updates score with combo multiplier, increments combo, adds floating text.
     */
    recordHit(quality: HitQuality, x: number, y: number): void {
      if (this.state !== GAME_STATE.PLAYING) return;

      // Update combo first (affects multiplier)
      this.combo += 1;
      if (this.combo > this.maxCombo) {
        this.maxCombo = this.combo;
      }

      // Calculate score with multiplier
      const baseScore = hitQualityScore(quality);
      const multiplier = getComboMultiplier(this.combo);
      this.score += baseScore * multiplier;

      // Track notes hit
      this.notesHit += 1;

      // Add floating text
      const text =
        multiplier > 1
          ? `${hitQualityText(quality)} x${String(multiplier)}`
          : hitQualityText(quality);

      const floatingScore: FloatingScore = {
        text,
        x,
        y,
        startTime: Date.now(),
        quality,
      };

      this.floatingText = [...this.floatingText, floatingScore];
    },

    /**
     * Record a missed note.
     * Resets combo, increments missed count.
     */
    recordMiss(): void {
      if (this.state !== GAME_STATE.PLAYING) return;

      this.combo = 0;
      this.notesMissed += 1;
    },

    /**
     * Update current playback time.
     */
    updateTime(time: number): void {
      this.currentTime = time;
    },

    /**
     * Remove floating text older than specified timestamp.
     */
    clearFloatingText(olderThan: number): void {
      this.floatingText = this.floatingText.filter(
        (ft) => ft.startTime > olderThan
      );
    },
  };
}

// ============================================================================
// Store Registration
// ============================================================================

export function registerGameStore(): void {
  Alpine.store('game', createGameStore());
}

// Type augmentation for Alpine
declare module 'alpinejs' {
  interface Stores {
    game: GameStore;
  }
}
