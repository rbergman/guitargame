/**
 * Branded types for domain safety.
 *
 * These types prevent accidental mixing of semantically different numeric values
 * (e.g., passing a BPM where a Frequency is expected).
 */

declare const brand: unique symbol;

type Brand<T, B> = T & { readonly [brand]: B };

/** Frequency in Hertz (Hz) - audible range 20-20000 Hz */
export type Frequency = Brand<number, 'Frequency'>;

/** Pitch deviation in cents (1/100th of a semitone) */
export type Cents = Brand<number, 'Cents'>;

/** Tempo in beats per minute */
export type BPM = Brand<number, 'BPM'>;

/** Time duration in milliseconds */
export type Milliseconds = Brand<number, 'Milliseconds'>;

/** Time duration in seconds */
export type Seconds = Brand<number, 'Seconds'>;

// ============================================================================
// Constructor functions with validation
// ============================================================================

/**
 * Creates a validated Frequency value.
 * @param hz - Frequency in Hertz
 * @throws RangeError if frequency is outside audible range (20-20000 Hz)
 */
export function frequency(hz: number): Frequency {
  if (hz < 20 || hz > 20000) {
    throw new RangeError(`Frequency ${String(hz)}Hz out of audible range (20-20000)`);
  }
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type constructor: this is the only valid way to create a Frequency value
  return hz as Frequency;
}

/**
 * Creates a Cents value for pitch deviation.
 * @param value - Cents deviation from perfect pitch (-50 to +50 typical)
 */
export function cents(value: number): Cents {
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type constructor: this is the only valid way to create a Cents value
  return value as Cents;
}

/**
 * Creates a validated BPM value.
 * @param value - Beats per minute
 * @throws RangeError if BPM is outside reasonable range (20-300)
 */
export function bpm(value: number): BPM {
  if (value < 20 || value > 300) {
    throw new RangeError(`BPM ${String(value)} out of reasonable range (20-300)`);
  }
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type constructor: this is the only valid way to create a BPM value
  return value as BPM;
}

/**
 * Creates a Milliseconds value.
 * @param value - Time in milliseconds
 * @throws RangeError if value is negative
 */
export function milliseconds(value: number): Milliseconds {
  if (value < 0) {
    throw new RangeError(`Milliseconds cannot be negative: ${String(value)}`);
  }
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type constructor: this is the only valid way to create a Milliseconds value
  return value as Milliseconds;
}

/**
 * Creates a Seconds value.
 * @param value - Time in seconds
 * @throws RangeError if value is negative
 */
export function seconds(value: number): Seconds {
  if (value < 0) {
    throw new RangeError(`Seconds cannot be negative: ${String(value)}`);
  }
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type constructor: this is the only valid way to create a Seconds value
  return value as Seconds;
}

// ============================================================================
// Conversion utilities
// ============================================================================

/** Convert Milliseconds to Seconds */
export function msToSeconds(ms: Milliseconds): Seconds {
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type conversion: arithmetic on branded types requires re-branding
  return (ms / 1000) as Seconds;
}

/** Convert Seconds to Milliseconds */
export function secondsToMs(s: Seconds): Milliseconds {
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions -- Branded type conversion: arithmetic on branded types requires re-branding
  return (s * 1000) as Milliseconds;
}
