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

// readableStyle adapts a built-in glamour style for a full-screen pager, then
// applies the theme's headings, rules and code-panel colours.
//
// glamour's defaults are tuned for `glamour.Render` dumping a snippet into a
// normal scrollback buffer, so H2-H6 are printed with their literal "## " /
// "### " markers and headings sit flush against the preceding paragraph. In a
// pager that reads as clutter, and a heading like "### 1. Foo" is
// indistinguishable from an ordered-list item. The default theme fades the
// hierarchy strictly from an H1 banner down to a dim grey H6, with a gutter bar
// (never "## ") that narrows with depth.
func readableStyle(base ansi.StyleConfig, wrap int, th Theme) ansi.StyleConfig {
	s := base

	s.Heading.BlockPrefix = "\n"

	// A rule that spans the page reads as a divider; glamour's default eight
	// dashes read as stray text. glamour indents the rule by its two-column
	// margin, so size it to wrap-2.
	ruleChar := firstRune(th.RuleChar, "─")
	s.HorizontalRule.Format = "\n" + strings.Repeat(ruleChar, wrap-2) + "\n"
	if th.RuleColor != "" {
		s.HorizontalRule.Color = strPtr(th.RuleColor)
	}

	// glamour pads inline code with a literal space inside the styled span. Those
	// spaces make reflow treat the whole span as one unbreakable token, which
	// orphans it (or its neighbour) onto a line of its own. Drop the padding; the
	// background alone separates it from the surrounding words well enough.
	s.Code.Prefix = ""
	s.Code.Suffix = ""

	blocks := [...]*ansi.StyleBlock{&s.H1, &s.H2, &s.H3, &s.H4, &s.H5, &s.H6}
	for i, hs := range th.Headings {
		p := ansi.StylePrimitive{Prefix: hs.Gutter, Bold: boolPtr(hs.Bold)}
		if hs.FG != "" {
			p.Color = strPtr(hs.FG)
		}
		if hs.BG != "" {
			p.BackgroundColor = strPtr(hs.BG)
		}
		if hs.Underline {
			p.Underline = boolPtr(true)
		}
		if hs.Upper {
			p.Upper = boolPtr(true)
		}
		if hs.Italic {
			p.Italic = boolPtr(true)
		}
		if hs.Faint {
			p.Faint = boolPtr(true)
		}
		if i == 0 && hs.Banner {
			p.BlockPrefix = h1Open + "\n"
			p.BlockSuffix = "\n" + h1Close
		}
		*blocks[i] = ansi.StyleBlock{StylePrimitive: p}
	}

	s.CodeBlock = panelCodeBlock(s.CodeBlock, th.CodeFG)

	return s
}

// firstRune returns s if non-empty, else fallback — used to keep a one-glyph
// theme field from silently blanking a rule.
func firstRune(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// panelCodeBlock brackets code blocks with the panel sentinels and sets the
// text foreground. The background fill and the blank lines around it are applied
// afterwards by fillPanels.
func panelCodeBlock(cb ansi.StyleCodeBlock, fg string) ansi.StyleCodeBlock {
	cb.StylePrimitive.Color = strPtr(firstRune(fg, "252"))
	cb.StylePrimitive.BlockPrefix = codeOpen + "\n"
	cb.StylePrimitive.BlockSuffix = "\n" + codeClose
	cb.Margin = uintPtr(1)
	return cb
}
