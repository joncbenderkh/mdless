// SPDX-License-Identifier: GPL-3.0-or-later

package pager

import (
	"strings"
	"testing"
)

func TestExpandFootnotes(t *testing.T) {
	in := "See here[^a] and also here[^b], plus a[^a] repeat.\n\n" +
		"[^a]: First note.\n  Continued on an indented line.\n" +
		"[^b]: Second note.\n" +
		"[^unused]: Never referenced.\n"
	out := expandFootnotes(in)

	if strings.Contains(out, "[^a]") || strings.Contains(out, "[^b]") {
		t.Fatalf("footnote markers left in body:\n%s", out)
	}
	if !strings.Contains(out, "here[1] and also here[2]") {
		t.Fatalf("references not numbered by first appearance:\n%s", out)
	}
	if !strings.Contains(out, "a[1] repeat") {
		t.Fatalf("repeat reference not reused:\n%s", out)
	}
	if !strings.Contains(out, "**Footnotes**") {
		t.Fatalf("no Footnotes section:\n%s", out)
	}
	if !strings.Contains(out, "1. First note. Continued on an indented line.") {
		t.Fatalf("continuation line not joined:\n%s", out)
	}
	if strings.Contains(out, "Never referenced") {
		t.Fatalf("unreferenced definition should be dropped:\n%s", out)
	}
}

func TestExpandFootnotesNoOp(t *testing.T) {
	in := "Plain text with a [link](x) and `[^notarealref]` in code.\n"
	if got := expandFootnotes(in); got != in {
		t.Fatalf("expandFootnotes changed a doc with no footnotes:\n%q", got)
	}
}

func TestExpandFootnotesIgnoresFences(t *testing.T) {
	in := "Real[^r].\n\n```\ncode[^r] stays literal\n```\n\n[^r]: def\n"
	out := expandFootnotes(in)
	if !strings.Contains(out, "code[^r] stays literal") {
		t.Fatalf("footnote ref inside a fence was rewritten:\n%s", out)
	}
	if !strings.Contains(out, "Real[1].") {
		t.Fatalf("real reference not rewritten:\n%s", out)
	}
}
