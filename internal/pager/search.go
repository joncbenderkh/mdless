// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ansiPattern matches CSI escape sequences emitted by the renderer.
var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

// updateSearchInput consumes keystrokes while the `/` prompt is open.
func (m model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.searching = false
		m.query = ""
		m.matches = nil
		return m, nil
	case "enter":
		m.searching = false
		m.recomputeMatches()
		if len(m.matches) > 0 {
			m.matchCursor = m.nearestMatch()
			m.scrollToMatch()
		}
		return m, nil
	case "backspace":
		if m.query != "" {
			m.query = m.query[:len(m.query)-1]
		}
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.query += string(msg.Runes)
		}
		return m, nil
	}
}

func (m *model) recomputeMatches() {
	m.matches = m.matches[:0]
	if m.query == "" {
		return
	}
	needle := strings.ToLower(m.query)
	for i, line := range m.plainLines {
		if strings.Contains(strings.ToLower(line), needle) {
			m.matches = append(m.matches, i)
		}
	}
	if m.matchCursor >= len(m.matches) {
		m.matchCursor = 0
	}
}

// nearestMatch returns the index of the first match at or after the current
// scroll position, wrapping to the top if none is below.
func (m model) nearestMatch() int {
	top := m.viewport.YOffset
	for i, line := range m.matches {
		if line >= top {
			return i
		}
	}
	return 0
}

func (m *model) jumpMatch(dir int) {
	if len(m.matches) == 0 {
		return
	}
	m.matchCursor = (m.matchCursor + dir + len(m.matches)) % len(m.matches)
	m.scrollToMatch()
}

func (m *model) scrollToMatch() {
	if len(m.matches) == 0 {
		return
	}
	target := m.matches[m.matchCursor] - m.viewport.Height/3
	if target < 0 {
		target = 0
	}
	m.viewport.SetYOffset(target)
}
