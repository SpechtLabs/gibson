package denied

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Quote is a line a turned-away viewer may get, and where it's from: a
// movie, a publication, or just someone who said it.
type Quote struct {
	Text string
	// Source is the movie (or magazine, or column) it's from; empty for
	// something a person said.
	Source string
	Year   int
	// Speaker is who says it. In a movie, leave it empty unless it's
	// certain.
	Speaker string
	// Apocryphal marks a line that's only ever been attributed to its
	// speaker, never shown to be theirs.
	Apocryphal bool
}

// Attribution is the line under the quote, e.g. "— HACKERS (1995)",
// "— ZERO COOL, HACKERS (1995)", "— BILL GATES, 2004" or
// "— ATTRIBUTED TO BILL GATES, 1981".
func (q Quote) Attribution() string {
	var src string
	switch {
	case q.Source != "" && q.Speaker != "":
		src = fmt.Sprintf("%s, %s (%d)", q.Speaker, q.Source, q.Year)
	case q.Source != "":
		src = fmt.Sprintf("%s (%d)", q.Source, q.Year)
	default:
		src = fmt.Sprintf("%s, %d", q.Speaker, q.Year)
	}
	if q.Apocryphal {
		src = "attributed to " + src
	}
	return "— " + strings.ToUpper(src)
}

// Sources quoted more than once.
const (
	hackers      = "Hackers"
	jurassicPark = "Jurassic Park"
)

// Quotes are the lines a turned-away viewer may get: Hackers first, then
// other movies with a line for being shut out, told no or told to come back
// later, then the computing industry's most confident predictions. Add
// more with their attribution; mark anything only attributed as Apocryphal.
var Quotes = []Quote{
	{Text: "HACK THE PLANET!", Source: hackers, Year: 1995},
	{Text: "Mess with the best, die like the rest.", Source: hackers, Year: 1995, Speaker: "Zero Cool"},
	{Text: "There is no right and wrong. There's only fun and boring.", Source: hackers, Year: 1995},
	{Text: "Hackers of the world, unite!", Source: hackers, Year: 1995},
	{Text: "Type \"cookie\", you idiot.", Source: hackers, Year: 1995},
	{Text: "Never fear, I is here.", Source: hackers, Year: 1995},
	{Text: "RISC architecture is gonna change everything.", Source: hackers, Year: 1995},
	{Text: "The pool on the roof must have a leak.", Source: hackers, Year: 1995},
	{Text: "Remember, hacking is more than just a crime. It's a survival trait.", Source: hackers, Year: 1995},
	{Text: "My crime is that of curiosity.", Source: hackers, Year: 1995},

	{Text: "I'm sorry, Dave. I'm afraid I can't do that.", Source: "2001: A Space Odyssey", Year: 1968, Speaker: "HAL 9000"},
	{Text: "Ah ah ah! You didn't say the magic word!", Source: jurassicPark, Year: 1993, Speaker: "Dennis Nedry"},
	{Text: "It's a UNIX system! I know this!", Source: jurassicPark, Year: 1993, Speaker: "Lex Murphy"},
	{Text: "Hold on to your butts.", Source: jurassicPark, Year: 1993, Speaker: "Ray Arnold"},
	{Text: "You shall not pass!", Source: "The Lord of the Rings: The Fellowship of the Ring", Year: 2001, Speaker: "Gandalf"},
	{Text: "None shall pass.", Source: "Monty Python and the Holy Grail", Year: 1975, Speaker: "The Black Knight"},
	{Text: "I'll be back.", Source: "The Terminator", Year: 1984, Speaker: "The Terminator"},
	{Text: "End of line.", Source: "Tron", Year: 1982, Speaker: "Master Control Program"},
	{Text: "A strange game. The only winning move is not to play.", Source: "WarGames", Year: 1983},
	{Text: "Shall we play a game?", Source: "WarGames", Year: 1983},
	{Text: "Too many secrets.", Source: "Sneakers", Year: 1992},
	{Text: "There is no spoon.", Source: "The Matrix", Year: 1999, Speaker: "Spoon Boy"},
	{Text: "It's a trap!", Source: "Return of the Jedi", Year: 1983, Speaker: "Admiral Ackbar"},
	{Text: "You're all going to die down here.", Source: "Resident Evil", Year: 2002, Speaker: "The Red Queen"},
	{Text: "Game over, man! Game over!", Source: "Aliens", Year: 1986, Speaker: "Hudson"},
	{Text: "Don't cross the streams.", Source: "Ghostbusters", Year: 1984, Speaker: "Egon Spengler"},
	{Text: "You get nothing! You lose! Good day, sir!", Source: "Willy Wonka & the Chocolate Factory", Year: 1971, Speaker: "Willy Wonka"},

	{Text: "640K ought to be enough for anybody.", Speaker: "Bill Gates", Year: 1981, Apocryphal: true},
	{Text: "I think there is a world market for maybe five computers.", Speaker: "Thomas Watson", Year: 1943, Apocryphal: true},
	{Text: "There is no reason for any individual to have a computer in his home.", Speaker: "Ken Olsen", Year: 1977, Apocryphal: true},
	{Text: "Computers in the future may weigh no more than 1.5 tons.", Source: "Popular Mechanics", Year: 1949},
	{Text: "I predict the Internet will soon go spectacularly supernova and in 1996 catastrophically collapse.", Speaker: "Robert Metcalfe", Year: 1995},
	{Text: "Two years from now, spam will be solved.", Speaker: "Bill Gates", Year: 2004},
	{Text: "There's no chance that the iPhone is going to get any significant market share. No chance.", Speaker: "Steve Ballmer", Year: 2007},
}

// Pick picks one of [Quotes] at random.
func Pick() Quote { return Quotes[rand.IntN(len(Quotes))] }
