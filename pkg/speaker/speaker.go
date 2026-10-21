//go:build !nosound

// Package speaker plays 8-bit unsigned mono PCM on the machine's audio
// output with oto: Core Audio on macOS, PulseAudio (or ALSA) on Linux,
// all without cgo. On Linux, oto loads those libraries at run time with
// purego, which links the binary dynamically against libc, even with
// CGO_ENABLED=0; the nosound build tag leaves sound out to keep it static,
// as the container image needs.
package speaker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

// latency is the device buffer: short, so a restart is heard at once,
// long enough not to crackle.
const latency = 80 * time.Millisecond

// ErrRate is returned when Play is asked for a rate other than the one the
// process's audio output was opened at; oto opens it once per process.
var ErrRate = errors.New("speaker: the audio output is already open at another sample rate")

var (
	mu     sync.Mutex
	device *oto.Context
	ready  chan struct{} // closed once device is up, or failed to come up
	opened int           // the rate device runs at
)

// open opens the audio output at rate, once per process. Bringing it up
// can take a moment, or forever on a machine whose sound server doesn't
// answer, so it gives up waiting when ctx is done.
func open(ctx context.Context, rate int) (*oto.Context, error) {
	mu.Lock()
	if device == nil {
		c, r, err := oto.NewContext(&oto.NewContextOptions{
			SampleRate:      rate,
			ChannelCount:    1,
			Format:          oto.FormatUnsignedInt8,
			BufferSize:      latency,
			ApplicationName: "gibson",
		})
		if err != nil {
			mu.Unlock()
			return nil, fmt.Errorf("speaker: %w", err)
		}
		device, ready, opened = c, r, rate
	}
	c, r := device, ready
	mu.Unlock()

	if rate != opened {
		return nil, ErrRate
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r:
	}
	if err := c.Err(); err != nil {
		return nil, fmt.Errorf("speaker: %w", err)
	}
	return c, nil
}

// Play plays r, 8-bit unsigned mono PCM at rate samples per second, until
// r runs out or ctx is done. It fails when there's no audio output.
func Play(ctx context.Context, r io.Reader, rate int) error {
	c, err := open(ctx, rate)
	if err != nil {
		return err
	}
	p := c.NewPlayer(r)
	// oto reads ahead of the device too, half a second by default; a
	// restart would lag by that much
	p.SetBufferSize(rate / 8)
	defer p.PauseAndStopReading()
	p.Play()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for p.IsPlaying() {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
	return p.Err()
}
