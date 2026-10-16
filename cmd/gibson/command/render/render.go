// Package render implements gibson render, which renders the intro to a
// PNG frame or a GIF instead of playing it.
package render

import (
	"errors"
	"fmt"
	"image/gif"
	"image/png"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/config"
	"github.com/spechtlabs/gibson/internal/scene"
	"github.com/spechtlabs/gibson/pkg/preview"
	r "github.com/spechtlabs/gibson/pkg/render"
)

// ErrNothingToRender is returned when neither --png nor --gif is set.
var ErrNothingToRender = errors.New("nothing to render: set --png or --gif")

// flags maps render's flags to their config keys.
var flags = map[string]string{
	"png":  config.RenderPNG,
	"gif":  config.RenderGIF,
	"at":   config.RenderAt,
	"dur":  config.RenderDur,
	"size": config.RenderSize,
}

// NewCommand returns the render command, which reads its settings from v.
func NewCommand(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render the intro to a PNG or GIF",
		Long: `Renders the intro offline, as it would look in a terminal of --size: a single
frame to --png, or the sequence to --gif. With both, only the PNG is
written.`,
		Example: `# The frame 6 seconds in
gibson render --png frame.png

# The first 10 seconds, in the specht theme
gibson render --gif intro.gif --theme specht`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Bind(v, cmd.Flags(), flags); err != nil {
				return err
			}
			look, err := config.Look(v)
			if err != nil {
				return err
			}
			pngOut, gifOut := v.GetString(config.RenderPNG), v.GetString(config.RenderGIF)
			if pngOut == "" && gifOut == "" {
				return ErrNothingToRender
			}
			var w, h int
			if _, err := fmt.Sscanf(v.GetString(config.RenderSize), "%dx%d", &w, &h); err != nil {
				return fmt.Errorf("bad --size: %w", err)
			}
			sc, err := look.New()
			if err != nil {
				return err
			}
			return offline(sc, w, h, pngOut, gifOut,
				v.GetFloat64(config.RenderAt), v.GetFloat64(config.RenderDur), v.GetInt(config.FPS))
		},
	}

	f := cmd.Flags()
	f.String("png", "", "render a single frame to this PNG")
	f.String("gif", "", "render the sequence to this GIF")
	f.Float64("at", 6, "time of the --png frame (seconds)")
	f.Float64("dur", 10, "duration of the --gif (seconds)")
	f.String("size", "120x40", "terminal size (cols x rows)")
	return cmd
}

func offline(sc *scene.Scene, w, h int, pngOut, gifOut string, at, dur float64, fps int) error {
	if pngOut != "" {
		f, err := os.Create(pngOut)
		if err != nil {
			return err
		}
		if err := png.Encode(f, preview.Image(sc.Render(at, w, h))); err != nil {
			_ = f.Close()
			return err
		}
		return f.Close()
	}
	th := sc.Palette()
	keys := append([]r.RGB{th.Specimen, th.Rings, th.Orbit, th.Beam, th.HUD, th.Text, th.Warn}, th.Grad...)
	keys = append(keys, th.Title...)
	q := preview.NewQuantizer(preview.Palette(keys))
	gfps := min(fps, 20)
	out := &gif.GIF{}
	for i := 0; float64(i) < dur*float64(gfps); i++ {
		t := float64(i) / float64(gfps)
		out.Image = append(out.Image, q.Paletted(preview.Image(sc.Render(t, w, h))))
		out.Delay = append(out.Delay, 100/gfps)
	}
	f, err := os.Create(gifOut)
	if err != nil {
		return err
	}
	if err := gif.EncodeAll(f, out); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
