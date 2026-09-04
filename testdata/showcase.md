# mdless showcase

A single document that exercises every Markdown feature `mdless` renders, so
style changes can be eyeballed in one pass:

```sh
go run . testdata/showcase.md
go run . --theme mono testdata/showcase.md   # or --list-themes
```

Rendering goes through [glamour][g] with goldmark's **GFM**, **definition-list**
and **emoji** extensions; footnotes are handled by `mdless` itself. See the
sections near the end.

[g]: https://github.com/charmbracelet/glamour

---

## Headings

The title above is H1: a banner with a heavy full-width rule beneath it. The
lines below are H2-H6 — a gutter bar that narrows with depth, and text that
steps down from a bold underlined accent (H2) to plain dim grey (H6). A theme
can paint its own page background, so `--theme light` is readable on any
terminal.

### Heading level 3

#### Heading level 4

##### Heading level 5

###### Heading level 6

## Inline text

Regular paragraph text with **bold**, *italic*, ***bold italic***,
~~strikethrough~~, `inline code`, and a hard line break at the end of this
line —  
the following sentence begins on a new line within the same paragraph.

Backslash line break at the end of this line —\
and its continuation.

Escaped characters should print literally: \*not italic\*, \# not a heading,
\`not code\`, 1\. not a list.

Entities and Unicode: &copy; 2026 · café · 世界 · ½ · → · ✓ · 😀

Nested emphasis: **bold with *italic* and `code` inside**, and
*italic containing **bold** and a [link](https://example.com)*.

## Links and images

- Inline link: [Charmbracelet](https://charm.sh)
- Link with title: [hover me](https://example.com "Example Domain")
- Reference link: [glamour repository][g]
- Bare autolink (GFM linkify): https://www.rfc-editor.org/rfc/rfc7763
- Angle autolink: <https://spec.commonmark.org>
- Email autolink: <mailto:nobody@example.com>
- Inline image: ![tiny placeholder](https://example.com/image.png "1x1")

## Blockquotes

> Single-level blockquote with **formatting**, `code`, and a [link](https://example.com).
>
> > Nested blockquote, second level.
> >
> > > Third level, with a list inside:
> > >
> > > - alpha
> > > - beta

## Lists

### Unordered, nested

- First item
- Second item
  - Nested one
  - Nested two
    - Deeper still
- Third item, one long source line — long enough that the renderer must reflow it across the available width so the wrapped continuation lines up under the text, not the bullet

### Ordered, with a custom start

1. Step one
2. Step two
   1. Sub-step a
   2. Sub-step b
3. Step three

<!-- start attribute -->
7. Restarted list beginning at seven
8. Eight
9. Nine

### Loose list (blank lines between items)

- Loose items render each paragraph with vertical spacing.

- Second loose item, still one paragraph.

  A second paragraph belonging to the same item.

### Task list (GFM)

- [x] Completed task
- [ ] Pending task
- [ ] Pending task with **bold** and `code`

## Code

Indented code block (four spaces):

    function indented() {
      return "no language, no highlighting";
    }

Fenced block without a language:

```
plain fenced block
  preserves   whitespace
```

Fenced block with a language (syntax highlighting):

```go
package main

import "fmt"

// Fibonacci prints the first n Fibonacci numbers.
func Fibonacci(n int) {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		fmt.Printf("%d ", a)
		a, b = b, a+b
	}
	fmt.Println()
}
```

```python
def quicksort(xs):
    if len(xs) <= 1:
        return xs
    pivot = xs[len(xs) // 2]
    left = [x for x in xs if x < pivot]
    mid = [x for x in xs if x == pivot]
    right = [x for x in xs if x > pivot]
    return quicksort(left) + mid + quicksort(right)
```

```json
{ "name": "mdless", "nested": { "array": [1, 2, 3], "flag": true } }
```

## Tables (GFM)

| Feature        | Supported | Notes                                  |
| -------------- | :-------: | -------------------------------------- |
| Headings       |    yes    | H1 banner, H2-H6 gutter                |
| Tables         |    yes    | alignment row controls each column     |
| Footnotes      |    no     | goldmark extension not enabled         |
| Very long cell |    yes    | this cell is deliberately wide so wrapping or truncation behaviour is visible |

Alignment check:

| Left | Center | Right |
| :--- | :----: | ----: |
| a    |   b    |     c |
| 100  |  1000  | 10000 |

## Definition list (glamour extension)

Term one
: Definition of the first term.

Term two
: First definition of the second term.
: Second definition of the second term.

## Horizontal rules

Three of them, each in a different syntax:

---

***

___

## Inline HTML

Raw HTML such as <b>bold tags</b>, <kbd>Ctrl</kbd>+<kbd>C</kbd>, and
<span style="color:red">a styled span</span> is sanitized by glamour and
generally rendered as plain text.

<div>
  Block-level HTML is passed through as-is.
</div>

## Footnotes and emoji

- Emoji shortcodes render: :tada: :rocket: :warning:
- A footnote reference[^note] is renumbered, and every definition is collected into a **Footnotes** section at the end of the document.[^2]

[^note]: The first footnote. Definitions may wrap onto
  indented continuation lines like this one.
[^2]: A second footnote, to show the numbering.

## Wrapping stress test

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. A word-that-is-unusually-long-and-hyphenated-to-probe-the-reflow-logic follows, then normal prose resumes to confirm the wrap recovers cleanly.

---

End of showcase.
