// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// testEnv builds a renderEnv the way tests want it: the default theme, an
// explicit style and colour profile, bypassing terminal detection.
func testEnv(style string, profile termenv.Profile) renderEnv {
	return renderEnv{style: style, profile: profile, theme: baseTheme()}
}

// With an explicit style and colour profile, headers must come back styled
// (ANSI escapes present) rather than as literal "#" text — the regression that
// autostyle-under-bubbletea caused.
func TestRenderStylesHeaders(t *testing.T) {
	env := testEnv("dark", termenv.TrueColor)
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
	env := testEnv("no-such-style", termenv.TrueColor)
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
	env := testEnv("dark", termenv.TrueColor)
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
	// H2: major gutter + upper-case. H3: the narrower gutter, mixed case.
	if !strings.Contains(plain, baseTheme().Headings[1].Gutter+"SECOND LEVEL") {
		t.Fatalf("H2 not rendered as %q...:\n%s", baseTheme().Headings[1].Gutter, plain)
	}
	if !strings.Contains(plain, baseTheme().Headings[2].Gutter+"3. Third Level") {
		t.Fatalf("H3 not rendered as %q...:\n%s", baseTheme().Headings[2].Gutter, plain)
	}
}

// H1 is a full-width banner: wider and heavier than any lower level.
func TestRenderH1Banner(t *testing.T) {
	env := testEnv("dark", termenv.ANSI256)
	const width = 60
	wrap := contentWrap(width, 0)
	out, err := render("# Title Here\n\nbody\n\n## Section\n\nbody\n", env, width)
	if err != nil {
		t.Fatal(err)
	}
	var h1 string
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(stripANSI(ln), "TITLE HERE") { // H1 is upper-cased
			h1 = ln
		}
	}
	if h1 == "" {
		t.Fatalf("H1 not found / not upper-cased:\n%s", stripANSI(out))
	}
	if lipgloss.Width(h1) != wrap {
		t.Fatalf("H1 banner not full-width (%d, want %d): %q", lipgloss.Width(h1), wrap, stripANSI(h1))
	}
	if !strings.HasPrefix(h1, "\x1b[48;5;63m") {
		t.Fatalf("H1 missing banner background: %q", h1)
	}
}

// H2 and H3 must not read as near-identical: H2 is bold + underlined, H3 is not.
func TestRenderH2H3Distinct(t *testing.T) {
	env := testEnv("dark", termenv.ANSI256)
	out, err := render("## Section Two\n\nbody\n\n### Section Three\n\nbody\n", env, 80)
	if err != nil {
		t.Fatal(err)
	}
	var h2, h3 string
	for _, ln := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(stripANSI(ln), "SECTION TWO"): // H2 is upper-cased
			h2 = ln
		case strings.Contains(stripANSI(ln), "Section Three"):
			h3 = ln
		}
	}
	if h2 == "" || h3 == "" {
		t.Fatalf("missing a heading:\n%s", stripANSI(out))
	}
	// termenv fuses the params, e.g. "38;5;81;4;1m" — bold (1) and underline (4).
	sgr := regexp.MustCompile(`\x1b\[([0-9;]*)m`)
	hasAttr := func(line string, attr string) bool {
		for _, m := range sgr.FindAllStringSubmatch(line, -1) {
			for _, p := range strings.Split(m[1], ";") {
				if p == attr {
					return true
				}
			}
		}
		return false
	}
	if !hasAttr(h2, "1") || !hasAttr(h2, "4") {
		t.Fatalf("H2 should be bold (1) and underlined (4): %q", h2)
	}
	if hasAttr(h3, "1") || hasAttr(h3, "4") {
		t.Fatalf("H3 should be neither bold nor underlined: %q", h3)
	}
	if !strings.Contains(stripANSI(h3), baseTheme().Headings[2].Gutter) {
		t.Fatalf("H3 should use the narrower gutter %q: %q", baseTheme().Headings[2].Gutter, stripANSI(h3))
	}
}

// H4-H6 must be visually distinct from each other, not just from the body.
func TestRenderHeadingLevelsDiffer(t *testing.T) {
	env := testEnv("dark", termenv.TrueColor)
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
	env := testEnv("dark", termenv.ANSI256)
	const width = 50
	wrap := contentWrap(width, 0)
	md := "A paragraph with `inline code` that should stay prose and wrap " +
		"naturally without being widened into a panel at all.\n\n" +
		"```\nplain block line\n```\n\n```go\nfunc x() {}\n```\n"

	out, err := render(md, env, width)
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

// A wrap width that is a multiple of three keeps glamour's reflow from wasting
// columns and orphaning a word onto its own line.
func TestContentWrapAvoidsOrphans(t *testing.T) {
	env := testEnv("dark", termenv.ANSI256)
	// A paragraph whose natural break lands mid-phrase at several widths.
	md := "The lines below fade from a bold upper-case accent at H2 all the " +
		"way down to plain dim grey at H6, one gutter bar for every level.\n"
	for _, width := range []int{78, 79, 80, 92, 100} {
		out, err := render(md, env, width)
		if err != nil {
			t.Fatal(err)
		}
		for _, ln := range strings.Split(out, "\n") {
			words := strings.Fields(stripANSI(ln))
			if len(words) == 1 && !strings.HasPrefix(words[0], "─") {
				t.Errorf("width %d: orphaned word %q", width, words[0])
			}
		}
	}
}

func TestContentWrapCap(t *testing.T) {
	cases := []struct{ term, max, want int }{
		{120, 72, 72}, // capped, 72 is already a multiple of 3
		{100, 72, 72}, // capped
		{60, 72, 60},  // terminal narrower than the cap
		{200, 0, 198}, // uncapped: rounds 200 down to a multiple of 3
		{10, 72, 21},  // floor
		{80, 100, 78}, // cap wider than the terminal → terminal wins, rounded
	}
	for _, c := range cases {
		if got := contentWrap(c.term, c.max); got != c.want {
			t.Errorf("contentWrap(%d, %d) = %d, want %d", c.term, c.max, got, c.want)
		}
	}
}

func TestExpandTabs(t *testing.T) {
	cases := map[string]string{
		"\tx":               "    x",
		"a\tb":              "a   b",
		"ab\tc":             "ab  c",
		"abcd\te":           "abcd    e",
		"\x1b[1m\tx\x1b[0m": "\x1b[1m    x\x1b[0m", // escape does not advance column
		"line1\n\tline2":    "line1\n    line2",    // column resets after newline
	}
	for in, want := range cases {
		if got := expandTabs(in, 4); got != want {
			t.Errorf("expandTabs(%q) = %q, want %q", in, got, want)
		}
	}
}

// A code block indented with tabs must not gain blank rows: unexpanded tabs
// overflow the width maths and wrap in the viewport.
func TestRenderTabbedCodeNoBlankRows(t *testing.T) {
	env := testEnv("dark", termenv.ANSI256)
	md := "```go\nfunc f() {\n\ta := 1\n\tb := 2\n\treturn a + b\n}\n```\n"
	out, err := render(md, env, 60)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\t") {
		t.Fatal("tab survived into rendered output")
	}
	inPanel := false
	for _, ln := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(ln, "\x1b[48;5;236m"):
			inPanel = true
			if strings.TrimSpace(stripANSI(ln)) == "" {
				t.Fatalf("blank row inside a tabbed code panel:\n%s", stripANSI(out))
			}
		case inPanel && strings.TrimSpace(stripANSI(ln)) == "":
			inPanel = false // one trailing edge blank is fine
		}
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
			env := testEnv(style, termenv.TrueColor)
			out, err := render(string(src), env, width)
			if err != nil {
				t.Fatalf("render(style=%s width=%d): %v", style, width, err)
			}
			if strings.TrimSpace(stripANSI(out)) == "" {
				t.Fatalf("render(style=%s width=%d): empty output", style, width)
			}
			// Code panels are ours to size; at usable widths they must not
			// overflow (a tab or missed width calc would wrap them). Below that,
			// an unbreakable code token can overflow — that is glamour's reflow,
			// not our padding.
			if width >= 40 {
				for _, ln := range strings.Split(out, "\n") {
					plain := stripANSI(ln)
					panel := strings.HasPrefix(ln, "\x1b[48;5;236m") ||
						strings.HasPrefix(ln, "\x1b[48;5;254m") ||
						strings.HasPrefix(ln, "\x1b[48;5;63m")
					rule := strings.Trim(plain, "─━ ") == "" && strings.TrimSpace(plain) != ""
					if (panel || rule) && lipgloss.Width(ln) > width {
						t.Fatalf("render(style=%s width=%d): line overflows (%d): %q",
							style, width, lipgloss.Width(ln), plain)
					}
				}
			}
		}
	}
}
