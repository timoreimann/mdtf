package pkg

import (
	"strings"

	"github.com/rivo/uniseg"
)

// Alignment represents the alignment of a table column.
type Alignment int

const (
	AlignDefault Alignment = iota
	AlignLeft
	AlignCenter
	AlignRight
)

// Table represents a parsed markdown table with its position in the source file.
type Table struct {
	// Indent is the leading whitespace of the header row. Every rendered row
	// gets it, so that tables in list items stay in their list item.
	Indent    string
	Header    []string
	Alignment []Alignment
	Rows      [][]string
	StartLine int // 1-based inclusive
	EndLine   int // 1-based inclusive
}

// ParseTable parses raw lines (header + separator + data rows) into a Table.
// startLine is the 1-based line number of the first line in the source file.
func ParseTable(lines []string, startLine int) Table {
	t := Table{
		StartLine: startLine,
		EndLine:   startLine + len(lines) - 1,
	}

	if len(lines) < 2 {
		return t
	}

	t.Indent = leadingWhitespace(lines[0])
	t.Header = parseCells(lines[0])
	t.Alignment = parseAlignments(lines[1])

	// Normalize column count to header width.
	colCount := len(t.Header)
	t.Alignment = normalizeSlice(t.Alignment, colCount)

	for _, line := range lines[2:] {
		row := parseCells(line)
		row = normalizeSlice(row, colCount)
		t.Rows = append(t.Rows, row)
	}

	return t
}

// Render formats the table as markdown lines.
// When pad is true, columns are space-padded to equal width.
func (t Table) Render(pad bool) []string {
	if len(t.Header) == 0 {
		return nil
	}

	colCount := len(t.Header)

	var widths []int
	if pad {
		widths = t.columnWidths(colCount)
	}

	var lines []string
	lines = append(lines, t.Indent+renderRow(t.Header, t.Alignment, widths, pad))
	lines = append(lines, t.Indent+renderSeparator(t.Alignment, widths, colCount, pad))
	for _, row := range t.Rows {
		lines = append(lines, t.Indent+renderRow(row, t.Alignment, widths, pad))
	}

	return lines
}

// columnWidths computes the maximum display width per column across all rows.
// The display width counts the terminal columns of the text, so that wide
// characters (CJK, most emoji) take two columns and combining marks none.
func (t Table) columnWidths(colCount int) []int {
	widths := make([]int, colCount)

	for i, cell := range t.Header {
		if w := uniseg.StringWidth(cell); w > widths[i] {
			widths[i] = w
		}
	}

	// Separator needs at least 3 chars (e.g. "---" or ":--").
	for i, a := range t.Alignment {
		minW := separatorMinWidth(a)
		if minW > widths[i] {
			widths[i] = minW
		}
	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if w := uniseg.StringWidth(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	return widths
}

// parseCells splits a markdown table row into trimmed cell values.
func parseCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	parts := splitUnescapedPipes(line)
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// splitUnescapedPipes splits on "|" but respects escaped "\|".
func splitUnescapedPipes(s string) []string {
	var parts []string
	var current strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && s[i+1] == '|' {
			current.WriteString(`\|`)
			i++ // skip the escaped pipe
			continue
		}
		if s[i] == '|' {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(s[i])
	}
	parts = append(parts, current.String())

	return parts
}

// parseAlignments extracts column alignments from a separator row.
func parseAlignments(line string) []Alignment {
	cells := parseCells(line)
	alignments := make([]Alignment, len(cells))

	for i, cell := range cells {
		cell = strings.TrimSpace(cell)
		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")

		switch {
		case left && right:
			alignments[i] = AlignCenter
		case right:
			alignments[i] = AlignRight
		case left:
			alignments[i] = AlignLeft
		default:
			alignments[i] = AlignDefault
		}
	}

	return alignments
}

// renderRow formats a single header or data row. When pad is true, each cell
// is padded to its column width according to the column alignment.
func renderRow(cells []string, alignments []Alignment, widths []int, pad bool) string {
	var b strings.Builder
	b.WriteByte('|')

	for i, cell := range cells {
		b.WriteByte(' ')
		if pad && i < len(widths) {
			b.WriteString(padCell(cell, alignments[i], widths[i]))
		} else {
			b.WriteString(cell)
		}
		b.WriteString(" |")
	}

	return b.String()
}

// renderSeparator formats the alignment/separator row.
func renderSeparator(alignments []Alignment, widths []int, colCount int, pad bool) string {
	var b strings.Builder
	b.WriteByte('|')

	for i := 0; i < colCount; i++ {
		b.WriteByte(' ')

		a := AlignDefault
		if i < len(alignments) {
			a = alignments[i]
		}

		w := 3 // minimum separator width
		if pad && i < len(widths) {
			w = widths[i]
		}

		b.WriteString(buildSeparatorCell(a, w))
		b.WriteString(" |")
	}

	return b.String()
}

// buildSeparatorCell creates a single separator cell like "---", ":--", ":-:", or "--:".
func buildSeparatorCell(a Alignment, width int) string {
	if width < 3 {
		width = 3
	}

	switch a {
	case AlignLeft:
		return ":" + strings.Repeat("-", width-1)
	case AlignCenter:
		return ":" + strings.Repeat("-", width-2) + ":"
	case AlignRight:
		return strings.Repeat("-", width-1) + ":"
	default:
		return strings.Repeat("-", width)
	}
}

// padCell pads a cell value with spaces to reach the target display width.
// Right-aligned cells get the spaces before the value, centered cells on both
// sides (the extra space of an odd remainder goes to the right), and all other
// cells after the value.
func padCell(cell string, a Alignment, width int) string {
	missing := width - uniseg.StringWidth(cell)
	if missing <= 0 {
		return cell
	}

	switch a {
	case AlignRight:
		return strings.Repeat(" ", missing) + cell
	case AlignCenter:
		left := missing / 2
		return strings.Repeat(" ", left) + cell + strings.Repeat(" ", missing-left)
	default:
		return cell + strings.Repeat(" ", missing)
	}
}

// leadingWhitespace returns the spaces and tabs at the start of line.
func leadingWhitespace(line string) string {
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}

// separatorMinWidth returns the minimum width for a separator cell given its alignment.
func separatorMinWidth(a Alignment) int {
	switch a {
	case AlignCenter:
		return 3 // ":"  + "-" + ":"
	case AlignLeft, AlignRight:
		return 3 // ":" + "--" or "--" + ":"
	default:
		return 3 // "---"
	}
}

// normalizeSlice pads or truncates a string slice to exactly n elements.
func normalizeSlice[T any](s []T, n int) []T {
	if len(s) >= n {
		return s[:n]
	}
	result := make([]T, n)
	copy(result, s)
	return result
}
