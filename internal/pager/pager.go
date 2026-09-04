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
	style   string
	profile termenv.Profile
}

func detectEnv() renderEnv {
	style := os.Getenv("GLAMOUR_STYLE")
	if style == "" {
		if termenv.HasDarkBackground() {
			style = "dark"
		} else {
			style = "light"
		}
	}
	return renderEnv{style: style, profile: termenv.ColorProfile()}
}

// Run renders markdown and blocks until the user quits the pager.
// title is shown in the header (a file path, or "(stdin)").
func Run(markdown, title string) error {
	m := newModel(markdown, title, detectEnv())

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
	_, err := p.Run()
	return err
}
