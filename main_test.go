package main

import (
	"bytes"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}, {"help"}, {}} {
		code, out, _ := runCLI(t, args...)
		if code != 0 {
			t.Errorf("%v: exit %d, want 0", args, code)
		}
		for _, want := range []string{"pgbook", "list", "read", "search", "next", "pdf"} {
			if !strings.Contains(out, want) {
				t.Errorf("%v: help missing %q:\n%s", args, want, out)
			}
		}
	}
}

func TestVersion(t *testing.T) {
	code, out, _ := runCLI(t, "--version")
	if code != 0 {
		t.Errorf("--version: exit %d, want 0", code)
	}
	if !strings.Contains(out, "pgbook") || !strings.Contains(out, version) {
		t.Errorf("--version output = %q", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, _, errOut := runCLI(t, "frobnicate")
	if code == 0 {
		t.Error("unknown command exited 0")
	}
	if !strings.Contains(errOut, "frobnicate") {
		t.Errorf("stderr should name the unknown command: %q", errOut)
	}
}

func TestReadRequiresTopic(t *testing.T) {
	code, _, errOut := runCLI(t, "read")
	if code == 0 {
		t.Error("read without a topic exited 0")
	}
	if !strings.Contains(errOut, "pgbook read") {
		t.Errorf("stderr should show usage: %q", errOut)
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	code, _, _ := runCLI(t, "search")
	if code == 0 {
		t.Error("search without a query exited 0")
	}
}

func TestAPIErrorIsReportedNotPanic(t *testing.T) {
	t.Setenv("PGBOOK_BASE_URL", "http://127.0.0.1:1") // nothing listens here
	code, _, errOut := runCLI(t, "list")
	if code == 0 {
		t.Error("list with unreachable API exited 0")
	}
	if errOut == "" {
		t.Error("no error message on stderr")
	}
}
