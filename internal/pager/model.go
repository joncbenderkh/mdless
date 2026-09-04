// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight = 1
	footerHeight = 1
)

var (
	chromeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("252"))
	matchCountStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

type model struct {
	title string
	raw   string
	env   renderEnv

	viewport  viewport.Model
	ready     bool
	renderErr error

	// plainLines mirrors the rendered viewport content with ANSI stripped,
	// so search can match on visible text and map back to a line offset.
	plainLines []string

	searching   bool
	query       string
	matches     []int
	matchCursor int

	mouseCapture bool
}

func newModel(markdown, title string, env renderEnv) model {
	return model{title: title, raw: markdown, env: env}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m.updateSearchInput(msg)
		}
		if cmd, handled := m.handleKey(msg); handled {
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *model) resize(width, height int) {
	body := height - headerHeight - footerHeight
	if body < 1 {
		body = 1
	}

	if !m.ready {
		m.viewport = viewport.New(width, body)
	} else {
		m.viewport.Width = width
		m.viewport.Height = body
	}

	rendered, err := render(m.raw, m.env, width)
	m.renderErr = err
	m.viewport.SetContent(rendered)
	m.plainLines = strings.Split(stripANSI(rendered), "\n")

	if m.query != "" {
		m.recomputeMatches()
	}
	m.ready = true
}

// contentWrap is the wrap width handed to glamour. Its reflow only ever uses a
// width that is a multiple of three (minus its own two-column margin); a
// non-multiple wastes up to two columns and drops the odd word onto a line of
// its own. Rounding down to a multiple of three makes it wrap to the full width.
func contentWrap(width int) int {
	wrap := width - width%3
	if wrap < 21 {
		wrap = 21
	}
	return wrap
}

func render(markdown string, env renderEnv, width int) (string, error) {
	wrap := contentWrap(width)
	r, err := newRenderer(env, wrap)
	if err != nil {
		return markdown, err
	}
	out, err := r.Render(markdown)
	if err != nil {
		return markdown, err
	}
	// glamour leaves tabs (common in code blocks) unexpanded. The terminal
	// renders them 8 wide while our width maths counts them as 1, so a tabbed
	// line overflows and wraps — showing as spurious blank rows. Expand first.
	out = expandTabs(out, 4)
	return strings.TrimRight(fillPanels(out, env, wrap), "\n"), nil
}

// expandTabs replaces tab characters with spaces to the next tab stop, counting
// visible columns only (ANSI escape sequences do not advance the column).
func expandTabs(s string, tabWidth int) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 16)
	col, inEsc := 0, false
	for _, r := range s {
		switch {
		case inEsc:
			b.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
		case r == '\x1b':
			inEsc = true
			b.WriteRune(r)
		case r == '\n':
			b.WriteRune(r)
			col = 0
		case r == '\t':
			n := tabWidth - col%tabWidth
			for i := 0; i < n; i++ {
				b.WriteByte(' ')
			}
			col += n
		default:
			b.WriteRune(r)
			col++
		}
	}
	return b.String()
}

const ansiReset = "\x1b[0m"

// fillPanels turns the sentinel-bracketed regions (the H1 and every code block,
// see readableStyle) into solid full-width panels: each line between a sentinel
// pair gets its background painted edge to edge, re-asserted after every inline
// reset that glamour's syntax highlighting emits. glamour cannot fill a block
// background on its own, so an unhighlighted code block would otherwise look
// like prose and the H1 banner would not stand out.
func fillPanels(rendered string, env renderEnv, wrap int) string {
	if !env.colored() {
		return stripSentinels(rendered) // no colour to paint
	}
	codeBG, h1BG := "\x1b[48;5;236m", "\x1b[48;5;63m"
	if env.style == "light" || env.style == "pink" {
		codeBG, h1BG = "\x1b[48;5;254m", "\x1b[48;5;62m"
	}

	isBlank := func(s string) bool { return strings.TrimSpace(stripANSI(s)) == "" }

	var out []string
	blankTail := func() bool { return len(out) == 0 || isBlank(out[len(out)-1]) }
	edge := func() {
		if !blankTail() {
			out = append(out, "")
		}
	}
	paint := func(ln, bg string) string {
		switch w := lipgloss.Width(ln); {
		case w < wrap:
			ln += strings.Repeat(" ", wrap-w) // pad short lines to the panel edge
		case w > wrap:
			ln = truncateVisible(ln, wrap) // trim glamour's over-eager padding
		}
		return bg + strings.ReplaceAll(ln, ansiReset, ansiReset+bg) + ansiReset
	}
	flush := func(lines []string, bg string) {
		for len(lines) > 0 && isBlank(lines[len(lines)-1]) {
			lines = lines[:len(lines)-1] // drop trailing blank lines in the panel
		}
		for _, l := range lines {
			out = append(out, paint(l, bg))
		}
	}

	var panel []string
	var panelBG string
	inPanel := false
	for _, ln := range strings.Split(rendered, "\n") {
		switch {
		case strings.Contains(ln, codeOpen):
			inPanel, panelBG, panel = true, codeBG, panel[:0]
			edge()
		case strings.Contains(ln, h1Open):
			inPanel, panelBG, panel = true, h1BG, panel[:0]
			edge()
		case strings.Contains(ln, codeClose), strings.Contains(ln, h1Close):
			inPanel = false
			flush(panel, panelBG)
			edge()
		case inPanel:
			panel = append(panel, ln)
		case isBlank(ln):
			edge() // collapse blank runs outside panels
		default:
			out = append(out, ln)
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

// truncateVisible drops visible columns past max while keeping every ANSI
// escape sequence, so styling (and any trailing reset) survives the cut.
func truncateVisible(s string, max int) string {
	var b strings.Builder
	col, inEsc := 0, false
	for _, r := range s {
		switch {
		case inEsc:
			b.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
		case r == '\x1b':
			inEsc = true
			b.WriteRune(r)
		case col < max:
			b.WriteRune(r)
			col++
		}
	}
	return b.String()
}

func stripSentinels(s string) string {
	var b strings.Builder
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, "\x00mdless:") {
			continue
		}
		b.WriteString(ln)
		b.WriteString("\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// newRenderer builds a glamour renderer with an explicit style and colour
// profile (never autostyle). The style is a built-in one adapted by
// readableStyle for full-screen use.
func newRenderer(env renderEnv, wrap int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStyles(readableStyle(baseStyleConfig(env.style), wrap)),
		glamour.WithColorProfile(env.profile),
		glamour.WithWordWrap(wrap),
	)
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "g", "home":
		m.viewport.GotoTop()
		return nil, true
	case "G", "end":
		m.viewport.GotoBottom()
		return nil, true
	case "d", "ctrl+d":
		m.viewport.HalfPageDown()
		return nil, true
	case "u", "ctrl+u":
		m.viewport.HalfPageUp()
		return nil, true
	case "b", "pgup":
		m.viewport.PageUp()
		return nil, true
	case " ", "pgdown":
		m.viewport.PageDown()
		return nil, true
	case "j":
		m.viewport.ScrollDown(1)
		return nil, true
	case "k":
		m.viewport.ScrollUp(1)
		return nil, true
	case "/":
		m.searching = true
		m.query = ""
		return nil, true
	case "m":
		m.mouseCapture = !m.mouseCapture
		if m.mouseCapture {
			return tea.EnableMouseCellMotion, true
		}
		return tea.DisableMouse, true
	case "n":
		m.jumpMatch(1)
		return nil, true
	case "N":
		m.jumpMatch(-1)
		return nil, true
	}
	return nil, false
}

func (m model) View() string {
	if !m.ready {
		return "loading…"
	}
	return strings.Join([]string{
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	}, "\n")
}

func (m model) headerView() string {
	title := m.title
	if m.renderErr != nil {
		title += errorStyle.Render("  (render error — showing raw source)")
	}
	return chromeStyle.Width(m.viewport.Width).Render(" " + title)
}

func (m model) footerView() string {
	if m.searching {
		return chromeStyle.Width(m.viewport.Width).Render("/" + m.query)
	}

	var left string
	if m.viewport.TotalLineCount() <= m.viewport.VisibleLineCount() {
		left = "  All"
	} else {
		left = fmt.Sprintf(" %3.0f%%", m.viewport.ScrollPercent()*100)
	}
	if m.query != "" {
		if len(m.matches) == 0 {
			left += matchCountStyle.Render(fmt.Sprintf("  /%s  no matches", m.query))
		} else {
			left += matchCountStyle.Render(fmt.Sprintf("  /%s  %d/%d",
				m.query, m.matchCursor+1, len(m.matches)))
		}
	}
	mouse := "m:select"
	if m.mouseCapture {
		mouse = "m:scroll"
	}
	right := "/ search · " + mouse + " · q quit "
	gap := m.viewport.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return chromeStyle.Width(m.viewport.Width).
		Render(left + strings.Repeat(" ", gap) + right)
}
