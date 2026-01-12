package render

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"guitargame/apps/desktop/internal/song"
)

// TabRenderer renders scrolling bass tablature
type TabRenderer struct {
	theme *material.Theme

	// Layout constants
	StringSpacing  float32
	PlayLineX      float32 // X position of the "now" line (right side)
	PixelsPerBeat  float32 // How many pixels per beat
	TabAreaHeight  float32
	TabAreaPadding float32
}

// NewTabRenderer creates a new tab renderer
func NewTabRenderer(theme *material.Theme) *TabRenderer {
	return &TabRenderer{
		theme:          theme,
		StringSpacing:  40,
		PlayLineX:      0.75, // 75% from left
		PixelsPerBeat:  80,
		TabAreaHeight:  200,
		TabAreaPadding: 20,
	}
}

// Colors - high contrast for readability
var (
	ColorBackground  = color.NRGBA{R: 20, G: 20, B: 30, A: 255}
	ColorString      = color.NRGBA{R: 140, G: 140, B: 160, A: 255}  // Brighter strings
	ColorPlayLine    = color.NRGBA{R: 100, G: 220, B: 255, A: 255}  // Brighter play line
	ColorNoteDefault = color.NRGBA{R: 255, G: 255, B: 255, A: 255}  // White notes
	ColorNotePerfect = color.NRGBA{R: 50, G: 255, B: 100, A: 255}   // Bright green
	ColorNoteGood    = color.NRGBA{R: 180, G: 255, B: 50, A: 255}   // Yellow-green
	ColorNoteOK      = color.NRGBA{R: 255, G: 220, B: 50, A: 255}   // Yellow
	ColorNoteMiss    = color.NRGBA{R: 255, G: 80, B: 80, A: 255}    // Red
	ColorFloatText   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

// StringNames for bass guitar
var StringNames = []string{"G", "D", "A", "E"}

// Layout renders the complete tab view
func (r *TabRenderer) Layout(gtx layout.Context, state *song.GameState) layout.Dimensions {
	width := float32(gtx.Constraints.Max.X)
	height := float32(gtx.Constraints.Max.Y)

	// Calculate play line position
	playLineX := width * r.PlayLineX

	// Calculate pixels per second based on BPM
	beatsPerSecond := state.Song.BPM / 60.0
	pixelsPerSecond := r.PixelsPerBeat * float32(beatsPerSecond)

	// Draw background
	r.drawBackground(gtx, int(width), int(height))

	// Calculate tab area bounds
	tabTop := r.TabAreaPadding + 60 // Leave room for header
	tabHeight := r.StringSpacing * 5 // 4 strings + padding

	// Draw string lines
	r.drawStrings(gtx, int(width), tabTop, tabHeight)

	// Draw play line (the "now" indicator)
	r.drawPlayLine(gtx, playLineX, tabTop, tabHeight)

	// Draw notes
	r.drawNotes(gtx, state, playLineX, tabTop, pixelsPerSecond)

	// Draw floating score text
	r.drawFloatingText(gtx, state)

	// Draw string labels on left
	r.drawStringLabels(gtx, tabTop)

	return layout.Dimensions{Size: image.Pt(int(width), int(height))}
}

func (r *TabRenderer) drawBackground(gtx layout.Context, width, height int) {
	defer clip.Rect{Max: image.Pt(width, height)}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: ColorBackground}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (r *TabRenderer) drawStrings(gtx layout.Context, width int, tabTop, tabHeight float32) {
	for i := 0; i < 4; i++ {
		y := int(tabTop + float32(i)*r.StringSpacing + r.StringSpacing/2)

		defer clip.Rect{
			Min: image.Pt(50, y),
			Max: image.Pt(width-10, y+2),
		}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: ColorString}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	}
}

func (r *TabRenderer) drawPlayLine(gtx layout.Context, x, tabTop, tabHeight float32) {
	// Vertical play line
	defer clip.Rect{
		Min: image.Pt(int(x)-1, int(tabTop)-10),
		Max: image.Pt(int(x)+2, int(tabTop+tabHeight)),
	}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: ColorPlayLine}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	// Glow effect (wider, more transparent)
	defer clip.Rect{
		Min: image.Pt(int(x)-4, int(tabTop)-10),
		Max: image.Pt(int(x)+5, int(tabTop+tabHeight)),
	}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color.NRGBA{R: 100, G: 200, B: 255, A: 50}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (r *TabRenderer) drawNotes(gtx layout.Context, state *song.GameState, playLineX, tabTop, pixelsPerSecond float32) {
	currentTime := state.CurrentTime

	// Calculate visible time range
	// Notes to the right of play line are in the future
	// Notes to the left have already passed
	timeAtLeft := currentTime - float64(playLineX/pixelsPerSecond)
	timeAtRight := currentTime + float64((float32(gtx.Constraints.Max.X)-playLineX)/pixelsPerSecond)

	for i := range state.Song.Notes {
		note := &state.Song.Notes[i]

		// Skip notes outside visible range
		if note.Time < timeAtLeft-1 || note.Time > timeAtRight+1 {
			continue
		}

		// Calculate X position
		timeDelta := note.Time - currentTime
		noteX := playLineX + float32(timeDelta)*pixelsPerSecond

		// Calculate Y position based on string
		noteY := tabTop + float32(note.String)*r.StringSpacing + r.StringSpacing/2

		// Determine note color based on state
		noteColor := ColorNoteDefault
		if note.Hit {
			switch note.HitQuality {
			case song.HitPerfect:
				noteColor = ColorNotePerfect
			case song.HitGood:
				noteColor = ColorNoteGood
			case song.HitOK:
				noteColor = ColorNoteOK
			case song.HitMiss:
				noteColor = ColorNoteMiss
			}
		}

		// Draw note background circle
		r.drawNoteCircle(gtx, noteX, noteY, 18, noteColor)

		// Draw fret number
		r.drawFretNumber(gtx, noteX, noteY, note.Fret)
	}
}

func (r *TabRenderer) drawNoteCircle(gtx layout.Context, x, y, radius float32, c color.NRGBA) {
	// Draw filled circle for note
	center := image.Pt(int(x), int(y))
	bounds := image.Rect(
		center.X-int(radius),
		center.Y-int(radius),
		center.X+int(radius),
		center.Y+int(radius),
	)

	defer clip.Ellipse{Min: bounds.Min, Max: bounds.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: c}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (r *TabRenderer) drawFretNumber(gtx layout.Context, x, y float32, fret int) {
	// Position text centered on the note
	offset := op.Offset(image.Pt(int(x)-8, int(y)-10)).Push(gtx.Ops)

	label := material.Body1(r.theme, fmt.Sprintf("%d", fret))
	label.Color = color.NRGBA{R: 30, G: 30, B: 40, A: 255}
	label.Alignment = text.Middle
	label.Layout(gtx)

	offset.Pop()
}

func (r *TabRenderer) drawStringLabels(gtx layout.Context, tabTop float32) {
	for i, name := range StringNames {
		y := tabTop + float32(i)*r.StringSpacing + r.StringSpacing/2 - 10

		offset := op.Offset(image.Pt(15, int(y))).Push(gtx.Ops)

		label := material.Body1(r.theme, name)
		label.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
		label.Layout(gtx)

		offset.Pop()
	}
}

func (r *TabRenderer) drawFloatingText(gtx layout.Context, state *song.GameState) {
	now := time.Now()

	for _, ft := range state.FloatingText {
		elapsed := now.Sub(ft.StartTime).Seconds()
		if elapsed > 1.0 {
			continue
		}

		// Float upward and fade out
		yOffset := float32(elapsed * 50) // Move up 50 pixels over 1 second
		alpha := uint8(255 * (1 - elapsed))

		// Choose color based on quality
		var textColor color.NRGBA
		switch ft.Quality {
		case song.HitPerfect:
			textColor = color.NRGBA{R: 100, G: 255, B: 100, A: alpha}
		case song.HitGood:
			textColor = color.NRGBA{R: 200, G: 255, B: 100, A: alpha}
		case song.HitOK:
			textColor = color.NRGBA{R: 255, G: 255, B: 100, A: alpha}
		default:
			textColor = color.NRGBA{R: 255, G: 100, B: 100, A: alpha}
		}

		offset := op.Offset(image.Pt(int(ft.X)-20, int(ft.Y-yOffset))).Push(gtx.Ops)

		label := material.H6(r.theme, ft.Text)
		label.Color = textColor
		label.Layout(gtx)

		offset.Pop()
	}
}

// DrawHeader renders the score and status header
func (r *TabRenderer) DrawHeader(gtx layout.Context, state *song.GameState) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Song title
			inset := layout.Inset{Left: unit.Dp(10), Top: unit.Dp(10)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.H6(r.theme, state.Song.Title)
				label.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
				return label.Layout(gtx)
			})
		}),
		layout.Flexed(1, layout.Spacer{}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Score
			inset := layout.Inset{Right: unit.Dp(20), Top: unit.Dp(10)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEnd}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(r.theme, fmt.Sprintf("Score: %d", state.Score))
						label.Color = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						comboColor := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
						if state.Combo >= 10 {
							comboColor = color.NRGBA{R: 255, G: 150, B: 50, A: 255}
						}
						label := material.Body1(r.theme, fmt.Sprintf("Combo: %d", state.Combo))
						label.Color = comboColor
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body1(r.theme, fmt.Sprintf("%.0f%%", state.Accuracy()))
						label.Color = color.NRGBA{R: 100, G: 200, B: 255, A: 255}
						return label.Layout(gtx)
					}),
				)
			})
		}),
	)
}

// DrawDetectedNote shows what note the player is currently playing
func (r *TabRenderer) DrawDetectedNote(gtx layout.Context, noteName string, frequency float64, confidence float64) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Left: unit.Dp(10), Bottom: unit.Dp(10)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(r.theme, "Playing: ")
						label.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255}
						return label.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						displayNote := noteName
						if noteName == "" {
							displayNote = "--"
						}
						noteColor := color.NRGBA{R: 100, G: 255, B: 150, A: 255}
						if noteName == "" {
							noteColor = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
						}
						label := material.H6(r.theme, displayNote)
						label.Color = noteColor
						return label.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(15)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := material.Body2(r.theme, fmt.Sprintf("%.1f Hz", frequency))
						label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
						return label.Layout(gtx)
					}),
				)
			})
		}),
	)
}
