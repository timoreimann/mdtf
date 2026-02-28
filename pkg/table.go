package pkg

import (
	"strings"
	"unicode/utf8"
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
	lines = append(lines, renderRow(t.Header, widths, pad))
	lines = append(lines, renderSeparator(t.Alignment, widths, colCount, pad))
	for _, row := range t.Rows {
		lines = append(lines, renderRow(row, widths, pad))
	}

	return lines
}

// columnWidths computes the maximum display width per column across all rows.
func (t Table) columnWidths(colCount int) []int {
	widths := make([]int, colCount)

	for i, cell := range t.Header {
		if w := utf8.RuneCountInString(cell); w > widths[i] {
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
			if w := utf8.RuneCountInString(cell); w > widths[i] {
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

// renderRow formats a single data row.
func renderRow(cells []string, widths []int, pad bool) string {
	var b strings.Builder
	b.WriteByte('|')

	for i, cell := range cells {
		b.WriteByte(' ')
		if pad && i < len(widths) {
			padded := padCell(cell, widths[i])
			b.WriteString(padded)
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

// padCell pads a cell value with trailing spaces to reach the target width.
func padCell(cell string, width int) string {
	cellWidth := utf8.RuneCountInString(cell)
	if cellWidth >= width {
		return cell
	}
	return cell + strings.Repeat(" ", width-cellWidth)
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
