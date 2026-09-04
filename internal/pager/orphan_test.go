// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"regexp"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func TestUnorphanPullsWordDown(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.ANSI256, theme: baseTheme()}
	// glamour strands "resumes" on its own line after breaking the long token.
	md := "aaaa. A-really-long-hyphenated-token-that-must-be-split-across-lines-" +
		"for-sure follows, then normal prose resumes to confirm the wrap is ok.\n"
	out, err := render(md, env, 72)
	if err != nil {
		t.Fatal(err)
	}
	for _, ln := range strings.Split(out, "\n") {
		words := strings.Fields(stripANSI(ln))
		if len(words) == 1 && !strings.ContainsAny(words[0], "─━") {
			t.Fatalf("word left orphaned: %q\n---\n%s", words[0], stripANSI(out))
		}
	}
}

func TestUnorphanLeavesBlocksAlone(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.ANSI256, theme: baseTheme()}
	// A code block whose last line is a single "}" must not be merged, and its
	// escape sequences must not be sliced.
	md := "text\n\n```go\nfunc f() {\n\treturn\n}\n```\n\nmore text here after it\n"
	out, err := render(md, env, 72)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\x00mdless") {
		t.Fatalf("sentinel leaked: %q", out)
	}
	// A sliced escape leaves a bare "<params>m<letter>" once real escapes are
	// stripped — e.g. "251mreturn".
	if sliced := regexp.MustCompile(`\d;?\d*m[A-Za-z]`).FindString(stripANSI(out)); sliced != "" {
		t.Fatalf("sliced escape sequence near %q\n---\n%s", sliced, stripANSI(out))
	}
	if !strings.Contains(stripANSI(out), "return") || !strings.Contains(stripANSI(out), "}") {
		t.Fatalf("code body damaged:\n%s", stripANSI(out))
	}
}

func TestUnorphanKeepsListsAndQuotes(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.ANSI256, theme: baseTheme()}
	md := "- a bullet item\n- another\n\n> a quote line\n> and more\n"
	out, err := render(md, env, 72)
	if err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(out)
	if !strings.Contains(plain, "• a bullet item") || !strings.Contains(plain, "│ a quote line") {
		t.Fatalf("structure disturbed:\n%s", plain)
	}
}
