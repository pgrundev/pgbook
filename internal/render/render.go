// Package render turns lesson markdown into terminal text.
package render

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Options controls rendering.
type Options struct {
	Color bool
}

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiInverse = "\x1b[7m"
	ansiCyan    = "\x1b[36m"
	ansiMagenta = "\x1b[35m"
	ansiYellow  = "\x1b[33m"
	ansiGreen   = "\x1b[32m"
)

// stepHeading matches tutorial step headings such as
// "## Step 2: Watch a query crawl". Steps are numbered in the source so
// the website and PDF read naturally; the terminal draws a tracker.
var stepHeading = regexp.MustCompile(`^## Step (\d+)\s*[:.—–-]\s*(.+?)\s*$`)

const stepRule = "────────────────────────────────────────────────────────"

// Render converts markdown to terminal output.
func Render(md string, opts Options) string {
	var b strings.Builder
	style := func(code, s string) string {
		if !opts.Color {
			return s
		}
		return code + s + ansiReset
	}

	lines := strings.Split(md, "\n")
	totalSteps := countSteps(lines)

	inCode := false
	codeLang := ""
	for _, line := range lines {
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "```"):
			inCode = !inCode
			codeLang = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "```")))
		case inCode && opts.Color && isSQL(codeLang):
			b.WriteString("    " + highlightSQL(line) + "\n")
		case inCode:
			b.WriteString("    " + style(ansiCyan, line) + "\n")
		case strings.HasPrefix(line, "### "):
			b.WriteString(style(ansiBold, strings.ToUpper(strings.TrimPrefix(line, "### "))) + "\n")
		case stepHeading.MatchString(line):
			m := stepHeading.FindStringSubmatch(line)
			n, _ := strconv.Atoi(m[1])
			b.WriteString(stepHeader(n, totalSteps, m[2], opts))
		case strings.HasPrefix(line, "## "):
			b.WriteString(style(ansiBold, strings.ToUpper(strings.TrimPrefix(line, "## "))) + "\n")
		case strings.HasPrefix(line, "# "):
			b.WriteString(style(ansiBold, strings.ToUpper(strings.TrimPrefix(line, "# "))) + "\n")
		case strings.HasPrefix(line, "- [ ] "):
			b.WriteString("  " + style(ansiYellow, "☐") + " " + inline(line[6:], opts) + "\n")
		case strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] "):
			b.WriteString("  " + style(ansiGreen, "☑") + " " + inline(line[6:], opts) + "\n")
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

// isSQL reports whether a code fence language should get SQL colors.
func isSQL(lang string) bool {
	return lang == "sql" || lang == "psql" || lang == "postgresql" || lang == "plpgsql"
}

// countSteps returns the number of step headings outside code blocks.
func countSteps(lines []string) int {
	n := 0
	inCode := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCode = !inCode
			continue
		}
		if !inCode && stepHeading.MatchString(line) {
			n++
		}
	}
	return n
}

// stepHeader draws a boxed step title with a progress tracker, e.g.
//
//	────────────────────────────────────────
//	 1 ─ [2] ─ 3 ─ 4 ─ 5   STEP 2 OF 5
//	 WATCH A QUERY CRAWL
//	────────────────────────────────────────
func stepHeader(n, total int, title string, opts Options) string {
	style := func(code, s string) string {
		if !opts.Color {
			return s
		}
		return code + s + ansiReset
	}
	var marks []string
	for i := 1; i <= total; i++ {
		switch {
		case i == n:
			marks = append(marks, style(ansiBold+ansiInverse, "["+strconv.Itoa(i)+"]"))
		case i < n:
			marks = append(marks, style(ansiDim, strconv.Itoa(i)))
		default:
			marks = append(marks, strconv.Itoa(i))
		}
	}
	tracker := strings.Join(marks, style(ansiDim, " ─ "))

	var b strings.Builder
	b.WriteString("\n" + style(ansiDim, stepRule) + "\n")
	b.WriteString(" " + tracker + "   " + style(ansiBold, fmt.Sprintf("STEP %d OF %d", n, total)) + "\n")
	b.WriteString(" " + style(ansiBold, strings.ToUpper(inline(title, Options{}))) + "\n")
	b.WriteString(style(ansiDim, stepRule) + "\n")
	return b.String()
}

// inline strips light markdown emphasis; with color, `code` spans dim
// and **bold** spans bold.
func inline(s string, opts Options) string {
	s = spans(s, "`", ansiDim, opts.Color)
	s = spans(s, "**", ansiBold, opts.Color)
	return strings.ReplaceAll(s, "**", "") // stray, unpaired markers
}

// spans removes a paired emphasis marker, wrapping the enclosed text in
// code when color is on. Unbalanced markers are left untouched.
func spans(s, marker, code string, color bool) string {
	if !strings.Contains(s, marker) || strings.Count(s, marker)%2 != 0 {
		return s
	}
	parts := strings.Split(s, marker)
	var out strings.Builder
	for i, p := range parts {
		if i%2 == 1 && color {
			out.WriteString(code + p + ansiReset)
		} else {
			out.WriteString(p)
		}
	}
	return out.String()
}

// ShouldColor reports whether output should use ANSI colors, given
// whether stdout is a terminal and the value of $NO_COLOR.
func ShouldColor(isTTY bool, noColorEnv string) bool {
	return isTTY && noColorEnv == ""
}
