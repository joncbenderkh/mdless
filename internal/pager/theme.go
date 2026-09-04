// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HeadingStyle is how one heading level (H1-H6) is drawn on top of glamour.
// An empty FG leaves glamour's own cascade in charge of the colour.
type HeadingStyle struct {
	Gutter    string `json:"gutter"`
	FG        string `json:"fg,omitempty"`
	BG        string `json:"bg,omitempty"`
	Bold      bool   `json:"bold,omitempty"`
	Underline bool   `json:"underline,omitempty"`
	Upper     bool   `json:"upper,omitempty"`
	Italic    bool   `json:"italic,omitempty"`
	Faint     bool   `json:"faint,omitempty"`
	// Banner draws the heading as a full-width filled panel with a rule under
	// it (only meaningful for H1).
	Banner bool `json:"banner,omitempty"`
}

// Theme is the complete set of colours and glyphs mdless layers over glamour.
// Colours are ANSI-256 indices ("81") or hex ("#2b2b2b").
type Theme struct {
	Name string `json:"name"`
	// GlamourStyle is the built-in glamour style to build on: "dark", "light",
	// "dracula", "tokyo-night", "pink", or "" to pick dark/light by the
	// terminal background.
	GlamourStyle string          `json:"glamour_style,omitempty"`
	Headings     [6]HeadingStyle `json:"headings"`
	H1RuleChar   string          `json:"h1_rule_char"`
	H1RuleColor  string          `json:"h1_rule_color,omitempty"`
	RuleChar     string          `json:"rule_char"`
	RuleColor    string          `json:"rule_color,omitempty"`
	CodeBG       string          `json:"code_bg"`
	CodeFG       string          `json:"code_fg"`
}

var defaultTheme = Theme{
	Name:         "default",
	GlamourStyle: "dark",
	Headings: [6]HeadingStyle{
		{Gutter: "  ", FG: "231", BG: "63", Bold: true, Upper: true, Banner: true},
		{Gutter: "▌ ", FG: "81", Bold: true, Upper: true, Underline: true},
		{Gutter: "▍ ", FG: "39"},
		{Gutter: "▏ ", FG: "250"},
		{Gutter: "▏ ", FG: "244"},
		{Gutter: "▏ ", FG: "238"},
	},
	H1RuleChar:  "━",
	H1RuleColor: "63",
	RuleChar:    "─",
	RuleColor:   "240",
	CodeBG:      "236",
	CodeFG:      "252",
}

var lightTheme = Theme{
	Name:         "light",
	GlamourStyle: "light",
	Headings: [6]HeadingStyle{
		{Gutter: "  ", FG: "231", BG: "61", Bold: true, Upper: true, Banner: true},
		{Gutter: "▌ ", FG: "25", Bold: true, Upper: true, Underline: true},
		{Gutter: "▍ ", FG: "24"},
		{Gutter: "▏ ", FG: "240"},
		{Gutter: "▏ ", FG: "244"},
		{Gutter: "▏ ", FG: "248"},
	},
	H1RuleChar:  "━",
	H1RuleColor: "61",
	RuleChar:    "─",
	RuleColor:   "250",
	CodeBG:      "254",
	CodeFG:      "236",
}

var monoTheme = Theme{
	Name:         "mono",
	GlamourStyle: "dark",
	Headings: [6]HeadingStyle{
		{Gutter: "  ", FG: "255", BG: "238", Bold: true, Upper: true, Banner: true},
		{Gutter: "▌ ", FG: "255", Bold: true, Upper: true, Underline: true},
		{Gutter: "▍ ", FG: "252", Bold: true},
		{Gutter: "▏ ", FG: "250"},
		{Gutter: "▏ ", FG: "245"},
		{Gutter: "▏ ", FG: "240"},
	},
	H1RuleChar:  "━",
	H1RuleColor: "244",
	RuleChar:    "─",
	RuleColor:   "240",
	CodeBG:      "235",
	CodeFG:      "252",
}

var builtinThemes = map[string]Theme{
	defaultTheme.Name: defaultTheme,
	lightTheme.Name:   lightTheme,
	monoTheme.Name:    monoTheme,
}

// ThemeNames lists the built-in themes, sorted.
func ThemeNames() []string {
	names := make([]string, 0, len(builtinThemes))
	for n := range builtinThemes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// DumpTheme writes a theme as indented JSON — a starting point for a custom
// one. An empty name dumps the default.
func DumpTheme(w io.Writer, name string) error {
	th, err := resolveTheme(name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(th)
}

// resolveTheme turns a --theme value into a Theme: a built-in name, a path to a
// JSON file, or "" for the default.
func resolveTheme(spec string) (Theme, error) {
	if spec == "" {
		return defaultTheme, nil
	}
	if th, ok := builtinThemes[spec]; ok {
		return th, nil
	}
	if strings.ContainsAny(spec, "/\\") || strings.HasSuffix(spec, ".json") {
		return loadThemeFile(spec)
	}
	return Theme{}, fmt.Errorf(
		"unknown theme %q; built-in themes: %s (or pass a path to a .json file)",
		spec, strings.Join(ThemeNames(), ", "))
}

// loadThemeFile reads a JSON theme. Scalar fields absent from the file keep the
// default's value; if the file has a "headings" array it must list all six.
func loadThemeFile(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, err
	}
	th := defaultTheme
	if err := json.Unmarshal(data, &th); err != nil {
		return Theme{}, fmt.Errorf("theme %s: %w", path, err)
	}
	if th.Name == "" || th.Name == defaultTheme.Name {
		th.Name = strings.TrimSuffix(filepath.Base(path), ".json")
	}
	return th, nil
}
