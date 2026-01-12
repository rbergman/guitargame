/**
 * Audio and pitch detection types.
 *
 * Modeled after the Go types in apps/desktop/internal/audio/
 */

import type { Cents, Frequency } from './branded.js';

// ============================================================================
// Pitch Detection
// ============================================================================

/** Result of pitch detection analysis */
export interface PitchResult {
  /** Detected frequency in Hz */
  readonly frequency: Frequency;
  /** Detection confidence (0-1) */
  readonly confidence: number;
  /** Detected note name (C, C#, D, etc.) or empty if invalid */
  readonly note: string;
  /** Octave number */
  readonly octave: number;
  /** Cents deviation from perfect pitch (-50 to +50) */
  readonly cents: Cents;
  /** Root mean square (signal strength) */
  readonly rms: number;
}

/** Check if a pitch result represents a valid detected note */
export function isPitchValid(result: PitchResult): boolean {
  return result.note !== '' && result.confidence > 0.5;
}

/** Get display name for a pitch result (e.g., "A2" or "--") */
export function getPitchDisplayName(result: PitchResult): string {
  if (result.note === '') {
    return '--';
  }
  return `${result.note}${String(result.octave)}`;
}

/** Create an empty/invalid pitch result */
export function createEmptyPitchResult(): PitchResult {
  return {
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Zero-value branded type for empty state: frequency(0) would throw RangeError
    frequency: 0 as Frequency,
    confidence: 0,
    note: '',
    octave: 0,
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Zero-value branded type for empty state
    cents: 0 as Cents,
    rms: 0,
  };
}

// ============================================================================
// Audio Input State
// ============================================================================

/** State of the audio input system */
export const AUDIO_STATE = {
  /** Not initialized */
  UNINITIALIZED: 'UNINITIALIZED',
  /** Requesting microphone permission */
  REQUESTING_PERMISSION: 'REQUESTING_PERMISSION',
  /** Permission denied by user */
  PERMISSION_DENIED: 'PERMISSION_DENIED',
  /** Initializing audio context and stream */
  INITIALIZING: 'INITIALIZING',
  /** Ready but not actively listening */
  READY: 'READY',
  /** Actively capturing and analyzing audio */
  LISTENING: 'LISTENING',
  /** Error state */
  ERROR: 'ERROR',
} as const;

export type AudioStateType = (typeof AUDIO_STATE)[keyof typeof AUDIO_STATE];

/** Audio system state */
export interface AudioState {
  /** Current state of the audio system */
  readonly state: AudioStateType;
  /** Error message if state is ERROR */
  readonly errorMessage?: string;
  /** Sample rate in Hz (typically 48000) */
  readonly sampleRate: number;
  /** Buffer size in samples (typically 2048) */
  readonly bufferSize: number;
  /** Latest pitch detection result */
  readonly latestPitch: PitchResult;
}

/** Default audio configuration */
export const DEFAULT_AUDIO_CONFIG = {
  SAMPLE_RATE: 48000,
  BUFFER_SIZE: 2048,
} as const;

/** Create initial audio state */
export function createInitialAudioState(): AudioState {
  return {
    state: AUDIO_STATE.UNINITIALIZED,
    sampleRate: DEFAULT_AUDIO_CONFIG.SAMPLE_RATE,
    bufferSize: DEFAULT_AUDIO_CONFIG.BUFFER_SIZE,
    latestPitch: createEmptyPitchResult(),
  };
}

// ============================================================================
// Note Conversion Utilities
// ============================================================================

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

/**
 * Convert a frequency to note name, octave, and cents deviation.
 * Uses A4 = 440 Hz as reference.
 */
export function frequencyToNote(freq: number): {
  note: string;
  octave: number;
  cents: number;
} {
  if (freq < 20 || freq > 5000) {
    return { note: '', octave: 0, cents: 0 };
  }

  // MIDI note number: 69 = A4 = 440 Hz
  const midiNote = 12 * Math.log2(freq / 440) + 69;
  const noteNum = Math.round(midiNote);
  const cents = Math.round((midiNote - noteNum) * 100);

  // Handle negative modulo correctly
  const noteIndex = ((noteNum % 12) + 12) % 12;
  const note = NOTE_NAMES[noteIndex] ?? '';
  const octave = Math.floor(noteNum / 12) - 1;

  return { note, octave, cents };
}

/**
 * Convert a note name and octave to frequency in Hz.
 * Uses A4 = 440 Hz as reference.
 */
export function noteToFrequency(note: string, octave: number): number {
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Required to check membership in readonly tuple
  const noteIndex = NOTE_NAMES.indexOf(note as (typeof NOTE_NAMES)[number]);
  if (noteIndex === -1) {
    return 0;
  }

  // MIDI note number
  const midiNote = (octave + 1) * 12 + noteIndex;
  return 440 * Math.pow(2, (midiNote - 69) / 12);
}
