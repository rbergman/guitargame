/**
 * Song, tablature, and tuning types.
 *
 * Modeled after the Go types in apps/desktop/internal/song/types.go
 */

import type { BPM, Seconds } from './branded.js';
import type { HitQuality } from './game.js';

// ============================================================================
// String Constants (Bass guitar)
// ============================================================================

/** Bass string indices (standard 4-string, high to low) */
export const STRING = {
  G: 0, // G string (highest, thinnest)
  D: 1, // D string
  A: 2, // A string
  E: 3, // E string (lowest, thickest)
  B: 4, // B string (5-string bass only)
} as const;

export type StringIndex = (typeof STRING)[keyof typeof STRING];

// ============================================================================
// Tuning
// ============================================================================

/** Tuning of a single string */
export interface StringTuning {
  /** Note name (C, C#, D, Db, etc.) */
  readonly note: string;
  /** Octave number */
  readonly octave: number;
}

/** Tuning for all strings (array from high to low) */
export type Tuning = readonly StringTuning[];

/** Predefined tunings */
export const TUNINGS = {
  /** Standard 4-string bass (G-D-A-E) */
  STANDARD: [
    { note: 'G', octave: 2 },
    { note: 'D', octave: 2 },
    { note: 'A', octave: 1 },
    { note: 'E', octave: 1 },
  ] as const satisfies Tuning,

  /** Drop D tuning (G-D-A-D) */
  DROP_D: [
    { note: 'G', octave: 2 },
    { note: 'D', octave: 2 },
    { note: 'A', octave: 1 },
    { note: 'D', octave: 1 },
  ] as const satisfies Tuning,

  /** Half step down (Gb-Db-Ab-Eb) */
  HALF_STEP_DOWN: [
    { note: 'Gb', octave: 2 },
    { note: 'Db', octave: 2 },
    { note: 'Ab', octave: 1 },
    { note: 'Eb', octave: 1 },
  ] as const satisfies Tuning,

  /** Full step down (F-C-G-D) */
  FULL_STEP_DOWN: [
    { note: 'F', octave: 2 },
    { note: 'C', octave: 2 },
    { note: 'G', octave: 1 },
    { note: 'D', octave: 1 },
  ] as const satisfies Tuning,

  /** Standard 5-string bass (G-D-A-E-B) */
  FIVE_STRING: [
    { note: 'G', octave: 2 },
    { note: 'D', octave: 2 },
    { note: 'A', octave: 1 },
    { note: 'E', octave: 1 },
    { note: 'B', octave: 0 },
  ] as const satisfies Tuning,
} as const;

/** Map of tuning names to tuning values */
export const TUNINGS_BY_NAME: Readonly<Record<string, Tuning>> = {
  standard: TUNINGS.STANDARD,
  'drop-d': TUNINGS.DROP_D,
  'half-step-down': TUNINGS.HALF_STEP_DOWN,
  'full-step-down': TUNINGS.FULL_STEP_DOWN,
  '5-string': TUNINGS.FIVE_STRING,
};

// ============================================================================
// Tab Note
// ============================================================================

/** A single note in tablature (immutable definition) */
export interface TabNoteDefinition {
  /** Time in seconds from song start */
  readonly time: Seconds;
  /** Beat number (alternative to time, converted using BPM) */
  readonly beat?: number;
  /** String index (0=G, 1=D, 2=A, 3=E) */
  readonly string: StringIndex;
  /** Fret number (0 = open string) */
  readonly fret: number;
  /** Note duration in seconds */
  readonly duration: Seconds;
}

/** Runtime state for a tab note (mutable during gameplay) */
export interface TabNoteState {
  /** Whether this note has been hit */
  hit: boolean;
  /** Quality of the hit */
  hitQuality: HitQuality;
  /** Time when the note was hit */
  hitTime: number;
}

/** Combined tab note with definition and runtime state */
export interface TabNote extends TabNoteDefinition {
  /** Runtime state (mutable) */
  state: TabNoteState;
}

// ============================================================================
// Song
// ============================================================================

/** Song definition (immutable, loaded from YAML) */
export interface SongDefinition {
  /** Song title */
  readonly title: string;
  /** Artist name */
  readonly artist: string;
  /** Tempo in beats per minute */
  readonly bpm: BPM;
  /** Tuning name or custom string (e.g., "standard", "drop-d", "G2,D2,A1,D1") */
  readonly tuningStr?: string;
  /** Tab notes (positions and timing) */
  readonly notes: readonly TabNoteDefinition[];
}

/** Song with runtime state */
export interface Song {
  /** Song metadata */
  readonly definition: SongDefinition;
  /** Parsed tuning (resolved from tuningStr or default) */
  readonly tuning: Tuning;
  /** Calculated duration in seconds */
  readonly duration: Seconds;
  /** Notes with runtime state */
  notes: TabNote[];
}

// ============================================================================
// Factory Functions
// ============================================================================

/** Create initial state for a tab note */
export function createTabNoteState(): TabNoteState {
  return {
    hit: false,
    hitQuality: 0, // HIT_QUALITY.MISS
    hitTime: 0,
  };
}

/** Create a TabNote from a definition */
export function createTabNote(definition: TabNoteDefinition): TabNote {
  return {
    ...definition,
    state: createTabNoteState(),
  };
}

// ============================================================================
// Utility Functions
// ============================================================================

/** Note names for semitone calculation */
const NOTE_SEMITONES: Readonly<Record<string, number>> = {
  C: 0,
  'C#': 1,
  Db: 1,
  D: 2,
  'D#': 3,
  Eb: 3,
  E: 4,
  Fb: 4,
  F: 5,
  'F#': 6,
  Gb: 6,
  G: 7,
  'G#': 8,
  Ab: 8,
  A: 9,
  'A#': 10,
  Bb: 10,
  B: 11,
  Cb: 11,
};

const NOTE_NAMES = [
  'C',
  'C#',
  'D',
  'D#',
  'E',
  'F',
  'F#',
  'G',
  'G#',
  'A',
  'A#',
  'B',
] as const;

/** Get semitone value (0-11) for a note name */
export function getSemitone(note: string): number {
  return NOTE_SEMITONES[note] ?? 0;
}

/** Get the note name for a tab position using the given tuning */
export function getNoteNameForTab(
  stringIndex: number,
  fret: number,
  tuning: Tuning
): string {
  const stringTuning = tuning[stringIndex];
  if (!stringTuning) {
    return '?';
  }

  if (fret === 0) {
    return stringTuning.note;
  }

  const baseNote = getSemitone(stringTuning.note);
  const noteIndex = (baseNote + fret) % 12;
  return NOTE_NAMES[noteIndex] ?? '?';
}

/** Get the octave for a tab position using the given tuning */
export function getOctaveForTab(
  stringIndex: number,
  fret: number,
  tuning: Tuning
): number {
  const stringTuning = tuning[stringIndex];
  if (!stringTuning) {
    return 1;
  }

  const baseNote = getSemitone(stringTuning.note);
  const notePos = baseNote + fret;

  return stringTuning.octave + Math.floor(notePos / 12);
}
