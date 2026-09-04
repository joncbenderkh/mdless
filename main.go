// SPDX-License-Identifier: GPL-3.0-or-later
//
// mdless — a less-style pager that renders Markdown in the terminal.
// Copyright (C) 2026 Jon Bender
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/joncbenderkh/mdless/internal/pager"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.1.0-dev"

const usage = `mdless — render Markdown in a less-style terminal pager

Usage:
  mdless [FILE]
  mdless < FILE
  some-command | mdless

Options:
  -h, --help           show this help and exit
  -v, --version        print version and exit
      --width N        wrap text at N columns (default 72; 0 = terminal width)
      --theme NAME     theme: a built-in name or a path to a .json theme
      --list-themes    list the built-in themes and exit
      --dump-theme N   print theme N as JSON (a starting point for your own)

Keys:
  j / k, ↓ / ↑     scroll one line          space / b   page down / up
  d / u            half page down / up      g / G       top / bottom
  /                search                   n / N       next / previous match
  m                toggle mouse capture     q, ctrl+c   quit

The mouse is not captured by default, so terminal click-drag selection
(copy/paste) keeps working; the wheel still scrolls in most terminals.
Press m to hand the mouse to the pager instead.

The theme is taken from --theme, else $MDLESS_THEME, else the terminal
background. Write a custom one with:  mdless --dump-theme default > my.json
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "mdless:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var path, theme string
	width := pager.DefaultMaxWidth
	if v := os.Getenv("MDLESS_WIDTH"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("MDLESS_WIDTH: %w", err)
		}
		width = n
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s needs a value", a)
			}
			i++
			return args[i], nil
		}
		switch {
		case a == "-h" || a == "--help":
			fmt.Print(usage)
			return nil
		case a == "-v" || a == "--version":
			fmt.Println("mdless", version)
			return nil
		case a == "--list-themes":
			for _, n := range pager.ThemeNames() {
				fmt.Println(n)
			}
			return nil
		case a == "--dump-theme":
			name, err := next()
			if err != nil {
				return err
			}
			return pager.DumpTheme(os.Stdout, name)
		case a == "--theme":
			v, err := next()
			if err != nil {
				return err
			}
			theme = v
		case strings.HasPrefix(a, "--theme="):
			theme = strings.TrimPrefix(a, "--theme=")
		case a == "--width", strings.HasPrefix(a, "--width="):
			v := strings.TrimPrefix(a, "--width=")
			if a == "--width" {
				var err error
				if v, err = next(); err != nil {
					return err
				}
			}
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return fmt.Errorf("--width needs a non-negative integer, got %q", v)
			}
			width = n
		case len(a) > 0 && a[0] == '-':
			return fmt.Errorf("unknown option %q (try --help)", a)
		default:
			path = a
		}
	}

	src, title, err := readInput(path)
	if err != nil {
		return err
	}
	return pager.Run(string(src), title, pager.Options{Theme: theme, MaxWidth: width})
}

func readInput(path string) (data []byte, title string, err error) {
	if path != "" {
		data, err = os.ReadFile(path)
		return data, path, err
	}
	if isTerminal(os.Stdin) {
		fmt.Print(usage)
		os.Exit(0)
	}
	data, err = io.ReadAll(os.Stdin)
	return data, "(stdin)", err
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
