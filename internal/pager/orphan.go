// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	trailPadRE = regexp.MustCompile(`(?:\x1b\[[0-9;]*m|[ \t])+$`)
	leadPadRE  = regexp.MustCompile(`^(?:\x1b\[[0-9;]*m|[ \t])+`)
	escSeqRE   = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
)

// structuralGlyphs start a line that carries its own structure (a quote bar,
// bullet, heading gutter, rule, table pipe, link ref) — never merge across one.
var structuralGlyphs = []string{"│", "•", "▌", "▍", "▏", "─", "━", "🠶", "|", "["}

// unorphan pulls a lone word left on its own line down into the following line
// when it fits. glamour's reflow occasionally strands a word this way after it
// has had to break a very long token; contentWrap's multiple-of-three trick
// handles the common case but not that one. It runs on glamour's plain output,
// before fillPanels, and never touches a sentinel-bracketed block.
func unorphan(rendered string, wrap int) string {
	lines := strings.Split(rendered, "\n")
	out := make([]string, 0, len(lines))

	inBlock := false
	for i := 0; i < len(lines); i++ {
		cur := lines[i]
		if strings.Contains(cur, "\x00mdless:") {
			inBlock = strings.Contains(cur, codeOpen) || strings.Contains(cur, h1Open)
			out = append(out, cur)
			continue
		}
		if inBlock {
			out = append(out, cur)
			continue
		}

		word := strings.TrimSpace(stripANSI(cur))
		lone := word != "" && !strings.ContainsAny(word, " \t") && !structural(cur)
		if lone && i+1 < len(lines) {
			nxt := lines[i+1]
			if !strings.Contains(nxt, "\x00mdless:") &&
				strings.TrimSpace(stripANSI(nxt)) != "" && !structural(nxt) {
				if merged, ok := mergeWord(cur, nxt, wrap); ok {
					lines[i+1] = merged
					continue // drop the orphan line
				}
			}
		}
		out = append(out, cur)
	}
	return strings.Join(out, "\n")
}

func structural(line string) bool {
	v := strings.TrimLeft(stripANSI(line), " ")
	if v == "" {
		return true
	}
	for _, g := range structuralGlyphs {
		if strings.HasPrefix(v, g) {
			return true
		}
	}
	return v[0] >= '0' && v[0] <= '9' // ordered-list item
}

// mergeWord splices the single word on `orphan` onto the front of `next`,
// keeping their shared styling prefix. It reports false when the two lines
// aren't styled alike or the result would overflow.
func mergeWord(orphan, next string, wrap int) (string, bool) {
	p := safePrefixLen(orphan, next)
	if p == 0 {
		return "", false
	}
	prefix := orphan[:p]
	word := trailPadRE.ReplaceAllString(orphan[p:], "") // keeps its indent
	body := leadPadRE.ReplaceAllString(trailPadRE.ReplaceAllString(next[p:], ""), "")
	if strings.TrimSpace(stripANSI(word)) == "" || strings.TrimSpace(stripANSI(body)) == "" {
		return "", false
	}

	merged := prefix + word + " " + body
	if lipgloss.Width(merged) > wrap {
		return "", false
	}
	return merged, true
}

// safePrefixLen is the length of the longest common byte prefix of a and b that
// does not fall inside an ANSI escape sequence.
func safePrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	// If the boundary sits inside an unterminated escape, back up to its start.
	if tail := a[:i]; strings.LastIndexByte(tail, '\x1b') >= 0 {
		last := strings.LastIndexByte(tail, '\x1b')
		if !escSeqRE.MatchString(tail[last:]) {
			return last
		}
	}
	return i
}
