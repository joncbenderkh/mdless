// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

func boolPtr(b bool) *bool { return &b }

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
		Bold:   boolPtr(true),
	}}
	s.H4 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: minorGutter,
		Bold:   boolPtr(false),
		Italic: boolPtr(true),
	}}
	s.H5 = s.H4
	s.H6 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Prefix: minorGutter,
		Bold:   boolPtr(false),
		Italic: boolPtr(true),
		Faint:  boolPtr(true),
	}}

	return s
}
