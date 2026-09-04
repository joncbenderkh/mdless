// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"embed"
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

	// Background and Foreground, when set, paint the whole page — so a theme
	// works regardless of the terminal's own colours. Leave empty to sit on
	// the terminal's background.
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
	// ChromeBG tints the header/footer bars; defaults to a shade near
	// Background, else glamour's grey.
	ChromeBG string `json:"chrome_bg,omitempty"`
}

// Built-in themes are JSON documents in themes/, embedded at build time — the
// same format a user's own theme file uses, so each doubles as an example.
//
//go:embed themes/*.json
var themeFS embed.FS

var builtinThemes = loadBuiltinThemes()

func loadBuiltinThemes() map[string]Theme {
	entries, err := themeFS.ReadDir("themes")
	if err != nil {
		panic("mdless: embedded themes missing: " + err.Error())
	}
	m := make(map[string]Theme, len(entries))
	for _, e := range entries {
		data, err := themeFS.ReadFile("themes/" + e.Name())
		if err != nil {
			panic("mdless: reading embedded theme " + e.Name() + ": " + err.Error())
		}
		var th Theme
		if err := json.Unmarshal(data, &th); err != nil {
			panic("mdless: embedded theme " + e.Name() + ": " + err.Error())
		}
		if th.Name == "" {
			th.Name = strings.TrimSuffix(e.Name(), ".json")
		}
		m[th.Name] = th
	}
	if _, ok := m["default"]; !ok {
		panic("mdless: no built-in theme named \"default\"")
	}
	return m
}

// baseTheme is the theme scalar fields fall back to when a theme file omits
// them, and the answer for resolveTheme("").
func baseTheme() Theme { return builtinThemes["default"] }

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
		return baseTheme(), nil
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
	th := baseTheme()
	if err := json.Unmarshal(data, &th); err != nil {
		return Theme{}, fmt.Errorf("theme %s: %w", path, err)
	}
	if th.Name == "" || th.Name == "default" {
		th.Name = strings.TrimSuffix(filepath.Base(path), ".json")
	}
	return th, nil
}
