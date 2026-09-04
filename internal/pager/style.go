// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }
func uintPtr(u uint) *uint    { return &u }

// Sentinels wrapped around every code block via its BlockPrefix/BlockSuffix.
// fillCodePanels finds the lines between them and paints the panel background;
// glamour cannot fill a block background itself. They use NUL so they can never
// collide with document text and are trivially stripped if one leaks.
const (
	codeOpen  = "\x00mdless:code\x00"
	codeClose = "\x00mdless:/code\x00"
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
// This keeps the H1 banner, replaces the markers with a coloured gutter bar,
// adds a blank line above every heading, upper-cases H2 so top-level sections
// stand apart, and fades H4-H6.
func readableStyle(base ansi.StyleConfig) ansi.StyleConfig {
	s := base

	s.Heading.BlockPrefix = "\n"

	s.H2 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: majorGutter,
		Bold:   boolPtr(true),
		Upper:  boolPtr(true),
	}}
	s.H3 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: majorGutter,
		Bold:   boolPtr(false),
	}}
	// H4-H6 are rare; distinguish them by indenting the gutter one step per
	// level and layering on italic, then faint.
	s.H4 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: minorGutter,
		Bold:   boolPtr(false),
	}}
	s.H5 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: "  " + minorGutter,
		Bold:   boolPtr(false),
		Italic: boolPtr(true),
	}}
	s.H6 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: "    " + minorGutter,
		Bold:   boolPtr(false),
		Italic: boolPtr(true),
		Faint:  boolPtr(true),
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
