# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

`mdless` — a `less`-style terminal pager that renders Markdown. Reads a file
argument or stdin, renders with `glamour`, and drives an interactive `bubbletea`
viewport with scroll/page/search keybindings.

## Decisions

- **License:** GPL-3.0-or-later (`LICENSE` at repo root; `SPDX-License-Identifier`
  header on each Go source file).
- **Language / stack:** Go 1.24+. Charm stack — `bubbletea`, `bubbles/viewport`,
  `glamour`, `lipgloss`. No other runtime dependencies.
- **Module path:** `github.com/joncbenderkh/mdless`.

## Layout

```
main.go                 CLI entry: flag/arg parsing (--theme, --dump-theme …)
internal/pager/
  pager.go              Run(); renderEnv + theme resolution before bubbletea
  model.go              bubbletea model: layout, render, key dispatch, views;
                        fillPanels post-processing (code + H1 background)
  style.go              readableStyle: glamour style + theme -> ansi.StyleConfig
  theme.go              Theme struct; loads themes/*.json (embedded), DumpTheme
  themes/*.json         the built-in themes — same format as a user theme file
  search.go             `/` search: ANSI stripping, match indexing, scroll-to
  *_test.go             unit tests for the pure helpers
testdata/showcase.md    every supported Markdown feature, for eyeballing
```

## Commands

```
go build -o mdless .    # build the binary
go test ./...           # run tests
go vet ./...             # vet
gofmt -l .              # list unformatted files (should be empty)
```

## Conventions

- Keep the rendering and search logic pure and testable; confine terminal I/O to
  `pager.go` / the `Update`/`View` methods.
- New keybindings go in `model.handleKey`; document them in `README.md` and the
  `usage` string in `main.go`.
- All colours and glyphs live in a `Theme`; built-ins are `themes/*.json`, not
  Go literals. `default` must stay pixel-identical to the current look. Don't
  hard-code a colour in `style.go` / `model.go` — add a `Theme` field and set
  it in every `themes/*.json`.
