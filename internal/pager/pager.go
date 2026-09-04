// SPDX-License-Identifier: GPL-3.0-or-later

// Package pager renders a Markdown document and drives an interactive,
// less-style terminal viewport over the rendered output.
package pager

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// renderEnv captures everything about the terminal that the Markdown renderer
// needs. It is resolved once, before the bubbletea program takes over the
// terminal — glamour's autostyle path queries the terminal for its background
// colour, and doing that while bubbletea holds stdin in raw mode yields a wrong
// answer (headers then fall back to the plain no-TTY style).
type renderEnv struct {
	style    string // resolved glamour base style ("dark", "light", …)
	profile  termenv.Profile
	theme    Theme
	maxWidth int // cap on the text measure; 0 = use the full terminal width
}

// Options configures a Run.
type Options struct {
	// Theme is a built-in name, a path to a JSON theme, or "" for the
	// terminal-appropriate default.
	Theme string
	// MaxWidth caps the text measure. 0 leaves it uncapped (full terminal).
	MaxWidth int
}

// DefaultMaxWidth is the text measure mdless wraps to unless told otherwise.
const DefaultMaxWidth = 72

// colored reports whether the terminal can render colour at all.
func (e renderEnv) colored() bool { return e.profile != termenv.Ascii }

// detectEnv resolves the theme (from --theme, else MDLESS_THEME, else the
// terminal background), the glamour base style it builds on, and the text
// measure.
func detectEnv(opts Options) (renderEnv, error) {
	dark := termenv.HasDarkBackground()

	themeSpec := opts.Theme
	if themeSpec == "" {
		themeSpec = os.Getenv("MDLESS_THEME")
	}
	if themeSpec == "" {
		if dark {
			themeSpec = "default"
		} else {
			themeSpec = "light"
		}
	}
	th, err := resolveTheme(themeSpec)
	if err != nil {
		return renderEnv{}, err
	}

	style := th.GlamourStyle
	if style == "" {
		style = os.Getenv("GLAMOUR_STYLE")
	}
	if style == "" {
		if dark {
			style = "dark"
		} else {
			style = "light"
		}
	}
	return renderEnv{
		style:    style,
		profile:  termenv.ColorProfile(),
		theme:    th,
		maxWidth: opts.MaxWidth,
	}, nil
}

// Run renders markdown and blocks until the user quits the pager. title is shown
// in the header (a file path, or "(stdin)").
func Run(markdown, title string, opts Options) error {
	env, err := detectEnv(opts)
	if err != nil {
		return err
	}
	m := newModel(markdown, title, env)

	// Pre-render at the current terminal size so the first frame is correct
	// and no glamour work happens after bubbletea grabs the terminal.
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		m.resize(w, h)
	}

	// Mouse reporting is left off so the terminal's own click-drag selection
	// keeps working (copy/paste). Most terminals still translate the wheel
	// into arrow keys under the alternate screen, so scrolling survives; `m`
	// toggles app-level mouse handling for the rest.
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	return err
}
