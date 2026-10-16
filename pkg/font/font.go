// Package font is a 5×7 block font for big terminal lettering, drawn two
// pixel rows per cell with half blocks. It covers the letters gibson
// spells; a letter it doesn't have is blank.
package font

// Width and Height are a glyph's size in pixels.
const Width, Height = 5, 7

// Glyph is one letter, a row of '#' (on) and '.' (off) per pixel row.
type Glyph [Height]string

// Lit reports whether pixel (x, y) of g is on.
func (g Glyph) Lit(x, y int) bool { return x < len(g[y]) && g[y][x] == '#' }

// Get returns r's glyph, blank if the font doesn't have it.
func Get(r rune) Glyph { return glyphs[r] }

var glyphs = map[rune]Glyph{
	'A': {".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'B': {"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
	'C': {".####", "#....", "#....", "#....", "#....", "#....", ".####"},
	'D': {"####.", "#...#", "#...#", "#...#", "#...#", "#...#", "####."},
	'E': {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
	'H': {"#...#", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'I': {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "#####"},
	'L': {"#....", "#....", "#....", "#....", "#....", "#....", "#####"},
	'N': {"#...#", "##..#", "#.#.#", "#.#.#", "#.#.#", "#..##", "#...#"},
	'P': {"####.", "#...#", "#...#", "####.", "#....", "#....", "#...."},
	'S': {".####", "#....", "#....", ".###.", "....#", "....#", "####."},
	'T': {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."},
}
