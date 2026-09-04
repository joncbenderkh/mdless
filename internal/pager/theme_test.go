// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func TestResolveTheme(t *testing.T) {
	if th, _ := resolveTheme(""); th.Name != "default" {
		t.Fatalf(`resolveTheme("") = %q, want "default"`, th.Name)
	}
	if th, _ := resolveTheme("mono"); th.Name != "mono" {
		t.Fatalf(`resolveTheme("mono") = %q`, th.Name)
	}
	if _, err := resolveTheme("nope"); err == nil {
		t.Fatal("resolveTheme(nope) should error")
	}
}

func TestThemeFileRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := DumpTheme(&buf, "default"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "t.json")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	th, err := resolveTheme(path)
	if err != nil {
		t.Fatal(err)
	}
	if th.Headings != baseTheme().Headings || th.CodeBG != baseTheme().CodeBG {
		t.Fatalf("round-tripped theme differs from default:\n%+v", th)
	}
	if th.Name != "t" { // name derived from the file
		t.Fatalf("theme name = %q, want \"t\"", th.Name)
	}
}

func TestThemeFilePartialScalars(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.json")
	os.WriteFile(path, []byte(`{"code_bg":"52"}`), 0o644)
	th, err := resolveTheme(path)
	if err != nil {
		t.Fatal(err)
	}
	if th.CodeBG != "52" {
		t.Fatalf("CodeBG = %q, want 52", th.CodeBG)
	}
	if th.Headings != baseTheme().Headings { // absent "headings" keeps the default's
		t.Fatal("partial theme lost the default headings")
	}
}

// A theme's heading colours and gutters must reach the rendered output.
func TestThemeAppliedToHeadings(t *testing.T) {
	th := baseTheme()
	th.Headings[2].Gutter = "» "
	th.Headings[2].FG = "201"
	env := renderEnv{style: "dark", profile: termenv.ANSI256, theme: th}

	out, err := render("### Custom\n\nbody\n", env, 60)
	if err != nil {
		t.Fatal(err)
	}
	var h3 string
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(stripANSI(ln), "Custom") {
			h3 = ln
		}
	}
	if !strings.Contains(stripANSI(h3), "» Custom") {
		t.Fatalf("custom gutter not applied: %q", stripANSI(h3))
	}
	if !strings.Contains(h3, "38;5;201") {
		t.Fatalf("custom colour not applied: %q", h3)
	}
}

func TestDumpThemeIsValidJSON(t *testing.T) {
	for _, name := range ThemeNames() {
		var buf bytes.Buffer
		if err := DumpTheme(&buf, name); err != nil {
			t.Fatalf("DumpTheme(%s): %v", name, err)
		}
		var th Theme
		if err := json.Unmarshal(buf.Bytes(), &th); err != nil {
			t.Fatalf("DumpTheme(%s) not valid JSON: %v", name, err)
		}
	}
}
