// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func cmdID(c tea.Cmd) uintptr { return reflect.ValueOf(c).Pointer() }

func TestMouseToggle(t *testing.T) {
	m := model{}
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")}

	cmd, handled := m.handleKey(key)
	if !handled || !m.mouseCapture {
		t.Fatalf("first m: handled=%v mouseCapture=%v", handled, m.mouseCapture)
	}
	if cmdID(cmd) != cmdID(tea.Cmd(tea.EnableMouseCellMotion)) {
		t.Fatal("first m should return the enable-mouse command")
	}

	cmd, _ = m.handleKey(key)
	if m.mouseCapture {
		t.Fatal("second m should disable mouse capture")
	}
	if cmdID(cmd) != cmdID(tea.Cmd(tea.DisableMouse)) {
		t.Fatal("second m should return the disable-mouse command")
	}
}

func TestFooterReflectsMouseState(t *testing.T) {
	m := model{}
	m.viewport.Width = 90

	if got := stripANSI(m.footerView()); !strings.Contains(got, "m:select") {
		t.Fatalf("default footer should show m:select, got %q", got)
	}
	m.mouseCapture = true
	if got := stripANSI(m.footerView()); !strings.Contains(got, "m:scroll") {
		t.Fatalf("captured footer should show m:scroll, got %q", got)
	}
}
