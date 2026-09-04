# mdless

A `less`-style pager that renders Markdown in the terminal. Point it at a file or
pipe Markdown into it and scroll, page, and search through the rendered output the
way you would in `less`.

```
mdless README.md
git show HEAD:CHANGELOG.md | mdless
export MANPAGER=  # unrelated, but you get the idea:
export PAGER=mdless
```

## Keys

| Key | Action | Key | Action |
| --- | --- | --- | --- |
| `j` / `k`, `↓` / `↑` | scroll a line | `space` / `b` | page down / up |
| `d` / `u` | half page down / up | `g` / `G` | top / bottom |
| `/` | search | `n` / `N` | next / previous match |
| `m` | toggle mouse capture | `q`, `ctrl+c` | quit |

The mouse is **not** captured by default, so your terminal's own
click-drag selection (copy/paste) keeps working, and the wheel still
scrolls in terminals that map it to arrow keys under the alternate
screen. Press `m` to give the mouse to the pager instead.

## Width

Text wraps at **72 columns** by default — a long measure is tiring to
read. Override with `--width N` (or `$MDLESS_WIDTH`); `--width 0` uses the
full terminal width.

## Themes

`mdless` picks a theme from `--theme`, else `$MDLESS_THEME`, else the
terminal background. Built-ins:

| Name | |
| --- | --- |
| `default` | the standard dark look, on the terminal's own background |
| `light` | paints its own light page — readable on any terminal; the auto-pick on a light background |
| `mono` | greyscale, for a dark background |

```
mdless --list-themes
mdless --theme mono README.md
```

A theme is a small JSON document controlling the heading colours and
gutter glyphs, the horizontal-rule and H1-underline characters, the
code-panel colours, and optionally a `background` / `foreground` that
paint the whole page (so the theme doesn't depend on the terminal's
colours). Start from a built-in and edit:

```
mdless --dump-theme default > ~/.config/mdless/mine.json
mdless --theme ~/.config/mdless/mine.json README.md
```

Scalar fields left out of a theme file inherit the default's value; if a
file includes the `headings` array it must list all six levels.

## Build

```
go build -o mdless .
go test ./...
```

[`testdata/showcase.md`](./testdata/showcase.md) exercises every Markdown
feature the renderer supports; run `go run . testdata/showcase.md` to eyeball
style changes.

Requires Go 1.24+. Rendering is provided by
[`glamour`](https://github.com/charmbracelet/glamour); the interactive viewport is
built on [`bubbletea`](https://github.com/charmbracelet/bubbletea).

## License

GPL-3.0-or-later. See [`LICENSE`](./LICENSE).
