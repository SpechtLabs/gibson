// Package app is the Bubble Tea model that plays the intro, both in the
// local terminal and in every SSH session.
package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	r "github.com/spechtlabs/gibson/pkg/render"
)

// Terminal size cap: a client can claim any pty size, and the canvas is
// allocated per cell, so a 10000x10000 window would be a cheap DoS.
const MaxCols, MaxRows = 400, 150

type tickMsg time.Time

// Renderer draws the frame at t seconds into a w×h cell frame, such as a
// [scene.Scene].
type Renderer interface {
	Render(t float64, w, h int) *r.Frame
}

// Model plays a Renderer at fps frames per second, starting offset seconds in.
type Model struct {
	sc     Renderer
	fps    int
	begin  time.Time
	offset float64
	w, h   int
	frame  *r.Frame

	pausable bool // TEMP: space toggles pause
	paused   bool
}

// New returns a model playing sc. pausable is TEMP: space toggles pause
// instead of quitting.
func New(sc Renderer, fps int, offset float64, pausable bool) *Model {
	return &Model{sc: sc, fps: fps, offset: offset, pausable: pausable}
}

// now is the sequence time; while paused it stays frozen at offset.
func (m *Model) now() float64 {
	if m.paused {
		return m.offset
	}
	return time.Since(m.begin).Seconds() + m.offset
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(m.fps), func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *Model) Init() tea.Cmd {
	m.begin = time.Now()
	return m.tick()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = min(msg.Width, MaxCols), min(msg.Height, MaxRows)
	case tea.KeyPressMsg:
		switch {
		case msg.String() == "r":
			m.begin, m.offset = time.Now(), 0
			return m, nil
		case m.pausable && msg.String() == "space":
			m.offset = m.now()
			m.begin = time.Now()
			m.paused = !m.paused
			return m, nil
		}
		return m, tea.Quit
	case tickMsg:
		if m.w > 0 && m.h > 0 {
			m.frame = m.sc.Render(m.now(), m.w, m.h)
		}
		return m, m.tick()
	}
	return m, nil
}

func (m *Model) View() tea.View {
	content := ""
	if m.frame != nil {
		content = m.frame.ANSI()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
