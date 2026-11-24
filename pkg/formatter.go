package pkg

import (
	"fmt"
	"strings"
)

// Options controls formatting behavior.
type Options struct {
	Pad   bool // Pad columns to equal width.
	Start int  // 1-based start line (0 = no range filter).
	End   int  // 1-based end line (0 = no range filter).
}

// Format reads markdown content, reformats tables according to opts, and
// returns the modified content.
func Format(content string, opts Options) (string, error) {
	if err := validateOptions(opts); err != nil {
		return "", err
	}

	lines := strings.Split(content, "\n")
	tables := detectTables(lines)

	if !hasRange(opts) {
		return applyAll(lines, tables, opts.Pad), nil
	}

	return applyRange(lines, tables, opts), nil
}

// validateOptions checks that range boundaries are sane.
func validateOptions(opts Options) error {
	if opts.Start < 0 || opts.End < 0 {
		return fmt.Errorf("line range values must be non-negative")
	}
	if opts.Start > 0 && opts.End > 0 && opts.Start > opts.End {
		return fmt.Errorf("start line (%d) must not exceed end line (%d)", opts.Start, opts.End)
	}
	return nil
}

// hasRange returns true when the caller specified a line range filter.
func hasRange(opts Options) bool {
	return opts.Start > 0 || opts.End > 0
}

// tableBlock is a contiguous group of lines that form one markdown table.
type tableBlock struct {
	startIdx int // 0-based index into the lines slice
	endIdx   int // 0-based index (inclusive)
}

// detectTables scans lines and returns the index ranges of every markdown table.
// Lines inside fenced code blocks are never part of a table: a table there is
// example text that must stay as it is.
func detectTables(lines []string) []tableBlock {
	var blocks []tableBlock
	var fence string // opening fence of the current code block, or empty
	i := 0

	for i < len(lines) {
		if fence != "" {
			if closesFence(lines[i], fence) {
				fence = ""
			}
			i++
			continue
		}
		if f := openingFence(lines[i]); f != "" {
			fence = f
			i++
			continue
		}

		if !isTableRow(lines[i]) {
			i++
			continue
		}

		// Need at least a header and a separator row.
		if i+1 >= len(lines) || !isSeparatorRow(lines[i+1]) {
			i++
			continue
		}

		start := i
		i += 2 // skip header + separator

		// Consume remaining data rows.
		for i < len(lines) && isTableRow(lines[i]) {
			i++
		}

		blocks = append(blocks, tableBlock{startIdx: start, endIdx: i - 1})
	}

	return blocks
}

// applyAll reformats every detected table.
func applyAll(lines []string, tables []tableBlock, pad bool) string {
	result := make([]string, len(lines))
	copy(result, lines)

	// Process in reverse so index shifts don't affect earlier tables.
	for i := len(tables) - 1; i >= 0; i-- {
		result = replaceTable(result, tables[i], pad)
	}

	return strings.Join(result, "\n")
}

// applyRange reformats only tables overlapping the given line range.
func applyRange(lines []string, tables []tableBlock, opts Options) string {
	result := make([]string, len(lines))
	copy(result, lines)

	for i := len(tables) - 1; i >= 0; i-- {
		tb := tables[i]
		tableLine1 := tb.startIdx + 1 // convert to 1-based
		tableLine2 := tb.endIdx + 1

		if overlaps(tableLine1, tableLine2, opts.Start, opts.End) {
			result = replaceTable(result, tb, opts.Pad)
		}
	}

	return strings.Join(result, "\n")
}

// overlaps returns true when the table's 1-based line range overlaps with the
// filter range. A zero value in start/end means unbounded on that side.
func overlaps(tStart, tEnd, fStart, fEnd int) bool {
	if fStart > 0 && tEnd < fStart {
		return false
	}
	if fEnd > 0 && tStart > fEnd {
		return false
	}
	return true
}

// replaceTable parses a table block and splices the re-rendered lines back in.
func replaceTable(lines []string, tb tableBlock, pad bool) []string {
	raw := lines[tb.startIdx : tb.endIdx+1]
	t := ParseTable(raw, tb.startIdx+1)
	rendered := t.Render(pad)

	if len(rendered) == 0 {
		return lines
	}

	out := make([]string, 0, len(lines)-len(raw)+len(rendered))
	out = append(out, lines[:tb.startIdx]...)
	out = append(out, rendered...)
	out = append(out, lines[tb.endIdx+1:]...)

	return out
}

// openingFence returns the fence (a run of at least three backticks or tildes)
// that opens a fenced code block on line, or an empty string. The fence may be
// indented, so that code blocks in list items count, too. A backtick fence
// must not have a backtick in its info string (CommonMark rule).
func openingFence(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" || (trimmed[0] != '`' && trimmed[0] != '~') {
		return ""
	}

	n := len(trimmed) - len(strings.TrimLeft(trimmed, trimmed[:1]))
	if n < 3 {
		return ""
	}
	if trimmed[0] == '`' && strings.Contains(trimmed[n:], "`") {
		return ""
	}
	return trimmed[:n]
}

// closesFence returns true if line closes a code block that fence opened: a
// fence of the same character, at least as long as the opening one, with
// nothing but whitespace after it.
func closesFence(line, fence string) bool {
	trimmed := strings.TrimSpace(line)
	n := len(trimmed) - len(strings.TrimLeft(trimmed, fence[:1]))
	return n >= len(fence) && n == len(trimmed)
}

// isTableRow returns true if the line looks like a markdown table row.
func isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")
}

// isSeparatorRow returns true if the line is a markdown table separator
// (contains only pipes, dashes, colons, and spaces).
func isSeparatorRow(line string) bool {
	if !isTableRow(line) {
		return false
	}

	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")

	for _, ch := range trimmed {
		switch ch {
		case '-', ':', '|', ' ':
			continue
		default:
			return false
		}
	}

	// Must contain at least one dash.
	return strings.ContainsRune(trimmed, '-')
}
