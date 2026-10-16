package denied

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Quote is a line a turned-away viewer may get, and where it's from.
type Quote struct {
	Text  string
	Movie string
	Year  int
	// Speaker is who says it, if it's worth naming; leave it empty rather
	// than guess.
	Speaker string
}

// Attribution is the line under the quote: "— HACKERS (1995)", or
// "— ZERO COOL, HACKERS (1995)" with a speaker.
func (q Quote) Attribution() string {
	src := fmt.Sprintf("%s (%d)", strings.ToUpper(q.Movie), q.Year)
	if q.Speaker != "" {
		src = strings.ToUpper(q.Speaker) + ", " + src
	}
	return "— " + src
}

const hackers = "Hackers"

// Quotes are the lines a turned-away viewer may get. Add more, from any
// movie, with their attribution.
var Quotes = []Quote{
	{Text: "HACK THE PLANET!", Movie: hackers, Year: 1995},
	{Text: "Mess with the best, die like the rest.", Movie: hackers, Year: 1995},
	{Text: "There is no right and wrong. There's only fun and boring.", Movie: hackers, Year: 1995},
	{Text: "Hackers of the world, unite!", Movie: hackers, Year: 1995},
	{Text: "Type \"cookie\", you idiot.", Movie: hackers, Year: 1995},
	{Text: "Never fear, I is here.", Movie: hackers, Year: 1995},
	{Text: "RISC architecture is gonna change everything.", Movie: hackers, Year: 1995},
	{Text: "The pool on the roof must have a leak.", Movie: hackers, Year: 1995},
	{Text: "Remember, hacking is more than just a crime. It's a survival trait.", Movie: hackers, Year: 1995},
}

// Pick picks one of [Quotes] at random.
func Pick() Quote { return Quotes[rand.IntN(len(Quotes))] }
