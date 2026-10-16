package scene

import (
	"fmt"
	"slices"

	"github.com/spechtlabs/gibson/pkg/components/containment"
	"github.com/spechtlabs/gibson/pkg/components/containment/orbiter"
	"github.com/spechtlabs/gibson/pkg/components/containment/ring"
	"github.com/spechtlabs/gibson/pkg/components/specimen"
	"github.com/spechtlabs/gibson/pkg/components/specimen/bird"
	"github.com/spechtlabs/gibson/pkg/components/specimen/blob"
	"github.com/spechtlabs/gibson/pkg/components/vessel"
	"github.com/spechtlabs/gibson/pkg/components/vessel/dodeca"
	"github.com/spechtlabs/gibson/pkg/components/vessel/hex"
	"github.com/spechtlabs/gibson/pkg/theme"
	"github.com/spechtlabs/gibson/pkg/theme/specht"
	"github.com/spechtlabs/gibson/pkg/theme/synth"
	"github.com/spechtlabs/gibson/pkg/theme/umbrella"
)

// The parts gibson can be configured with, by the name they're picked by.
// Constructors rather than values, so every scene gets parts of its own.
var (
	themes = []entry[theme.Theme]{
		{"synth", synth.New},
		{"specht", specht.New},
		{"umbrella", umbrella.New},
	}
	specimens = []entry[specimen.Specimen]{
		{"bird", func() specimen.Specimen { return bird.New() }},
		{"blob", func() specimen.Specimen { return blob.New() }},
	}
	vessels = []entry[vessel.Vessel]{
		{"hex", func() vessel.Vessel { return hex.New() }},
		{"dodeca", func() vessel.Vessel { return dodeca.New() }},
	}
)

type entry[T any] struct {
	name string
	make func() T
}

// Field returns the containment field that comes with v: rings around the
// hex prism, orbiting comets around the dodecahedron. Which one is the
// point, so it returns the interface.
func Field(v vessel.Vessel) containment.Field {
	if v.Name() == "dodeca" {
		return orbiter.New(orbiter.WithRadius(1.35))
	}
	return ring.New(ring.WithCount(3))
}

// Themes lists the theme names.
func Themes() []string { return names(themes) }

// Specimens lists the specimen names.
func Specimens() []string { return names(specimens) }

// Vessels lists the vessel names.
func Vessels() []string { return names(vessels) }

func names[T any](es []entry[T]) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.name
	}
	return out
}

func lookup[T any](kind string, es []entry[T], name string) (T, error) {
	for _, e := range es {
		if e.name == name {
			return e.make(), nil
		}
	}
	var zero T
	return zero, fmt.Errorf("unknown %s %q (have: %v)", kind, name, names(es))
}

// Look is a scene picked by names, as configuration and SSH sessions do.
type Look struct {
	Theme, Specimen, Vessel string
	Tagline                 string
}

// Override returns l with every theme, specimen or vessel name among words
// replacing its own; other words are ignored.
func (l Look) Override(words []string) Look {
	for _, w := range words {
		switch {
		case slices.Contains(Themes(), w):
			l.Theme = w
		case slices.Contains(Specimens(), w):
			l.Specimen = w
		case slices.Contains(Vessels(), w):
			l.Vessel = w
		}
	}
	return l
}

// Options resolves l's names into scene options, and fails naming the
// valid values for any it doesn't know.
func (l Look) Options() ([]Option, error) {
	th, err := lookup("theme", themes, l.Theme)
	if err != nil {
		return nil, err
	}
	sp, err := lookup("specimen", specimens, l.Specimen)
	if err != nil {
		return nil, err
	}
	ve, err := lookup("vessel", vessels, l.Vessel)
	if err != nil {
		return nil, err
	}
	return []Option{WithTheme(th), WithSpecimen(sp), WithVessel(ve), WithTagline(l.Tagline)}, nil
}

// New assembles the scene l picks.
func (l Look) New() (*Scene, error) {
	opts, err := l.Options()
	if err != nil {
		return nil, err
	}
	return New(opts...), nil
}
