package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"guitargame/apps/desktop/internal/audio"
	"guitargame/apps/desktop/internal/game"
	"guitargame/apps/desktop/internal/render"
	"guitargame/apps/desktop/internal/song"
)

const (
	screenWidth  = 1000
	screenHeight = 500
)

// AppState represents the current screen
type AppState int

const (
	StateMenu AppState = iota
	StatePreStart
	StatePlaying
	StateResults
)

type App struct {
	audioInput    *audio.AudioInput
	pitchDetector *audio.PitchDetector
	currentPitch  audio.PitchResult

	theme       *material.Theme
	tabRenderer *render.TabRenderer
	hitDetector *game.HitDetector
	gameState   *song.GameState

	// Song selection
	exercises     []*song.Song
	selectedIndex int

	// UI state
	state            AppState
	lastNoteDetected bool
}

func NewApp() (*App, error) {
	sampleRate := float64(audio.DefaultSampleRate)
	bufferSize := audio.DefaultBufferSize

	audioInput, err := audio.NewAudioInput(sampleRate, bufferSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio input: %w", err)
	}

	pitchDetector := audio.NewPitchDetector(bufferSize, sampleRate)

	if err := audioInput.Start(); err != nil {
		audioInput.Close()
		return nil, fmt.Errorf("failed to start audio: %w", err)
	}

	theme := material.NewTheme()
	tabRenderer := render.NewTabRenderer(theme)

	// Load songs from directory
	exercises, err := loadSongs()
	if err != nil {
		log.Printf("Warning: could not load songs: %v", err)
	}
	if len(exercises) == 0 {
		// Fall back to default exercises
		exercises = song.GetDefaultExercises()
	}

	// Initialize with first exercise
	gameState := song.NewGameState(exercises[0])
	hitDetector := game.NewHitDetector(gameState)

	return &App{
		audioInput:    audioInput,
		pitchDetector: pitchDetector,
		theme:         theme,
		tabRenderer:   tabRenderer,
		hitDetector:   hitDetector,
		gameState:     gameState,
		exercises:     exercises,
		selectedIndex: 0,
		state:         StateMenu,
	}, nil
}

func (a *App) Update() {
	// Get audio and detect pitch
	buffer := a.audioInput.GetBuffer()
	a.currentPitch = a.pitchDetector.Detect(buffer)

	if a.state != StatePlaying {
		return
	}

	// Update game state
	a.gameState.Update()

	// Check for hits
	playLineX := float32(screenWidth) * a.tabRenderer.PlayLineX
	a.hitDetector.CheckHit(a.currentPitch, playLineX)
	a.hitDetector.Update()

	// Check if song finished
	if a.gameState.IsFinished {
		a.state = StateResults
	}
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	a.Update()

	// Background
	paint.ColorOp{Color: color.NRGBA{R: 20, G: 20, B: 30, A: 255}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	switch a.state {
	case StateMenu:
		return a.layoutMenuScreen(gtx)
	case StatePreStart:
		return a.layoutPreStartScreen(gtx)
	case StatePlaying:
		return a.layoutGameScreen(gtx)
	case StateResults:
		return a.layoutResultsScreen(gtx)
	}

	return layout.Dimensions{}
}

func (a *App) layoutMenuScreen(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Title
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Top: unit.Dp(20), Left: unit.Dp(20)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.H4(a.theme, "Bass Guitar Practice")
				label.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
				return label.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Left: unit.Dp(20), Bottom: unit.Dp(20)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(a.theme, "Select an exercise (play a note to select)")
				label.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255}
				return label.Layout(gtx)
			})
		}),
		// Exercise list
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.layoutExerciseList(gtx)
		}),
		// Detected note display
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tabRenderer.DrawDetectedNote(gtx, a.currentPitch.FullNoteName(), a.currentPitch.Frequency, a.currentPitch.Confidence)
		}),
		// Instructions
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{Left: unit.Dp(20), Bottom: unit.Dp(15)}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body2(a.theme, "E1=41Hz  A1=55Hz  D2=73Hz  G2=98Hz")
				label.Color = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
				return label.Layout(gtx)
			})
		}),
	)
}

func (a *App) layoutExerciseList(gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{Left: unit.Dp(20), Right: unit.Dp(20)}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var children []layout.FlexChild
				for i, ex := range a.exercises {
					idx := i // capture for closure
					exercise := ex
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.layoutExerciseItem(gtx, idx, exercise)
					}))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			}),
		)
	})
}

func (a *App) layoutExerciseItem(gtx layout.Context, index int, exercise *song.Song) layout.Dimensions {
	isSelected := index == a.selectedIndex

	bgColor := color.NRGBA{R: 35, G: 35, B: 45, A: 255}
	if isSelected {
		bgColor = color.NRGBA{R: 50, G: 70, B: 90, A: 255}
	}

	return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Background
		width := gtx.Constraints.Max.X
		height := gtx.Dp(unit.Dp(50))

		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.ColorOp{Color: bgColor}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		// Selection indicator
		if isSelected {
			defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
			paint.ColorOp{Color: color.NRGBA{R: 100, G: 200, B: 255, A: 50}}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
		}

		// Content
		inset := layout.Inset{Left: unit.Dp(15), Top: unit.Dp(10), Right: unit.Dp(15)}
		inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							titleColor := color.NRGBA{R: 200, G: 200, B: 200, A: 255}
							if isSelected {
								titleColor = color.NRGBA{R: 100, G: 200, B: 255, A: 255}
							}
							label := material.Body1(a.theme, exercise.Title)
							label.Color = titleColor
							return label.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := material.Body2(a.theme, exercise.Artist)
							label.Color = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
							return label.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					info := fmt.Sprintf("%.0f BPM • %d notes", exercise.BPM, len(exercise.Notes))
					label := material.Body2(a.theme, info)
					label.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255}
					return label.Layout(gtx)
				}),
			)
		})

		return layout.Dimensions{Size: gtx.Constraints.Constrain(
			layout.Dimensions{Size: gtx.Constraints.Max}.Size,
		)}
		return layout.Dimensions{Size: struct{ X, Y int }{width, height}}
	})
}

func (a *App) layoutPreStartScreen(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceAround}.Layout(gtx,
		layout.Flexed(1, layout.Spacer{}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H5(a.theme, a.gameState.Song.Title)
			label.Color = color.NRGBA{R: 150, G: 200, B: 255, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, fmt.Sprintf("%.0f BPM  •  %d notes", a.gameState.Song.BPM, len(a.gameState.Song.Notes)))
			label.Color = color.NRGBA{R: 120, G: 120, B: 120, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tabRenderer.DrawDetectedNote(gtx, a.currentPitch.FullNoteName(), a.currentPitch.Frequency, a.currentPitch.Confidence)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(30)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, "Play any note to start!")
			label.Color = color.NRGBA{R: 100, G: 200, B: 100, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Flexed(1, layout.Spacer{}.Layout),
	)
}

func (a *App) layoutGameScreen(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header with score
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tabRenderer.DrawHeader(gtx, a.gameState)
		}),
		// Tab area
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.tabRenderer.Layout(gtx, a.gameState)
		}),
		// Detected note display
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.tabRenderer.DrawDetectedNote(gtx, a.currentPitch.FullNoteName(), a.currentPitch.Frequency, a.currentPitch.Confidence)
		}),
	)
}

func (a *App) layoutResultsScreen(gtx layout.Context) layout.Dimensions {
	accuracy := a.gameState.Accuracy()
	grade := getGrade(accuracy)

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle, Spacing: layout.SpaceAround}.Layout(gtx,
		layout.Flexed(1, layout.Spacer{}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H4(a.theme, "Exercise Complete!")
			label.Color = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H2(a.theme, grade)
			label.Color = getGradeColor(grade)
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(15)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.H5(a.theme, fmt.Sprintf("Score: %d", a.gameState.Score))
			label.Color = color.NRGBA{R: 255, G: 215, B: 0, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, fmt.Sprintf("Accuracy: %.1f%%  •  Max Combo: %d  •  Notes: %d/%d",
				accuracy, a.gameState.MaxCombo, a.gameState.NotesHit, a.gameState.TotalNotes))
			label.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, "Play a note to return to menu")
			label.Color = color.NRGBA{R: 100, G: 200, B: 100, A: 255}
			return layout.Center.Layout(gtx, label.Layout)
		}),
		layout.Flexed(1, layout.Spacer{}.Layout),
	)
}

func (a *App) SelectExercise(index int) {
	if index >= 0 && index < len(a.exercises) {
		a.selectedIndex = index
		a.gameState = song.NewGameState(a.exercises[index])
		a.hitDetector = game.NewHitDetector(a.gameState)
	}
}

func (a *App) StartGame() {
	a.state = StatePlaying
	a.gameState.Start()
}

func (a *App) GoToMenu() {
	a.state = StateMenu
	a.SelectExercise(a.selectedIndex)
}

func (a *App) Close() {
	if a.pitchDetector != nil {
		a.pitchDetector.Close()
	}
	if a.audioInput != nil {
		a.audioInput.Stop()
		a.audioInput.Close()
	}
}

func getGrade(accuracy float64) string {
	switch {
	case accuracy >= 95:
		return "S"
	case accuracy >= 90:
		return "A"
	case accuracy >= 80:
		return "B"
	case accuracy >= 70:
		return "C"
	case accuracy >= 60:
		return "D"
	default:
		return "F"
	}
}

func getGradeColor(grade string) color.NRGBA {
	switch grade {
	case "S":
		return color.NRGBA{R: 255, G: 215, B: 0, A: 255}
	case "A":
		return color.NRGBA{R: 100, G: 255, B: 100, A: 255}
	case "B":
		return color.NRGBA{R: 100, G: 200, B: 255, A: 255}
	case "C":
		return color.NRGBA{R: 255, G: 255, B: 100, A: 255}
	case "D":
		return color.NRGBA{R: 255, G: 150, B: 50, A: 255}
	default:
		return color.NRGBA{R: 255, G: 100, B: 100, A: 255}
	}
}

// loadSongs tries to load songs from various locations
func loadSongs() ([]*song.Song, error) {
	// Try these directories in order:
	// 1. ./songs (relative to current directory)
	// 2. songs/ next to executable
	// 3. ~/.config/guitargame/songs

	searchPaths := []string{
		"songs",
	}

	// Add path relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		searchPaths = append(searchPaths, filepath.Join(exeDir, "songs"))
	}

	// Add config directory
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(home, ".config", "guitargame", "songs"))
	}

	for _, path := range searchPaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			songs, err := song.LoadSongsFromDirectory(path)
			if err == nil && len(songs) > 0 {
				fmt.Printf("Loaded %d songs from %s\n", len(songs), path)
				return songs, nil
			}
		}
	}

	return nil, fmt.Errorf("no songs found in any search path")
}

func main() {
	fmt.Println("Bass Guitar Practice Game")
	fmt.Println("=========================")
	fmt.Println()

	fmt.Println("Listing audio devices...")
	if err := audio.ListDevices(); err != nil {
		log.Printf("Warning: could not list devices: %v", err)
	}
	fmt.Println()

	application, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer application.Close()

	fmt.Println("Starting game...")
	fmt.Println("Exercises available:")
	for i, ex := range application.exercises {
		fmt.Printf("  %d. %s (%.0f BPM, %d notes)\n", i+1, ex.Title, ex.BPM, len(ex.Notes))
	}
	fmt.Println()

	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Bass Guitar Practice"),
			app.Size(unit.Dp(screenWidth), unit.Dp(screenHeight)),
		)

		var ops op.Ops

		// 60 FPS ticker
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				w.Invalidate()
			}
		}()

		// Debounce note detection for menu navigation
		lastNoteTime := time.Time{}
		noteCooldown := 300 * time.Millisecond

		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				if e.Err != nil {
					log.Fatal(e.Err)
				}
				os.Exit(0)

			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				noteDetected := application.currentPitch.IsValid()
				noteJustPlayed := noteDetected && time.Since(lastNoteTime) > noteCooldown

				if noteJustPlayed {
					lastNoteTime = time.Now()

					switch application.state {
					case StateMenu:
						// Cycle through exercises or start selected
						if application.lastNoteDetected {
							// Second note - start the game
							application.state = StatePreStart
						} else {
							// First note - cycle selection
							application.selectedIndex = (application.selectedIndex + 1) % len(application.exercises)
							application.SelectExercise(application.selectedIndex)
						}
					case StatePreStart:
						application.StartGame()
					case StateResults:
						application.GoToMenu()
					}
				}

				application.lastNoteDetected = noteDetected

				application.Layout(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}()

	app.Main()
}
