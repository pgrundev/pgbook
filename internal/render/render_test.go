package render

import (
	"strings"
	"testing"
)

const lesson = "## Row locks vs table locks\n\nA paragraph of text.\n\n- first item\n- second item\n\n> Note: locks are released at commit.\n\n```sql\nSELECT * FROM accounts FOR UPDATE;\n```\n"

func TestRenderPlainHasNoANSI(t *testing.T) {
	out := Render(lesson, Options{Color: false})
	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain render contains ANSI escapes:\n%s", out)
	}
	for _, want := range []string{
		"ROW LOCKS VS TABLE LOCKS",
		"A paragraph of text.",
		"• first item",
		"Note: locks are released at commit.",
		"SELECT * FROM accounts FOR UPDATE;",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plain render missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "```") {
		t.Errorf("code fences leaked into output:\n%s", out)
	}
	if strings.Contains(out, "## ") {
		t.Errorf("heading markers leaked into output:\n%s", out)
	}
}

func TestRenderColorUsesANSIAndResets(t *testing.T) {
	out := Render(lesson, Options{Color: true})
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("color render has no ANSI escapes:\n%s", out)
	}
	if !strings.HasSuffix(strings.TrimRight(out, "\n"), "\x1b[0m") && !strings.Contains(out, "\x1b[0m") {
		t.Errorf("color render never resets attributes")
	}
}

func TestRenderIndentsCodeBlocks(t *testing.T) {
	out := Render(lesson, Options{Color: false})
	if !strings.Contains(out, "    SELECT * FROM accounts FOR UPDATE;") {
		t.Errorf("code block not indented:\n%s", out)
	}
}

func TestShouldColor(t *testing.T) {
	if ShouldColor(true, "1") {
		t.Error("NO_COLOR set: ShouldColor = true, want false")
	}
	if ShouldColor(false, "") {
		t.Error("not a TTY: ShouldColor = true, want false")
	}
	if !ShouldColor(true, "") {
		t.Error("TTY without NO_COLOR: ShouldColor = false, want true")
	}
}
