// SPDX-License-Identifier: GPL-3.0-or-later
//
// mdless — a less-style pager that renders Markdown in the terminal.
// Copyright (C) 2026 Jon Bender
package main

import (
	"fmt"
	"io"
	"os"

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
  -h, --help       show this help and exit
  -v, --version    print version and exit

Keys:
  j / k, ↓ / ↑     scroll one line          space / b   page down / up
  d / u            half page down / up      g / G       top / bottom
  /                search                   n / N       next / previous match
  m                toggle mouse capture     q, ctrl+c   quit

The mouse is not captured by default, so terminal click-drag selection
(copy/paste) keeps working; the wheel still scrolls in most terminals.
Press m to hand the mouse to the pager instead.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "mdless:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var path string
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(usage)
			return nil
		case "-v", "--version":
			fmt.Println("mdless", version)
			return nil
		default:
			if len(a) > 0 && a[0] == '-' {
				return fmt.Errorf("unknown option %q (try --help)", a)
			}
			path = a
		}
	}

	src, title, err := readInput(path)
	if err != nil {
		return err
	}
	return pager.Run(string(src), title)
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
