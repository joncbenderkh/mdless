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
