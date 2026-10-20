package chip

import (
	"context"
	"encoding/binary"
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

// Stream writes p as a WAV of unknown length to w, paced to play in real
// time with lead of it buffered ahead, until ctx is done or w stops taking
// writes. Pacing keeps a listener in step with whatever started alongside
// it, rather than the song racing as far ahead as the pipe will buffer.
func Stream(ctx context.Context, w io.Writer, p *Player, rate int, lead time.Duration) error {
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
			_, _ = p.Read(buf)
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
