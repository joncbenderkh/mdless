// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	fnDefRE = regexp.MustCompile(`^\[\^([^\]]+)\]:[ \t]*(.*)$`)
	fnRefRE = regexp.MustCompile(`\[\^([^\]]+)\]`)
	fenceRE = regexp.MustCompile("^\\s*(```|~~~)")
)

// expandFootnotes rewrites `text[^id]` … `[^id]: definition` into numbered
// references and a "Footnotes" section at the end. glamour's own footnote
// support is registered but unimplemented, so this keeps them readable without
// fighting it. Content inside fenced code blocks is left alone.
func expandFootnotes(md string) string {
	lines := strings.Split(md, "\n")

	// Pass 1: pull out the definitions (with indented continuation lines).
	defs := map[string]string{}
	var body []string
	inFence := false
	for i := 0; i < len(lines); i++ {
		if fenceRE.MatchString(lines[i]) {
			inFence = !inFence
			body = append(body, lines[i])
			continue
		}
		if m := fnDefRE.FindStringSubmatch(lines[i]); !inFence && m != nil {
			text := m[2]
			for i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" &&
				(strings.HasPrefix(lines[i+1], "  ") || strings.HasPrefix(lines[i+1], "\t")) {
				i++
				text += " " + strings.TrimSpace(lines[i])
			}
			defs[m[1]] = strings.TrimSpace(text)
			continue
		}
		body = append(body, lines[i])
	}
	if len(defs) == 0 {
		return md
	}

	// Pass 2: number the references in order of first appearance.
	var order []string
	num := map[string]int{}
	inFence = false
	for i, line := range body {
		if fenceRE.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		body[i] = fnRefRE.ReplaceAllStringFunc(line, func(ref string) string {
			label := fnRefRE.FindStringSubmatch(ref)[1]
			if _, ok := defs[label]; !ok {
				return ref
			}
			if _, ok := num[label]; !ok {
				order = append(order, label)
				num[label] = len(order)
			}
			return fmt.Sprintf("[%d]", num[label])
		})
	}
	if len(order) == 0 {
		return md
	}

	var b strings.Builder
	b.WriteString(strings.TrimRight(strings.Join(body, "\n"), "\n"))
	b.WriteString("\n\n---\n\n**Footnotes**\n\n")
	for _, label := range order {
		fmt.Fprintf(&b, "%d. %s\n", num[label], defs[label])
	}
	return b.String()
}
