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

## Build

```
go build -o mdless .
go test ./...
```

Requires Go 1.24+. Rendering is provided by
[`glamour`](https://github.com/charmbracelet/glamour); the interactive viewport is
built on [`bubbletea`](https://github.com/charmbracelet/bubbletea).

## License

GPL-3.0-or-later. See [`LICENSE`](./LICENSE).
