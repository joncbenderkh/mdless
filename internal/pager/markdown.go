// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"bytes"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// markdownRenderer is glamour's pipeline rebuilt directly so the emoji extension
// can be added — glamour omits it from NewTermRenderer. (Footnotes are handled
// before parsing, by expandFootnotes.)
type markdownRenderer struct{ md goldmark.Markdown }

// glamour registers its ANSI renderer at this priority (see glamour.go).
const ansiRendererPriority = 1000

func newRenderer(env renderEnv, wrap int) (*markdownRenderer, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			emoji.Emoji,
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	ar := ansi.NewRenderer(ansi.Options{
		WordWrap:     wrap,
		ColorProfile: env.profile,
		Styles:       readableStyle(baseStyleConfig(env.style), wrap, env.theme),
	})
	md.SetRenderer(renderer.NewRenderer(
		renderer.WithNodeRenderers(util.Prioritized(ar, ansiRendererPriority)),
	))
	return &markdownRenderer{md: md}, nil
}

func (r *markdownRenderer) Render(in string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(in), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
