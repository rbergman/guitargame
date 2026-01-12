/**
 * Type definitions barrel export.
 *
 * Re-exports all types from the types module for convenient imports:
 * import { GameState, Song, PitchResult } from './types';
 */

// Branded types for domain safety
export type { BPM, Cents, Frequency, Milliseconds, Seconds } from './branded.js';
export {
  bpm,
  cents,
  frequency,
  milliseconds,
  msToSeconds,
  seconds,
  secondsToMs,
} from './branded.js';

// Game state and gameplay types
export type { FloatingScore, GameState, GameStateType } from './game.js';
export {
  calculateAccuracy,
  createInitialGameState,
  GAME_STATE,
  getComboMultiplier,
  HIT_QUALITY,
  hitQualityScore,
  hitQualityText,
} from './game.js';
export type { HitQuality } from './game.js';

// Song, tablature, and tuning types
export type {
  Song,
  SongDefinition,
  StringIndex,
  StringTuning,
  TabNote,
  TabNoteDefinition,
  TabNoteState,
  Tuning,
} from './song.js';
export {
  createTabNote,
  createTabNoteState,
  getNoteNameForTab,
  getOctaveForTab,
  getSemitone,
  STRING,
  TUNINGS,
  TUNINGS_BY_NAME,
} from './song.js';

// Audio and pitch detection types
export type { AudioState, AudioStateType, PitchResult } from './audio.js';
export {
  AUDIO_STATE,
  createEmptyPitchResult,
  createInitialAudioState,
  DEFAULT_AUDIO_CONFIG,
  frequencyToNote,
  getPitchDisplayName,
  isPitchValid,
  noteToFrequency,
} from './audio.js';
