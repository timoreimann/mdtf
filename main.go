package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/timoreimann/mdtf/pkg"

	flag "github.com/spf13/pflag"
)

func main() {
	noPad := flag.Bool("no-pad", false, "disable column padding (compact output)")
	lineRange := flag.String("range", "", "line range to format (e.g. 5:12); omit to format all tables")
	write := flag.BoolP("write", "w", false, "write result back to file instead of stdout")
	flag.Usage = usage
	flag.Parse()

	opts, err := buildOptions(*noPad, *lineRange)
	if err != nil {
		fatal(err)
	}

	content, filePath, err := readInput(flag.Args())
	if err != nil {
		fatal(err)
	}

	out, err := pkg.Format(content, opts)
	if err != nil {
		fatal(err)
	}

	if err := writeOutput(out, filePath, *write); err != nil {
		fatal(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: mdtf [flags] [file]\n\n")
	fmt.Fprintf(os.Stderr, "Formats markdown tables in the given file (or stdin).\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func buildOptions(noPad bool, lineRange string) (pkg.Options, error) {
	opts := pkg.Options{Pad: !noPad}

	if lineRange == "" {
		return opts, nil
	}

	start, end, err := parseRange(lineRange)
	if err != nil {
		return pkg.Options{}, err
	}
	opts.Start = start
	opts.End = end

	return opts, nil
}

func parseRange(s string) (int, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range %q: expected format start:end (e.g. 5:12)", s)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range start %q: %s", parts[0], err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range end %q: %s", parts[1], err)
	}

	return start, end, nil
}

func readInput(args []string) (string, string, error) {
	if len(args) == 0 {
		return readStdin()
	}
	return readFile(args[0])
}

func readStdin() (string, string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("reading stdin: %s", err)
	}
	return string(data), "", nil
}

func readFile(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("reading file: %s", err)
	}
	return string(data), path, nil
}

func writeOutput(content, filePath string, inPlace bool) error {
	if !inPlace || filePath == "" {
		_, err := fmt.Print(content)
		return err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file: %s", err)
	}

	if err := os.WriteFile(filePath, []byte(content), info.Mode()); err != nil {
		return fmt.Errorf("writing file: %s", err)
	}

	return nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "mdtf: %s\n", err)
	os.Exit(1)
}
