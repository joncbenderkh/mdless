// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"strings"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }
func uintPtr(u uint) *uint    { return &u }

// Sentinels wrapped around the H1 and code blocks via their
// BlockPrefix/BlockSuffix. fillPanels finds the lines between a pair and paints
// a full-width background behind them; glamour cannot fill a block background
// itself. They use NUL so they can never collide with document text and are
// trivially stripped if one leaks. Every pair shares the "\x00mdless:" prefix.
const (
	codeOpen  = "\x00mdless:code\x00"
	codeClose = "\x00mdless:/code\x00"
	h1Open    = "\x00mdless:h1\x00"
	h1Close   = "\x00mdless:/h1\x00"
)

// baseStyleConfig returns the built-in glamour style named by name, or the dark
// style when the name is unknown.
func baseStyleConfig(name string) ansi.StyleConfig {
	if cfg, ok := styles.DefaultStyles[name]; ok && cfg != nil {
		return *cfg
	}
	return styles.DarkStyleConfig
}

// Heading gutter markers. A solid bar reads as structure at a glance and,
// unlike glamour's default "## " / "### ", never looks like list markup.
const (
	majorGutter = "▌ "
	midGutter   = "▍ "
	minorGutter = "▏ "
)

// readableStyle adapts a built-in glamour style for a full-screen pager.
//
// glamour's defaults are tuned for `glamour.Render` dumping a snippet into a
// normal scrollback buffer, so H2-H6 are printed with their literal "## " /
// "### " markers and headings sit flush against the preceding paragraph. In a
// pager that reads as clutter, and a heading like "### 1. Foo" is
// indistinguishable from an ordered-list item.
//
// The hierarchy is a strict fade from H1 down: a full-width banner, then
// bold-accent, accent, and three progressively dimmer plain levels. Markers are
// a coloured gutter bar (never "## ", which reads as list markup) and every
// heading gets a blank line above it.
func readableStyle(base ansi.StyleConfig, wrap int) ansi.StyleConfig {
	s := base

	s.Heading.BlockPrefix = "\n"

	// A rule that spans the page reads as a divider; glamour's default eight
	// dashes read as stray text.
	// glamour indents the rule by its two-column margin, so size it to wrap-2.
	s.HorizontalRule.Format = "\n" + strings.Repeat("─", wrap-2) + "\n"

	// glamour pads inline code with a literal space inside the styled span. Those
	// spaces make reflow treat the whole span as one unbreakable token, which
	// orphans it (or its neighbour) onto a line of its own. Drop the padding; the
	// background alone separates it from the surrounding words well enough.
	s.Code.Prefix = ""
	s.Code.Suffix = ""

	// H1 — full-width banner (filled by fillPanels), bold and upper-cased.
	s.H1 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		BlockPrefix: h1Open + "\n",
		BlockSuffix: "\n" + h1Close,
		Prefix:      "  ",
		Color:       strPtr("231"),
		Bold:        boolPtr(true),
		Upper:       boolPtr(true),
	}}
	// H2 — bold, upper-cased, underlined, bright accent, gutter bar.
	s.H2 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix:    majorGutter,
		Color:     strPtr("81"),
		Bold:      boolPtr(true),
		Upper:     boolPtr(true),
		Underline: boolPtr(true),
	}}
	// H3 — accent colour, a narrower gutter, regular weight and case.
	s.H3 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: midGutter,
		Bold:   boolPtr(false),
	}}
	// H4-H6 — no accent, thin gutter, a plain grey fading step by step. No
	// italic or faint: terminals without real italics render it as inverse,
	// which makes the deepest heading look the loudest.
	s.H4 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: minorGutter,
		Bold:   boolPtr(false),
		Color:  strPtr("250"),
	}}
	s.H5 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: minorGutter,
		Bold:   boolPtr(false),
		Color:  strPtr("244"),
	}}
	s.H6 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: minorGutter,
		Bold:   boolPtr(false),
		Color:  strPtr("238"),
	}}

	s.CodeBlock = panelCodeBlock(s.CodeBlock)

	return s
}

// panelCodeBlock brackets code blocks with the panel sentinels and gives the
// text a slightly brighter foreground. The background fill and the blank lines
// around it are applied afterwards by fillCodePanels.
func panelCodeBlock(cb ansi.StyleCodeBlock) ansi.StyleCodeBlock {
	cb.StylePrimitive.Color = strPtr("252")
	cb.StylePrimitive.BlockPrefix = codeOpen + "\n"
	cb.StylePrimitive.BlockSuffix = "\n" + codeClose
	cb.Margin = uintPtr(1)
	return cb
}
