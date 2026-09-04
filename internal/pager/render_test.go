// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

// H4-H6 must be visually distinct from each other, not just from the body.
func TestRenderHeadingLevelsDiffer(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.TrueColor}
	md := "#### Four\n\n##### Five\n\n###### Six\n"
	out, err := render(md, env, 80)
	if err != nil {
		t.Fatal(err)
	}
	lines := map[string]string{}
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(stripANSI(l), "Four"):
			lines["4"] = l
		case strings.Contains(stripANSI(l), "Five"):
			lines["5"] = l
		case strings.Contains(stripANSI(l), "Six"):
			lines["6"] = l
		}
	}
	if lines["4"] == "" || lines["5"] == "" || lines["6"] == "" {
		t.Fatalf("missing a heading line: %#v", lines)
	}
	norm := func(s string) string { return strings.TrimRight(stripANSI(s), " ") }
	if norm(lines["4"]) == norm(lines["5"]) || norm(lines["5"]) == norm(lines["6"]) {
		t.Fatalf("H4-H6 not distinguishable:\n4:%q\n5:%q\n6:%q",
			norm(lines["4"]), norm(lines["5"]), norm(lines["6"]))
	}
}

// Code blocks become a solid, full-width background panel; inline code inside a
// paragraph must not trigger the panel treatment.
func TestRenderCodePanels(t *testing.T) {
	env := renderEnv{style: "dark", profile: termenv.ANSI256}
	const wrap = 48
	md := "A paragraph with `inline code` that should stay prose and wrap " +
		"naturally without being widened into a panel at all.\n\n" +
		"```\nplain block line\n```\n\n```go\nfunc x() {}\n```\n"

	out, err := render(md, env, wrap+2)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\x00mdless") {
		t.Fatalf("sentinel leaked into output:\n%q", out)
	}

	var codeLines, proseInlineCode int
	for _, ln := range strings.Split(out, "\n") {
		plain := stripANSI(ln)
		switch {
		case strings.Contains(plain, "plain block line"), strings.Contains(plain, "func x()"):
			codeLines++
			if lipgloss.Width(ln) != wrap {
				t.Fatalf("code line not padded to wrap %d (got %d): %q", wrap, lipgloss.Width(ln), plain)
			}
			if !strings.HasPrefix(ln, "\x1b[48;5;236m") {
				t.Fatalf("code line missing panel background: %q", ln)
			}
		case strings.Contains(plain, "inline code"):
			proseInlineCode++
			if lipgloss.Width(ln) >= wrap {
				t.Fatalf("prose line with inline code was widened to a panel: %q", plain)
			}
		}
	}
	if codeLines != 2 {
		t.Fatalf("expected 2 painted code lines, saw %d", codeLines)
	}
	if proseInlineCode == 0 {
		t.Fatal("did not find the inline-code prose line")
	}
}

// The showcase document must render cleanly at any width — it is the manual
// regression fixture for style changes, so a parser or style panic here should
// fail the build.
func TestRenderShowcase(t *testing.T) {
	src, err := os.ReadFile("../../testdata/showcase.md")
	if err != nil {
		t.Fatalf("read showcase: %v", err)
	}
	for _, style := range []string{"dark", "light", "notty"} {
		for _, width := range []int{20, 40, 80, 200} {
			env := renderEnv{style: style, profile: termenv.TrueColor}
			out, err := render(string(src), env, width)
			if err != nil {
				t.Fatalf("render(style=%s width=%d): %v", style, width, err)
			}
			if strings.TrimSpace(stripANSI(out)) == "" {
				t.Fatalf("render(style=%s width=%d): empty output", style, width)
			}
		}
	}
}
