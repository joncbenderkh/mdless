// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import "testing"

func TestStripANSI(t *testing.T) {
	in := "\x1b[1mbold\x1b[0m plain \x1b[38;5;203mred\x1b[0m"
	want := "bold plain red"
	if got := stripANSI(in); got != want {
		t.Fatalf("stripANSI() = %q, want %q", got, want)
	}
}

func TestRecomputeMatches(t *testing.T) {
	m := &model{
		plainLines: []string{"first line", "Second LINE", "third", "another line"},
		query:      "line",
	}
	m.recomputeMatches()
	want := []int{0, 1, 3}
	if len(m.matches) != len(want) {
		t.Fatalf("matches = %v, want %v", m.matches, want)
	}
	for i, v := range want {
		if m.matches[i] != v {
			t.Fatalf("matches = %v, want %v", m.matches, want)
		}
	}
}

func TestNearestMatchWraps(t *testing.T) {
	m := model{matches: []int{2, 8, 20}}
	m.viewport.YOffset = 25
	if got := m.nearestMatch(); got != 0 {
		t.Fatalf("nearestMatch() = %d, want 0 (wrap to top)", got)
	}
}
