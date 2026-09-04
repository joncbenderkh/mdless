// SPDX-License-Identifier: GPL-3.0-or-later

// Package pager renders a Markdown document and drives an interactive,
// less-style terminal viewport over the rendered output.
package pager

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run renders markdown and blocks until the user quits the pager.
// title is shown in the header (a file path, or "(stdin)").
func Run(markdown, title string) error {
	p := tea.NewProgram(
		newModel(markdown, title),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
