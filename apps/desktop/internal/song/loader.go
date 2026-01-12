package song

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// LoadSong loads a song from a YAML file
func LoadSong(path string) (*Song, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var song Song
	if err := yaml.Unmarshal(data, &song); err != nil {
		return nil, err
	}

	// Convert beat numbers to time if specified
	if song.BPM > 0 {
		beatDuration := 60.0 / song.BPM
		for i := range song.Notes {
			// If beat is specified but time is not, convert beat to time
			if song.Notes[i].Beat > 0 && song.Notes[i].Time == 0 {
				song.Notes[i].Time = song.Notes[i].Beat * beatDuration
			}
			// Default duration to one beat if not specified
			if song.Notes[i].Duration == 0 {
				song.Notes[i].Duration = beatDuration * 0.9
			}
		}
	}

	// Sort notes by time
	sort.Slice(song.Notes, func(i, j int) bool {
		return song.Notes[i].Time < song.Notes[j].Time
	})

	// Parse tuning
	if song.TuningStr != "" {
		song.Tuning = ParseTuning(song.TuningStr)
	} else {
		song.Tuning = TuningStandard
	}

	song.CalculateDuration()
	return &song, nil
}

// LoadSongsFromDirectory loads all .yaml and .yml files from a directory
func LoadSongsFromDirectory(dir string) ([]*Song, error) {
	var songs []*Song

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		song, err := LoadSong(path)
		if err != nil {
			// Log error but continue loading other songs
			continue
		}

		songs = append(songs, song)
	}

	// Sort songs by title
	sort.Slice(songs, func(i, j int) bool {
		return songs[i].Title < songs[j].Title
	})

	return songs, nil
}

// SaveSong saves a song to a YAML file
func SaveSong(song *Song, path string) error {
	data, err := yaml.Marshal(song)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetDefaultExercises returns built-in exercises if no songs directory exists
func GetDefaultExercises() []*Song {
	return []*Song{
		createEMinorScale(),
	}
}

// createEMinorScale creates a simple E minor scale exercise
func createEMinorScale() *Song {
	beatDuration := 60.0 / 80.0

	notes := []TabNote{
		{Time: 0 * beatDuration, String: StringE, Fret: 0, Duration: beatDuration * 0.9},
		{Time: 1 * beatDuration, String: StringE, Fret: 2, Duration: beatDuration * 0.9},
		{Time: 2 * beatDuration, String: StringE, Fret: 3, Duration: beatDuration * 0.9},
		{Time: 3 * beatDuration, String: StringE, Fret: 5, Duration: beatDuration * 0.9},
		{Time: 4 * beatDuration, String: StringA, Fret: 0, Duration: beatDuration * 0.9},
		{Time: 5 * beatDuration, String: StringA, Fret: 2, Duration: beatDuration * 0.9},
		{Time: 6 * beatDuration, String: StringA, Fret: 3, Duration: beatDuration * 0.9},
		{Time: 7 * beatDuration, String: StringA, Fret: 5, Duration: beatDuration * 0.9},
	}

	song := &Song{
		Title:  "E Minor Scale (Default)",
		Artist: "Built-in",
		BPM:    80,
		Notes:  notes,
		Tuning: TuningStandard,
	}
	song.CalculateDuration()
	return song
}
