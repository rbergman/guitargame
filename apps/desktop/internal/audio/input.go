package audio

import (
	"fmt"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	DefaultSampleRate = 48000
	DefaultBufferSize = 2048
)

type AudioInput struct {
	stream     *portaudio.Stream
	buffer     []float32
	sampleRate float64
	bufferSize int
	mu         sync.Mutex
	latest     []float32
}

func NewAudioInput(sampleRate float64, bufferSize int) (*AudioInput, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	input := &AudioInput{
		buffer:     make([]float32, bufferSize),
		latest:     make([]float32, bufferSize),
		sampleRate: sampleRate,
		bufferSize: bufferSize,
	}

	stream, err := portaudio.OpenDefaultStream(
		1,          // input channels (mono)
		0,          // output channels
		sampleRate, // sample rate
		bufferSize, // frames per buffer
		input.processAudio,
	)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	input.stream = stream
	return input, nil
}

func (a *AudioInput) processAudio(in []float32) {
	a.mu.Lock()
	copy(a.latest, in)
	a.mu.Unlock()
}

func (a *AudioInput) Start() error {
	return a.stream.Start()
}

func (a *AudioInput) Stop() error {
	return a.stream.Stop()
}

func (a *AudioInput) Close() error {
	if err := a.stream.Close(); err != nil {
		return err
	}
	return portaudio.Terminate()
}

func (a *AudioInput) GetBuffer() []float32 {
	a.mu.Lock()
	defer a.mu.Unlock()
	copy(a.buffer, a.latest)
	return a.buffer
}

func (a *AudioInput) SampleRate() float64 {
	return a.sampleRate
}

func (a *AudioInput) BufferSize() int {
	return a.bufferSize
}

func ListDevices() error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return err
	}

	fmt.Println("Available audio devices:")
	for i, d := range devices {
		if d.MaxInputChannels > 0 {
			fmt.Printf("  [%d] %s (inputs: %d, sample rate: %.0f)\n",
				i, d.Name, d.MaxInputChannels, d.DefaultSampleRate)
		}
	}
	return nil
}
