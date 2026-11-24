package pkg

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

// lines joins its arguments with newlines, so that test inputs read like the
// Markdown they stand for.
func lines(l ...string) string {
	return strings.Join(l, "\n")
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		in   string
		opts Options
		want string
	}{
		{
			name: "pads columns",
			in: lines(
				"| Name | Age | City |",
				"|---|---|---|",
				"| Alice | 30 | San Francisco |",
				"| Bob | 25 | New York |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| Name  | Age | City          |",
				"| ----- | --- | ------------- |",
				"| Alice | 30  | San Francisco |",
				"| Bob   | 25  | New York      |",
			),
		},
		{
			name: "compact output without padding",
			in: lines(
				"|  Name|Age  |",
				"|:-----|----:|",
				"|Alice   |30|",
			),
			opts: Options{Pad: false},
			want: lines(
				"| Name | Age |",
				"| :-- | --: |",
				"| Alice | 30 |",
			),
		},
		{
			name: "aligns cell content to the column alignment",
			in: lines(
				"| left | center | right | default |",
				"|:-|:-:|-:|-|",
				"| a | b | c | d |",
				"| aaaaaa | bbbbbbb | cccccc | dddddddd |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| left   | center  |  right | default  |",
				"| :----- | :-----: | -----: | -------- |",
				"| a      |    b    |      c | d        |",
				"| aaaaaa | bbbbbbb | cccccc | dddddddd |",
			),
		},
		{
			name: "puts the extra space of a centered cell on the right",
			in: lines(
				"| ab |",
				"|:----:|",
				"| x |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| ab  |",
				"| :-: |",
				"|  x  |",
			),
		},
		{
			name: "keeps escaped pipes in cells",
			in: lines(
				`| expr | meaning |`,
				`|---|---|`,
				`| a \| b | a or b |`,
			),
			opts: Options{Pad: true},
			want: lines(
				`| expr   | meaning |`,
				`| ------ | ------- |`,
				`| a \| b | a or b  |`,
			),
		},
		{
			name: "fills short rows and cuts long rows to the header width",
			in: lines(
				"| a | b |",
				"|---|---|",
				"| 1 |",
				"| 1 | 2 | 3 |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| a   | b   |",
				"| --- | --- |",
				"| 1   |     |",
				"| 1   | 2   |",
			),
		},
		{
			name: "measures the display width of wide and combining characters",
			in: lines(
				"| text | note |",
				"|---|---|",
				"| 日本語 | CJK |",
				"| 🦀 | emoji |",
				"| 👩‍💻 | ZWJ sequence |",
				"| é | combining mark |",
				"| é | precomposed |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| text   | note           |",
				"| ------ | -------------- |",
				"| 日本語 | CJK            |",
				"| 🦀     | emoji          |",
				"| 👩‍💻     | ZWJ sequence   |",
				"| é      | combining mark |",
				"| é      | precomposed    |",
			),
		},
		{
			name: "formats every table and leaves other lines alone",
			in: lines(
				"# Title",
				"",
				"|a|b|",
				"|-|-|",
				"|1|2|",
				"Some | text with | pipes",
				"",
				"|c|",
				"|-|",
			),
			opts: Options{Pad: true},
			want: lines(
				"# Title",
				"",
				"| a   | b   |",
				"| --- | --- |",
				"| 1   | 2   |",
				"Some | text with | pipes",
				"",
				"| c   |",
				"| --- |",
			),
		},
		{
			name: "needs a separator row to detect a table",
			in: lines(
				"| a | b |",
				"| not | separator |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| a | b |",
				"| not | separator |",
			),
		},
		{
			name: "needs a dash in the separator row",
			in: lines(
				"| a |",
				"| :: |",
			),
			opts: Options{Pad: true},
			want: lines(
				"| a |",
				"| :: |",
			),
		},
		{
			name: "keeps the indentation of a table in a list item",
			in: lines(
				"- item",
				"",
				"  |a|bb|",
				"  |-|-|",
				"  |ccc|d|",
			),
			opts: Options{Pad: true},
			want: lines(
				"- item",
				"",
				"  | a   | bb  |",
				"  | --- | --- |",
				"  | ccc | d   |",
			),
		},
		{
			name: "keeps a trailing newline",
			in:   "|a|\n|-|\n|b|\n",
			opts: Options{Pad: true},
			want: "| a   |\n| --- |\n| b   |\n",
		},
		{
			name: "leaves content without tables unchanged",
			in:   "plain text\nwith | a pipe\n",
			opts: Options{Pad: true},
			want: "plain text\nwith | a pipe\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.in, tt.opts)
			assert.NilError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestFormatCodeBlocks(t *testing.T) {
	// unformatted is a table that Format changes when it is outside of a code
	// block. formatted is the result.
	unformatted := lines("|q|r|", "|-|-|")
	formatted := lines("| q   | r   |", "| --- | --- |")

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "backtick fence",
			in:   lines("```", unformatted, "```", unformatted),
			want: lines("```", unformatted, "```", formatted),
		},
		{
			name: "tilde fence with info string",
			in:   lines("~~~markdown", unformatted, "~~~", unformatted),
			want: lines("~~~markdown", unformatted, "~~~", formatted),
		},
		{
			name: "longer fence needs a closing fence of the same length",
			in:   lines("````", "```", unformatted, "````", unformatted),
			want: lines("````", "```", unformatted, "````", formatted),
		},
		{
			name: "fence of the other character does not close the block",
			in:   lines("```", "~~~", unformatted, "```", unformatted),
			want: lines("```", "~~~", unformatted, "```", formatted),
		},
		{
			name: "fence with text after it does not close the block",
			in:   lines("```", "``` text", unformatted, "```", unformatted),
			want: lines("```", "``` text", unformatted, "```", formatted),
		},
		{
			name: "indented fence in a list item",
			in:   lines("- item", "", "  ```", "  "+unformatted, "  ```", unformatted),
			want: lines("- item", "", "  ```", "  "+unformatted, "  ```", formatted),
		},
		{
			name: "unclosed fence runs to the end of the document",
			in:   lines("```", unformatted, "", unformatted),
			want: lines("```", unformatted, "", unformatted),
		},
		{
			name: "backtick in the info string is no fence",
			in:   lines("``` a`b", unformatted),
			want: lines("``` a`b", formatted),
		},
		{
			name: "two backticks are no fence",
			in:   lines("``", unformatted),
			want: lines("``", formatted),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.in, Options{Pad: true})
			assert.NilError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestFormatRange(t *testing.T) {
	// Lines 1-2 hold the first table, lines 4-5 the second one.
	in := lines(
		"|a|",
		"|-|",
		"",
		"|b|",
		"|-|",
	)
	first := lines("| a   |", "| --- |")
	second := lines("| b   |", "| --- |")

	tests := []struct {
		name       string
		start, end int
		want       string
	}{
		{name: "first table only", start: 1, end: 1, want: lines(first, "", "|b|", "|-|")},
		{name: "second table only", start: 5, end: 5, want: lines("|a|", "|-|", "", second)},
		{name: "range that overlaps both tables", start: 2, end: 4, want: lines(first, "", second)},
		{name: "range between the tables", start: 3, end: 3, want: in},
		{name: "open end", start: 4, end: 0, want: lines("|a|", "|-|", "", second)},
		{name: "open start", start: 0, end: 2, want: lines(first, "", "|b|", "|-|")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(in, Options{Pad: true, Start: tt.start, End: tt.end})
			assert.NilError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestFormatInvalidRange(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{name: "negative start", opts: Options{Start: -1}, want: "line range values must be non-negative"},
		{name: "negative end", opts: Options{End: -1}, want: "line range values must be non-negative"},
		{name: "start after end", opts: Options{Start: 5, End: 2}, want: "start line (5) must not exceed end line (2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Format("|a|\n|-|\n", tt.opts)
			assert.Error(t, err, tt.want)
		})
	}
}
