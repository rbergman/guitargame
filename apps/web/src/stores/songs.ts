/**
 * Alpine.js store for song/exercise data management.
 *
 * Handles loading exercises from JSON, selecting songs, and providing
 * note access methods for the game loop.
 */

import Alpine from 'alpinejs';
import type { BPM, Seconds } from '../types/branded.js';
import { bpm, seconds } from '../types/branded.js';
import type { Song, SongDefinition, TabNote, TabNoteDefinition, Tuning } from '../types/song.js';
import { createTabNote, createTabNoteState, TUNINGS_BY_NAME, TUNINGS } from '../types/song.js';

// ============================================================================
// JSON Schema Types (what we receive from exercises.json)
// ============================================================================

/** Raw note data from JSON (beat-based, not time-based) */
interface RawNote {
  beat: number;
  string: number;
  fret: number;
  duration?: number;
}

/** Raw exercise data from JSON */
interface RawExercise {
  id?: string;
  title: string;
  artist: string;
  bpm: number;
  tuning?: string;
  notes: RawNote[];
}

// ============================================================================
// Store State Interface
// ============================================================================

export interface SongsStore {
  /** Loaded exercises */
  exercises: Song[];
  /** Currently selected song for gameplay */
  currentSong: Song | null;
  /** Index of selected song in exercises array */
  selectedIndex: number;
  /** Whether exercises are currently loading */
  loading: boolean;
  /** Error message if loading failed */
  error: string | null;

  // Methods
  loadExercises(url: string): Promise<void>;
  selectSong(index: number): void;
  getNotesInRange(startTime: number, endTime: number): TabNote[];
  nextUnhitNote(): TabNote | null;
  resetNoteStates(): void;
}

// Type augmentation for Alpine
declare module 'alpinejs' {
  interface Stores {
    songs: SongsStore;
  }
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Convert beat to time in seconds using BPM.
 * Formula: time = beat * (60 / bpm)
 */
function beatToTime(beat: number, songBpm: BPM): Seconds {
  return seconds(beat * (60 / songBpm));
}

/**
 * Resolve tuning string to Tuning array.
 * Supports named tunings like "standard", "drop-d", etc.
 */
function resolveTuning(tuningStr: string | undefined): Tuning {
  if (!tuningStr) {
    return TUNINGS.STANDARD;
  }

  const normalized = tuningStr.toLowerCase().trim();
  const namedTuning = TUNINGS_BY_NAME[normalized];

  if (namedTuning) {
    return namedTuning;
  }

  // Default to standard if unknown
  return TUNINGS.STANDARD;
}

/**
 * Validate and convert string index from JSON.
 * Valid values: 0 (G), 1 (D), 2 (A), 3 (E), 4 (B for 5-string).
 */
function validateStringIndex(value: number): 0 | 1 | 2 | 3 | 4 {
  if (value >= 0 && value <= 4) {
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- JSON validation: we've verified the value is in valid range
    return value as 0 | 1 | 2 | 3 | 4;
  }
  throw new RangeError(`Invalid string index: ${String(value)}, expected 0-4`);
}

/**
 * Convert raw exercise data from JSON to Song format.
 */
function rawExerciseToSong(raw: RawExercise): Song {
  const songBpm = bpm(raw.bpm);
  const tuning = resolveTuning(raw.tuning);

  // Convert beat-based notes to time-based TabNotes
  const noteDefinitions: TabNoteDefinition[] = raw.notes.map((note) => ({
    time: beatToTime(note.beat, songBpm),
    beat: note.beat,
    string: validateStringIndex(note.string),
    fret: note.fret,
    duration: seconds(note.duration ?? 0.5),
  }));

  // Create TabNotes with runtime state
  const notes: TabNote[] = noteDefinitions.map(createTabNote);

  // Calculate song duration (last note time + some buffer)
  const lastNoteTime = notes.length > 0 ? Math.max(...notes.map((n) => n.time + n.duration)) : 0;
  const duration = seconds(lastNoteTime + 2); // 2 second buffer after last note

  const definition: SongDefinition = {
    title: raw.title,
    artist: raw.artist,
    bpm: songBpm,
    tuningStr: raw.tuning,
    notes: noteDefinitions,
  };

  return {
    definition,
    tuning,
    duration,
    notes,
  };
}

// ============================================================================
// Store Factory
// ============================================================================

function createSongsStore(): SongsStore {
  return {
    // State
    exercises: [],
    currentSong: null,
    selectedIndex: -1,
    loading: false,
    error: null,

    // Methods
    async loadExercises(url: string): Promise<void> {
      this.loading = true;
      this.error = null;

      try {
        const response = await fetch(url);

        if (!response.ok) {
          throw new Error(`Failed to load exercises: ${String(response.status)} ${response.statusText}`);
        }

        const data: unknown = await response.json();

        // Validate that we received an array
        if (!Array.isArray(data)) {
          throw new Error('Invalid exercises data: expected an array');
        }

        // Convert raw exercises to Song format
        // After Array.isArray check, data is unknown[], each element validated by rawExerciseToSong
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- JSON parsing: Array.isArray narrows to unknown[], we trust JSON structure matches RawExercise
        this.exercises = (data as RawExercise[]).map(rawExerciseToSong);
        this.loading = false;
      } catch (err) {
        this.loading = false;
        this.error = err instanceof Error ? err.message : 'Unknown error loading exercises';
        console.error('Failed to load exercises:', err);
      }
    },

    selectSong(index: number): void {
      if (index < 0 || index >= this.exercises.length) {
        this.currentSong = null;
        this.selectedIndex = -1;
        return;
      }

      this.selectedIndex = index;
      const song = this.exercises[index];

      if (!song) {
        this.currentSong = null;
        return;
      }

      // Deep copy the song to avoid mutating the exercise template
      // Reset note states for fresh gameplay
      this.currentSong = {
        ...song,
        notes: song.notes.map((note: TabNote) => ({
          ...note,
          state: createTabNoteState(),
        })),
      };
    },

    getNotesInRange(startTime: number, endTime: number): TabNote[] {
      if (!this.currentSong) {
        return [];
      }

      return this.currentSong.notes.filter(
        (note: TabNote) => note.time >= startTime && note.time <= endTime
      );
    },

    nextUnhitNote(): TabNote | null {
      if (!this.currentSong) {
        return null;
      }

      // Find first note that hasn't been hit, sorted by time
      const unhitNotes = this.currentSong.notes
        .filter((note: TabNote) => !note.state.hit)
        .sort((a: TabNote, b: TabNote) => a.time - b.time);

      return unhitNotes[0] ?? null;
    },

    resetNoteStates(): void {
      if (!this.currentSong) {
        return;
      }

      for (const note of this.currentSong.notes) {
        note.state = createTabNoteState();
      }
    },
  };
}

// ============================================================================
// Store Registration
// ============================================================================

/**
 * Register the songs store with Alpine.js.
 * Must be called before Alpine.start().
 */
export function registerSongsStore(): void {
  Alpine.store('songs', createSongsStore());
}
