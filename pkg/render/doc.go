// Package render is a braille sub-pixel canvas for the terminal. Lines and
// dots accumulate light per sub-pixel, a per-cell depth test tucks things
// behind each other, and Compose resolves it all into colored cells with
// bloom, ready to print as ANSI.
package render
