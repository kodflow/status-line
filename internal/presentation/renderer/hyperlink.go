// Package renderer provides status line rendering.
package renderer

import (
	"os"
	"strings"
)

// Hyperlink escape sequence parts (OSC 8).
const (
	// linkEnv enables clickable segments.
	linkEnv string = "STATUSLINE_LINKS"
	// oscStart opens a hyperlink and precedes its target.
	oscStart string = "\033]8;;"
	// oscEnd terminates an OSC string. BEL is what the Claude Code status line
	// documentation uses, and more terminals accept it than the ST form.
	oscEnd string = "\a"
	// fileScheme prefixes a local path turned into a URL.
	fileScheme string = "file://"
)

// linksEnabled reports whether clickable segments are on.
//
// They are on by default: Claude Code documents OSC 8 as a supported status
// line feature and detects whether the terminal handles hyperlinks, so a
// terminal that does not gets plain text rather than visible escape codes.
// Setting the variable to 0 opts out.
var linksEnabled = os.Getenv(linkEnv) != "0"

// Hyperlink wraps text in an OSC 8 hyperlink pointing at a local file.
//
// A status line is not an interactive program: the binary writes its bytes and
// exits, so nothing of ours is running when you move the mouse. Mouse reporting
// belongs to whichever program owns the terminal — Claude Code — and is not
// ours to claim. OSC 8 is the one mechanism that survives that: the terminal
// itself remembers which cells carry a target and handles the click.
//
// Params:
//   - text: text to make clickable
//   - path: local file the click should open
//
// Returns:
//   - string: text wrapped in a hyperlink, or unchanged when disabled
func Hyperlink(text, path string) string {
	// Leave the text alone when links are off or there is nowhere to point
	if !linksEnabled || path == "" {
		return text
	}
	return oscStart + fileScheme + escapePath(path) + oscEnd + text + oscStart + oscEnd
}

// HyperlinkURL wraps text in an OSC 8 hyperlink pointing at an absolute URL.
//
// Params:
//   - text: text to make clickable
//   - url: address the click should open
//
// Returns:
//   - string: text wrapped in a hyperlink, or unchanged when disabled
func HyperlinkURL(text, url string) string {
	// Leave the text alone when links are off or there is nowhere to point
	if !linksEnabled || url == "" {
		return text
	}
	return oscStart + escapePath(url) + oscEnd + text + oscStart + oscEnd
}

// usageURL is where the account's own rate limits are shown in full.
const usageURL string = "https://claude.ai/settings/usage"

// escapePath makes a filesystem path safe to carry inside an OSC 8 target.
//
// Params:
//   - path: local filesystem path
//
// Returns:
//   - string: path with the characters that would terminate the sequence removed
func escapePath(path string) string {
	// Strip the control characters that would close the escape sequence early
	return strings.Map(func(r rune) rune {
		// Drop anything below space, which would break out of the OSC string
		if r < ' ' {
			return -1
		}
		return r
	}, path)
}
