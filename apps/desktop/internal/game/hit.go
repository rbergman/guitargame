package game

import (
	"math"

	"guitargame/apps/desktop/internal/audio"
	"guitargame/apps/desktop/internal/song"
)

// Timing windows in seconds
const (
	PerfectWindow = 0.050 // ±50ms
	GoodWindow    = 0.100 // ±100ms
	OKWindow      = 0.150 // ±150ms
	MissWindow    = 0.300 // After 300ms, note is missed
)

// HitDetector handles matching played notes to expected notes
type HitDetector struct {
	state *song.GameState

	// Note frequencies for matching (bass range)
	noteFrequencies map[string]float64
}

// NewHitDetector creates a new hit detector
func NewHitDetector(state *song.GameState) *HitDetector {
	return &HitDetector{
		state:           state,
		noteFrequencies: buildNoteFrequencies(),
	}
}

// buildNoteFrequencies creates a map of note names to frequencies
func buildNoteFrequencies() map[string]float64 {
	notes := make(map[string]float64)
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	// Also support flat names
	flatNames := map[string]string{"Db": "C#", "Eb": "D#", "Fb": "E", "Gb": "F#", "Ab": "G#", "Bb": "A#", "Cb": "B"}

	// Generate frequencies for octaves 0-4 (extended bass range for 5-string)
	for octave := 0; octave <= 4; octave++ {
		for i, name := range noteNames {
			midiNote := (octave+1)*12 + i
			freq := 440.0 * math.Pow(2, float64(midiNote-69)/12.0)
			key := name + string(rune('0'+octave))
			notes[key] = freq
		}
	}

	// Add flat note aliases
	for flat, sharp := range flatNames {
		for octave := 0; octave <= 4; octave++ {
			sharpKey := sharp + string(rune('0'+octave))
			flatKey := flat + string(rune('0'+octave))
			if freq, ok := notes[sharpKey]; ok {
				notes[flatKey] = freq
			}
		}
	}

	return notes
}

// CheckHit checks if the detected pitch matches any pending note
func (h *HitDetector) CheckHit(pitch audio.PitchResult, playLineX float32) {
	if !pitch.IsValid() {
		return
	}

	currentTime := h.state.CurrentTime

	// Find notes within the hit window
	for i := range h.state.Song.Notes {
		note := &h.state.Song.Notes[i]

		// Skip already hit notes
		if note.Hit {
			continue
		}

		// Check if note is within timing window
		timeDiff := note.Time - currentTime
		absTimeDiff := math.Abs(timeDiff)

		// Note is too far in the future
		if timeDiff > MissWindow {
			continue
		}

		// Note was missed (too far in the past)
		if timeDiff < -MissWindow {
			// Mark as missed
			h.state.RegisterHit(note, song.HitMiss, playLineX, float32(80+note.String*40))
			continue
		}

		// Check if the played note matches
		if h.notesMatch(pitch, note) {
			quality := h.getHitQuality(absTimeDiff)
			h.state.RegisterHit(note, quality, playLineX, float32(80+note.String*40))
			return // Only hit one note per detection
		}
	}
}

// notesMatch checks if the detected pitch matches the expected note
func (h *HitDetector) notesMatch(pitch audio.PitchResult, note *song.TabNote) bool {
	// Use the song's tuning to determine the expected note
	expectedNote := h.state.Song.NoteAt(note)
	expectedOctave := h.state.Song.OctaveAt(note)
	expectedKey := expectedNote + string(rune('0'+expectedOctave))

	expectedFreq, ok := h.noteFrequencies[expectedKey]
	if !ok {
		return false
	}

	// Allow some tolerance in frequency matching
	// Use cents - 100 cents = 1 semitone
	// We'll allow ±50 cents (half a semitone)
	centsDiff := 1200 * math.Log2(pitch.Frequency/expectedFreq)

	return math.Abs(centsDiff) < 50
}

// getHitQuality determines hit quality based on timing
func (h *HitDetector) getHitQuality(absTimeDiff float64) song.HitQuality {
	if absTimeDiff <= PerfectWindow {
		return song.HitPerfect
	}
	if absTimeDiff <= GoodWindow {
		return song.HitGood
	}
	if absTimeDiff <= OKWindow {
		return song.HitOK
	}
	return song.HitMiss
}

// Update checks for missed notes
func (h *HitDetector) Update() {
	currentTime := h.state.CurrentTime

	for i := range h.state.Song.Notes {
		note := &h.state.Song.Notes[i]

		if note.Hit {
			continue
		}

		// Check if note was missed
		if currentTime-note.Time > MissWindow {
			h.state.RegisterHit(note, song.HitMiss, 0, float32(80+note.String*40))
		}
	}
}

// GetExpectedNote returns the next note the player should play
func (h *HitDetector) GetExpectedNote() *song.TabNote {
	return h.state.Song.NextUnhitNote(h.state.CurrentTime)
}
