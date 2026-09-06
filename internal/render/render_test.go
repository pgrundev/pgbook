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

const tutorial = "## Why some queries are instant\n\nIntro.\n\n## Step 1: Get a Postgres to play with\n\nText.\n\n### Your turn\n\n- [ ] Start Postgres.\n- [x] Open psql.\n\n## Step 2: Watch a query crawl\n\nMore.\n\n```sql\n## Step 9: not a real step, inside a code block\n```\n\n## Step 3: Add an index\n\n## Step 4: When the index does not help\n\n## Step 5: Indexes are not free\n\n## What you learned\n"

func TestRenderStepHeadersShowProgress(t *testing.T) {
	out := Render(tutorial, Options{Color: false})
	for _, want := range []string{
		"STEP 1 OF 5",
		"GET A POSTGRES TO PLAY WITH",
		"[1] ─ 2 ─ 3 ─ 4 ─ 5",
		"STEP 2 OF 5",
		"1 ─ [2] ─ 3 ─ 4 ─ 5",
		"STEP 5 OF 5",
		"1 ─ 2 ─ 3 ─ 4 ─ [5]",
		"WHAT YOU LEARNED",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("step render missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "STEP 9") || strings.Contains(out, "OF 6") {
		t.Errorf("step heading inside a code block was counted as a step:\n%s", out)
	}
	if strings.Contains(out, "## Step 1") || strings.Contains(out, "Step 1:") {
		t.Errorf("raw step heading leaked into output:\n%s", out)
	}
}

func TestRenderStepHeadersColorResets(t *testing.T) {
	out := Render(tutorial, Options{Color: true})
	if !strings.Contains(out, "STEP 2 OF 5") {
		t.Fatalf("color step render missing header:\n%s", out)
	}
	if strings.Count(out, "\x1b[0m") == 0 {
		t.Errorf("color step render never resets attributes")
	}
}

func TestRenderChecklist(t *testing.T) {
	out := Render(tutorial, Options{Color: false})
	if !strings.Contains(out, "☐ Start Postgres.") {
		t.Errorf("unchecked item not rendered as ☐:\n%s", out)
	}
	if !strings.Contains(out, "☑ Open psql.") {
		t.Errorf("checked item not rendered as ☑:\n%s", out)
	}
	if strings.Contains(out, "[ ]") || strings.Contains(out, "[x]") {
		t.Errorf("raw checkbox markers leaked into output:\n%s", out)
	}
}

func TestRenderBoldSpans(t *testing.T) {
	plain := Render("**1. A function** around the column.\n", Options{Color: false})
	if !strings.Contains(plain, "1. A function around the column.") || strings.Contains(plain, "**") {
		t.Errorf("plain bold span wrong: %q", plain)
	}
	color := Render("**1. A function** around the column.\n", Options{Color: true})
	if !strings.Contains(color, "\x1b[1m1. A function\x1b[0m") {
		t.Errorf("color bold span wrong: %q", color)
	}
}

func TestBanner(t *testing.T) {
	plain := Banner(false)
	if strings.Contains(plain, "\x1b[") {
		t.Errorf("plain banner contains ANSI escapes:\n%s", plain)
	}
	if !strings.Contains(plain, "pgbook.dev") {
		t.Errorf("banner should name the site:\n%s", plain)
	}
	for _, line := range strings.Split(plain, "\n") {
		if n := len([]rune(line)); n > 80 {
			t.Errorf("banner line is %d columns, want <= 80: %q", n, line)
		}
	}
	color := Banner(true)
	if !strings.Contains(color, "\x1b[") || !strings.Contains(color, "\x1b[0m") {
		t.Errorf("color banner has no ANSI escapes or never resets:\n%s", color)
	}
}
