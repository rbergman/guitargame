package song

import (
	"strings"
	"time"
)

// Bass guitar strings (standard tuning)
const (
	StringG = 0 // G string (highest, thinnest)
	StringD = 1 // D string
	StringA = 2 // A string
	StringE = 3 // E string (lowest, thickest)
)

// StringTuning represents the tuning of a single string
type StringTuning struct {
	Note   string // Note name (C, C#, D, etc.)
	Octave int    // Octave number
}

// Semitone returns the semitone value (0-11) for this note
func (s StringTuning) Semitone() int {
	noteMap := map[string]int{
		"C": 0, "C#": 1, "Db": 1,
		"D": 2, "D#": 3, "Eb": 3,
		"E": 4, "Fb": 4,
		"F": 5, "F#": 6, "Gb": 6,
		"G": 7, "G#": 8, "Ab": 8,
		"A": 9, "A#": 10, "Bb": 10,
		"B": 11, "Cb": 11,
	}
	return noteMap[s.Note]
}

// Tuning represents the tuning for all strings (high to low)
type Tuning []StringTuning

// Predefined tunings
var (
	// TuningStandard is standard 4-string bass tuning (G-D-A-E)
	TuningStandard = Tuning{
		{Note: "G", Octave: 2},
		{Note: "D", Octave: 2},
		{Note: "A", Octave: 1},
		{Note: "E", Octave: 1},
	}

	// TuningDropD is drop D tuning (G-D-A-D)
	TuningDropD = Tuning{
		{Note: "G", Octave: 2},
		{Note: "D", Octave: 2},
		{Note: "A", Octave: 1},
		{Note: "D", Octave: 1},
	}

	// TuningHalfStepDown is half step down tuning (Gb-Db-Ab-Eb)
	TuningHalfStepDown = Tuning{
		{Note: "Gb", Octave: 2},
		{Note: "Db", Octave: 2},
		{Note: "Ab", Octave: 1},
		{Note: "Eb", Octave: 1},
	}

	// TuningFullStepDown is full step down tuning (F-C-G-D)
	TuningFullStepDown = Tuning{
		{Note: "F", Octave: 2},
		{Note: "C", Octave: 2},
		{Note: "G", Octave: 1},
		{Note: "D", Octave: 1},
	}

	// Tuning5StringStandard is standard 5-string bass tuning (G-D-A-E-B)
	Tuning5StringStandard = Tuning{
		{Note: "G", Octave: 2},
		{Note: "D", Octave: 2},
		{Note: "A", Octave: 1},
		{Note: "E", Octave: 1},
		{Note: "B", Octave: 0},
	}

	// TuningsByName maps tuning names to tuning values
	TuningsByName = map[string]Tuning{
		"standard":       TuningStandard,
		"drop-d":         TuningDropD,
		"half-step-down": TuningHalfStepDown,
		"full-step-down": TuningFullStepDown,
		"5-string":       Tuning5StringStandard,
	}
)

// ParseTuning parses a tuning from a string or slice
// Accepts: "standard", "drop-d", or custom like "G2,D2,A1,D1"
func ParseTuning(s string) Tuning {
	s = strings.ToLower(strings.TrimSpace(s))

	// Check predefined tunings
	if t, ok := TuningsByName[s]; ok {
		return t
	}

	// Try to parse custom tuning (e.g., "G2,D2,A1,E1")
	parts := strings.Split(s, ",")
	if len(parts) >= 4 {
		tuning := make(Tuning, len(parts))
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if len(part) >= 2 {
				// Last character is octave
				octave := int(part[len(part)-1] - '0')
				note := strings.ToUpper(part[:len(part)-1])
				tuning[i] = StringTuning{Note: note, Octave: octave}
			}
		}
		return tuning
	}

	// Default to standard
	return TuningStandard
}

// Hit quality levels
type HitQuality int

const (
	HitMiss HitQuality = iota
	HitOK
	HitGood
	HitPerfect
)

func (h HitQuality) String() string {
	switch h {
	case HitPerfect:
		return "Perfect!"
	case HitGood:
		return "Good"
	case HitOK:
		return "OK"
	default:
		return "Miss"
	}
}

func (h HitQuality) Score() int {
	switch h {
	case HitPerfect:
		return 100
	case HitGood:
		return 50
	case HitOK:
		return 25
	default:
		return 0
	}
}

// TabNote represents a single note in tablature
type TabNote struct {
	Time     float64    `yaml:"time"`     // Time in seconds from song start
	Beat     float64    `yaml:"beat"`     // Beat number (converted to time using BPM)
	String   int        `yaml:"string"`   // 0=G, 1=D, 2=A, 3=E
	Fret     int        `yaml:"fret"`     // Fret number (0 = open string)
	Duration float64    `yaml:"duration"` // Note duration in seconds (optional)

	// Runtime state (not serialized)
	Hit        bool       `yaml:"-"`
	HitQuality HitQuality `yaml:"-"`
	HitTime    float64    `yaml:"-"`
}

// NoteWithTuning returns the note name for this tab position using the given tuning
func (n *TabNote) NoteWithTuning(tuning Tuning) string {
	notes := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

	if n.String >= len(tuning) {
		return "?"
	}

	openString := tuning[n.String]
	baseNote := openString.Semitone()

	if n.Fret == 0 {
		return openString.Note
	}

	noteIndex := (baseNote + n.Fret) % 12
	return notes[noteIndex]
}

// OctaveWithTuning returns the octave for this tab position using the given tuning
func (n *TabNote) OctaveWithTuning(tuning Tuning) int {
	if n.String >= len(tuning) {
		return 1
	}

	openString := tuning[n.String]
	baseOctave := openString.Octave
	baseNote := openString.Semitone()

	notePos := baseNote + n.Fret

	// Add octaves for every 12 semitones
	return baseOctave + notePos/12
}

// Note returns the note name using standard tuning (for backwards compatibility)
func (n *TabNote) Note() string {
	return n.NoteWithTuning(TuningStandard)
}

// Octave returns the octave using standard tuning (for backwards compatibility)
func (n *TabNote) Octave() int {
	return n.OctaveWithTuning(TuningStandard)
}

// Song represents a complete song with tablature
type Song struct {
	Title     string    `yaml:"title"`
	Artist    string    `yaml:"artist"`
	BPM       float64   `yaml:"bpm"`
	TuningStr string    `yaml:"tuning"` // Tuning name or custom (e.g., "standard", "drop-d", "G2,D2,A1,D1")
	Notes     []TabNote `yaml:"notes"`

	// Runtime state
	Duration float64 `yaml:"-"`
	Tuning   Tuning  `yaml:"-"` // Parsed tuning (set during load)
}

// GetTuning returns the song's tuning, defaulting to standard if not set
func (s *Song) GetTuning() Tuning {
	if s.Tuning != nil {
		return s.Tuning
	}
	if s.TuningStr != "" {
		return ParseTuning(s.TuningStr)
	}
	return TuningStandard
}

// NoteAt returns the note name for a given TabNote using this song's tuning
func (s *Song) NoteAt(note *TabNote) string {
	return note.NoteWithTuning(s.GetTuning())
}

// OctaveAt returns the octave for a given TabNote using this song's tuning
func (s *Song) OctaveAt(note *TabNote) int {
	return note.OctaveWithTuning(s.GetTuning())
}

// NoteAtTime returns notes that should be played at the given time
func (s *Song) NotesInRange(startTime, endTime float64) []*TabNote {
	var notes []*TabNote
	for i := range s.Notes {
		if s.Notes[i].Time >= startTime && s.Notes[i].Time <= endTime {
			notes = append(notes, &s.Notes[i])
		}
	}
	return notes
}

// NextUnhitNote returns the next note that hasn't been hit yet
func (s *Song) NextUnhitNote(currentTime float64) *TabNote {
	for i := range s.Notes {
		if !s.Notes[i].Hit && s.Notes[i].Time >= currentTime-0.5 {
			return &s.Notes[i]
		}
	}
	return nil
}

// CalculateDuration sets the song duration based on the last note
func (s *Song) CalculateDuration() {
	if len(s.Notes) == 0 {
		s.Duration = 0
		return
	}
	lastNote := s.Notes[len(s.Notes)-1]
	s.Duration = lastNote.Time + lastNote.Duration + 2.0 // 2 second buffer
}

// GameState holds the current game state
type GameState struct {
	Song         *Song
	StartTime    time.Time
	CurrentTime  float64
	Score        int
	Combo        int
	MaxCombo     int
	NotesHit     int
	NotesMissed  int
	TotalNotes   int
	IsPlaying    bool
	IsFinished   bool
	FloatingText []FloatingScore
}

// FloatingScore represents floating score text
type FloatingScore struct {
	Text      string
	X, Y      float32
	StartTime time.Time
	Quality   HitQuality
}

// NewGameState creates a new game state for a song
func NewGameState(song *Song) *GameState {
	song.CalculateDuration()
	return &GameState{
		Song:         song,
		TotalNotes:   len(song.Notes),
		FloatingText: make([]FloatingScore, 0),
	}
}

// Start begins the game
func (g *GameState) Start() {
	g.StartTime = time.Now()
	g.IsPlaying = true
	g.IsFinished = false
}

// Update updates the game state
func (g *GameState) Update() {
	if !g.IsPlaying {
		return
	}

	g.CurrentTime = time.Since(g.StartTime).Seconds()

	// Check for finished
	if g.CurrentTime > g.Song.Duration {
		g.IsPlaying = false
		g.IsFinished = true
	}

	// Clean up old floating text (fade after 1 second)
	newFloating := make([]FloatingScore, 0)
	for _, f := range g.FloatingText {
		if time.Since(f.StartTime).Seconds() < 1.0 {
			newFloating = append(newFloating, f)
		}
	}
	g.FloatingText = newFloating
}

// RegisterHit records a note hit
func (g *GameState) RegisterHit(note *TabNote, quality HitQuality, x, y float32) {
	note.Hit = true
	note.HitQuality = quality
	note.HitTime = g.CurrentTime

	points := quality.Score()

	if quality != HitMiss {
		g.Combo++
		if g.Combo > g.MaxCombo {
			g.MaxCombo = g.Combo
		}
		// Combo multiplier
		multiplier := 1
		if g.Combo >= 10 {
			multiplier = 2
		}
		if g.Combo >= 25 {
			multiplier = 3
		}
		if g.Combo >= 50 {
			multiplier = 4
		}
		points *= multiplier
		g.NotesHit++
	} else {
		g.Combo = 0
		g.NotesMissed++
	}

	g.Score += points

	// Add floating text
	text := quality.String()
	if points > 0 && g.Combo > 1 {
		text = quality.String() + " x" + string(rune('0'+min(g.Combo, 9)))
	}
	g.FloatingText = append(g.FloatingText, FloatingScore{
		Text:      text,
		X:         x,
		Y:         y,
		StartTime: time.Now(),
		Quality:   quality,
	})
}

// Accuracy returns the hit accuracy as a percentage
func (g *GameState) Accuracy() float64 {
	total := g.NotesHit + g.NotesMissed
	if total == 0 {
		return 100.0
	}
	return float64(g.NotesHit) / float64(total) * 100.0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
