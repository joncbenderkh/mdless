// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

// With an explicit style and colour profile, headers must come back styled
// (ANSI escapes present) rather than as literal "#" text — the regression that
// autostyle-under-bubbletea caused.
func TestRenderStylesHeaders(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.TrueColor}
	out, err := render("# Big Header\n\nbody text\n", env, 80)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI-styled output, got plain:\n%s", out)
	}
	if strings.Contains(stripANSI(out), "# Big Header") {
		t.Fatalf("h1 still shows its literal marker:\n%s", stripANSI(out))
	}
}

func TestRenderUnknownStyleFallsBack(t *testing.T) {
	env := renderEnv{style: "no-such-style", profile: termenv.TrueColor}
	out, err := render("# Heading\n", env, 80)
	if err != nil {
		t.Fatalf("unknown style should fall back, not error: %v", err)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("fallback render should still be styled:\n%s", out)
	}
}

// readableStyle drops glamour's literal "## " / "### " heading markers and
// replaces them with a gutter bar so a heading never reads as list markup.
func TestRenderDropsHeadingMarkers(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.TrueColor}
	md := "## Second Level\n\ntext\n\n### 3. Third Level\n\nmore\n"
	out, err := render(md, env, 80)
	if err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(out)
	for _, marker := range []string{"## Second", "### 3.", "# "} {
		if strings.Contains(plain, marker) {
			t.Fatalf("heading marker %q still present:\n%s", marker, plain)
		}
	}
	// H2 gutter + upper-case, H3 gutter.
	if !strings.Contains(plain, majorGutter+"SECOND LEVEL") {
		t.Fatalf("H2 not rendered as %q...:\n%s", majorGutter, plain)
	}
	if !strings.Contains(plain, majorGutter+"3. Third Level") {
		t.Fatalf("H3 not rendered as %q...:\n%s", majorGutter, plain)
	}
}
