package audio

import (
	"math"

	aubio "github.com/coral/aubio-go"
)

var noteNames = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

type PitchResult struct {
	Frequency  float64
	Confidence float64
	Note       string
	Octave     int
	Cents      int
	RMS        float64
}

type PitchDetector struct {
	detector   *aubio.Pitch
	sampleRate float64
	bufferSize int
}

func NewPitchDetector(bufferSize int, sampleRate float64) *PitchDetector {
	hopSize := bufferSize / 2

	detector := aubio.NewPitch(aubio.PitchYin, uint(bufferSize), uint(hopSize), uint(sampleRate))
	detector.SetUnit(aubio.PitchOutFreq)
	detector.SetTolerance(0.8)

	return &PitchDetector{
		detector:   detector,
		sampleRate: sampleRate,
		bufferSize: bufferSize,
	}
}

func (p *PitchDetector) Detect(samples []float32) PitchResult {
	data := make([]float64, len(samples))
	for i, s := range samples {
		data[i] = float64(s)
	}

	buf := aubio.NewSimpleBufferData(uint(len(samples)), data)
	defer buf.Free()

	p.detector.Do(buf)
	outBuf := p.detector.Buffer()

	freq := 0.0
	if outBuf != nil && outBuf.Size() > 0 {
		slice := outBuf.Slice()
		if len(slice) > 0 {
			freq = slice[0]
		}
	}

	rms := computeRMS(samples)
	conf := computeConfidence(freq, rms)

	note, octave, cents := frequencyToNote(freq)

	return PitchResult{
		Frequency:  freq,
		Confidence: conf,
		Note:       note,
		Octave:     octave,
		Cents:      cents,
		RMS:        rms,
	}
}

func computeRMS(samples []float32) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func computeConfidence(freq, rms float64) float64 {
	if freq < 20 || freq > 500 {
		return 0
	}
	if rms < 0.001 {
		return 0
	}
	conf := math.Min(rms*50, 1.0)
	return conf
}

func (p *PitchDetector) Close() {
	if p.detector != nil {
		p.detector.Free()
	}
}

func frequencyToNote(freq float64) (string, int, int) {
	if freq < 20 || freq > 5000 {
		return "", 0, 0
	}

	midiNote := 12*math.Log2(freq/440) + 69
	noteNum := int(math.Round(midiNote))
	cents := int((midiNote - float64(noteNum)) * 100)

	name := noteNames[((noteNum%12)+12)%12]
	octave := (noteNum / 12) - 1

	return name, octave, cents
}

func NoteToFrequency(note string, octave int) float64 {
	noteIndex := -1
	for i, n := range noteNames {
		if n == note {
			noteIndex = i
			break
		}
	}
	if noteIndex == -1 {
		return 0
	}

	midiNote := (octave+1)*12 + noteIndex
	return 440 * math.Pow(2, float64(midiNote-69)/12)
}

func (r PitchResult) NoteName() string {
	if r.Note == "" {
		return "--"
	}
	return r.Note
}

func (r PitchResult) FullNoteName() string {
	if r.Note == "" {
		return "--"
	}
	return r.Note + string(rune('0'+r.Octave))
}

func (r PitchResult) IsValid() bool {
	return r.Note != "" && r.Confidence > 0.5
}
