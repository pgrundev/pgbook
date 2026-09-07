package render

import (
	"strings"
	"testing"
)

const sqlLesson = "```sql\n-- find one user\nSELECT count(*) FROM users WHERE email = 'a''b' AND id > 42;\n\\timing on\n```\n\n```bash\ndocker run postgres:17\n```\n"

func TestHighlightSQLColorsTokens(t *testing.T) {
	out := Render(sqlLesson, Options{Color: true})
	for _, want := range []string{
		ansiBold + ansiMagenta + "SELECT" + ansiReset, // keyword
		ansiBold + ansiMagenta + "FROM" + ansiReset,
		ansiCyan + "count" + ansiReset + "(",         // function call
		ansiGreen + "'a''b'" + ansiReset,             // string with escaped quote
		ansiYellow + "42" + ansiReset,                // number
		ansiDim + "-- find one user" + ansiReset,     // comment
		ansiYellow + "\\timing on" + ansiReset,       // psql meta-command
		"    " + ansiCyan + "docker run postgres:17", // non-SQL blocks stay cyan
	} {
		if !strings.Contains(out, want) {
			t.Errorf("sql render missing %q:\n%q", want, out)
		}
	}
	if strings.Contains(out, ansiCyan+"users") || strings.Contains(out, ansiBold+ansiMagenta+"users") {
		t.Errorf("identifier should be plain:\n%q", out)
	}
}

func TestHighlightSQLPlainIsUntouched(t *testing.T) {
	out := Render(sqlLesson, Options{Color: false})
	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain render contains ANSI escapes:\n%s", out)
	}
	if !strings.Contains(out, "    SELECT count(*) FROM users WHERE email = 'a''b' AND id > 42;") {
		t.Errorf("plain sql line altered:\n%s", out)
	}
}

func TestHighlightSQLKeywordsAreCaseInsensitive(t *testing.T) {
	out := highlightSQL("select 1")
	if !strings.HasPrefix(out, ansiBold+ansiMagenta+"select"+ansiReset) {
		t.Errorf("lowercase keyword not highlighted: %q", out)
	}
	if strings.Contains(highlightSQL("selected"), ansiMagenta) {
		t.Errorf("identifier with keyword prefix was highlighted")
	}
}
