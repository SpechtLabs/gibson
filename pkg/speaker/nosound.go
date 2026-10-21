//go:build nosound

package speaker

import (
	"context"
	"errors"
	"io"
)

// ErrNoSound is returned by Play in builds made with the nosound tag.
var ErrNoSound = errors.New("speaker: built without sound (nosound)")

// Play fails: this build has no sound.
func Play(context.Context, io.Reader, int) error { return ErrNoSound }
