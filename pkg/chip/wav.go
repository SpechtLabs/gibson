package chip

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Streaming is the data length [WAVHeader] takes for a WAV of unknown
// length; players read it until the stream ends.
const Streaming = -1

// WAVHeader is the header of an 8-bit unsigned mono WAV at rate with n
// bytes of samples, or of unknown length when n is [Streaming].
func WAVHeader(rate, n int) []byte {
	data, riff := uint32(0xFFFFFFFF), uint32(0xFFFFFFFF)
	if n >= 0 {
		data, riff = uint32(n), uint32(n+36) //nolint:gosec // a WAV can't hold more than 4 GiB anyway
	}
	h := make([]byte, 0, 44)
	h = append(h, "RIFF"...)
	h = binary.LittleEndian.AppendUint32(h, riff)
	h = append(h, "WAVEfmt "...)
	h = binary.LittleEndian.AppendUint32(h, 16)           // fmt chunk size
	h = binary.LittleEndian.AppendUint16(h, 1)            // PCM
	h = binary.LittleEndian.AppendUint16(h, 1)            // mono
	h = binary.LittleEndian.AppendUint32(h, uint32(rate)) //nolint:gosec // sample rates are small
	h = binary.LittleEndian.AppendUint32(h, uint32(rate)) //nolint:gosec // one byte per sample
	h = binary.LittleEndian.AppendUint16(h, 1)            // block align
	h = binary.LittleEndian.AppendUint16(h, 8)            // bits per sample
	h = append(h, "data"...)
	return binary.LittleEndian.AppendUint32(h, data)
}

// ErrNotWAV is returned by [ReadWAVHeader] for anything but an 8-bit
// mono PCM WAV.
var ErrNotWAV = errors.New("chip: not an 8-bit mono PCM WAV")

// ReadWAVHeader reads a WAV's header from r, up to the start of its
// samples, and returns its sample rate. It takes only the 8-bit mono PCM
// that [WAVHeader] describes.
func ReadWAVHeader(r io.Reader) (rate int, err error) {
	var riff [12]byte
	if _, err := io.ReadFull(r, riff[:]); err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}
	if string(riff[:4]) != "RIFF" || string(riff[8:]) != "WAVE" {
		return 0, ErrNotWAV
	}
	for {
		var chunk [8]byte
		if _, err := io.ReadFull(r, chunk[:]); err != nil {
			return 0, fmt.Errorf("read: %w", err)
		}
		id, size := string(chunk[:4]), binary.LittleEndian.Uint32(chunk[4:])
		if id == "data" {
			if rate == 0 {
				return 0, ErrNotWAV // no fmt chunk before the samples
			}
			return rate, nil
		}
		if id != "fmt " {
			if _, err := io.CopyN(io.Discard, r, int64(size+size%2)); err != nil {
				return 0, fmt.Errorf("read: %w", err)
			}
			continue
		}
		if size < 16 || size > 64 {
			return 0, ErrNotWAV
		}
		f := make([]byte, size+size%2)
		if _, err := io.ReadFull(r, f); err != nil {
			return 0, fmt.Errorf("read: %w", err)
		}
		le := binary.LittleEndian
		if le.Uint16(f[0:]) != 1 || le.Uint16(f[2:]) != 1 || le.Uint16(f[14:]) != 8 {
			return 0, ErrNotWAV
		}
		rate = int(le.Uint32(f[4:]))
	}
}

// Stream writes pcm, 8-bit unsigned mono at rate, which never runs out
// (a [Player] doesn't), as a WAV of unknown length to w, paced to play in real
// time with lead of it buffered ahead, until ctx is done or w stops taking
// writes. Pacing keeps a listener in step with whatever started alongside
// it, rather than the song racing as far ahead as the pipe will buffer.
func Stream(ctx context.Context, w io.Writer, pcm io.Reader, rate int, lead time.Duration) error {
	if _, err := w.Write(WAVHeader(rate, Streaming)); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	start := time.Now()
	var sent int
	var buf []byte
	for {
		due := int((time.Since(start) + lead).Seconds() * float64(rate))
		if due > sent {
			buf = buf[:0]
			buf = append(buf, make([]byte, due-sent)...)
			if _, err := io.ReadFull(pcm, buf); err != nil {
				return fmt.Errorf("read: %w", err)
			}
			if _, err := w.Write(buf); err != nil {
				return fmt.Errorf("write: %w", err)
			}
			sent = due
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
	}
}
