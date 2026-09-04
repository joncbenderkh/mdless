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

// readableStyle adapts a built-in glamour style for a full-screen pager.
//
// glamour's defaults are tuned for `glamour.Render` dumping a snippet into a
// normal scrollback buffer, so H2-H6 are printed with their literal "## " /
// "### " markers and headings sit flush against the preceding paragraph. In a
// pager that reads as clutter. This keeps the H1 banner but drops the markers,
// adds a blank line above every heading, and leans on weight, an underline, and
// a faint italic to convey depth.
func readableStyle(base ansi.StyleConfig) ansi.StyleConfig {
	s := base

	s.Heading.BlockPrefix = "\n"

	s.H2 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Underline: boolPtr(true),
	}}
	s.H3 = ansi.StyleBlock{}
	s.H4 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Bold:   boolPtr(false),
		Italic: boolPtr(true),
	}}
	s.H5 = s.H4
	s.H6 = ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{
		Bold:   boolPtr(false),
		Italic: boolPtr(true),
		Faint:  boolPtr(true),
	}}

	return s
}
