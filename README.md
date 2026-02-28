# mdtf

A CLI tool that formats markdown tables.

## Installation

```bash
go install github.com/treimann/mdtf@latest
```

## Usage

```bash
mdtf [flags] [file]
```

Reads from a file argument or stdin. Writes formatted output to stdout by default.

### Flags

| Flag | Description |
| -------------- | --------------------------------------------------------- |
| `--no-pad` | Disable column padding (compact output) |
| `--range S:E` | Only format tables within lines S through E (1-based) |
| `-w, --write` | Write result back to the file instead of stdout |

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
- Preserves column alignment markers (`:---`, `:---:`, `---:`)
- Handles escaped pipes (`\|`) inside cells
- Normalizes uneven column counts to the header width
- Supports UTF-8 content
