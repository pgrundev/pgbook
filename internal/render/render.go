// Package render turns lesson markdown into terminal text.
package render

import "strings"

// Options controls rendering.
type Options struct {
	Color bool
}

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiCyan   = "\x1b[36m"
	ansiYellow = "\x1b[33m"
)

// Render converts markdown to terminal output.
func Render(md string, opts Options) string {
	var b strings.Builder
	style := func(code, s string) string {
		if !opts.Color {
			return s
		}
		return code + s + ansiReset
	}

	inCode := false
	for _, line := range strings.Split(md, "\n") {
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "```"):
			inCode = !inCode
		case inCode:
			b.WriteString("    " + style(ansiCyan, line) + "\n")
		case strings.HasPrefix(line, "### "):
			b.WriteString(style(ansiBold, strings.ToUpper(strings.TrimPrefix(line, "### "))) + "\n")
		case strings.HasPrefix(line, "## "):
			b.WriteString(style(ansiBold, strings.ToUpper(strings.TrimPrefix(line, "## "))) + "\n")
		case strings.HasPrefix(line, "# "):
			b.WriteString(style(ansiBold, strings.ToUpper(strings.TrimPrefix(line, "# "))) + "\n")
		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			b.WriteString("  • " + inline(line[2:], opts) + "\n")
		case strings.HasPrefix(line, "> "):
			b.WriteString("  " + style(ansiYellow, inline(strings.TrimPrefix(line, "> "), Options{})) + "\n")
		case line == ">":
			b.WriteString("\n")
		default:
			b.WriteString(inline(line, opts) + "\n")
		}
	}
	return b.String()
}

// inline strips light markdown emphasis; with color, `code` spans dim.
func inline(s string, opts Options) string {
	if strings.Count(s, "`")%2 == 0 && strings.Contains(s, "`") {
		parts := strings.Split(s, "`")
		var out strings.Builder
		for i, p := range parts {
			if i%2 == 1 && opts.Color {
				out.WriteString(ansiDim + p + ansiReset)
			} else {
				out.WriteString(p)
			}
		}
		s = out.String()
	}
	s = strings.ReplaceAll(s, "**", "")
	return s
}

// ShouldColor reports whether output should use ANSI colors, given
// whether stdout is a terminal and the value of $NO_COLOR.
func ShouldColor(isTTY bool, noColorEnv string) bool {
	return isTTY && noColorEnv == ""
}
