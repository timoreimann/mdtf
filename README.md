# mdtf

A CLI tool that formats markdown tables.

## Installation

```bash
go install github.com/timoreimann/mdtf@latest
```

## Usage

```bash
mdtf [flags] [file]
```

Reads from a file argument or stdin. Writes formatted output to stdout by default.

### Flags

| Flag          | Description                                           |
| ------------- | ----------------------------------------------------- |
| `--no-pad`    | Disable column padding (compact output)               |
| `--range S:E` | Only format tables within lines S through E (1-based) |
| `-w, --write` | Write result back to the file instead of stdout       |

### Examples

Format all tables in a file (padded, to stdout):

```bash
mdtf README.md
```

Format in place:

```bash
mdtf -w README.md
```

Compact output (no padding):

```bash
mdtf --no-pad README.md
```

Format only tables within a line range:

```bash
mdtf --range 5:12 README.md
```

Pipe from stdin:

```bash
cat README.md | mdtf
```

### Before / After

Input:

```markdown
| Name | Age | City |
|---|---|---|
| Alice | 30 | San Francisco |
| Bob | 25 | New York |
```

Padded (default):

```markdown
| Name  | Age | City          |
| ----- | --- | ------------- |
| Alice | 30  | San Francisco |
| Bob   | 25  | New York      |
```

Padded cells follow the column alignment. With `| :-- | :-: | --: |` as the separator row, the cells are left-aligned, centered, and right-aligned:

```markdown
| Name  | Age |          City |
| :---- | :-: | ------------: |
| Alice | 30  | San Francisco |
| Bob   | 25  |      New York |
```

No padding (`--no-pad`):

```markdown
| Name | Age | City |
| --- | --- | --- |
| Alice | 30 | San Francisco |
| Bob | 25 | New York |
```

## Features

- Formats all tables in a file, or only those within a specified line range
- Padded (aligned columns) and compact (no padding) modes
- Keeps the column alignment markers (`:---`, `:---:`, `---:`) and aligns padded cells to them
- Measures the display width of cell content, so that wide characters (CJK, emoji) line up
- Leaves tables in fenced code blocks (` ``` ` or `~~~`) unchanged
- Keeps the indentation of tables, for example in list items
- Handles escaped pipes (`\|`) inside cells
- Normalizes uneven column counts to the header width: it fills short rows and drops the extra cells of long rows

## Limitations

- A table row must start and end with `|`.
- Tables in block quotes (`> | a |`) are not detected.
- Tables in indented code blocks (four spaces, no fence) are formatted like other tables.
