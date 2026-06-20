package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"uisketch/layout"
	md "uisketch/markdown"
	"uisketch/render/ascii"
	"uisketch/render/svg"
	"uisketch/sketch"
	"uisketch/validate"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: uisketch render [--format svg|ascii] [--output <file>] [input]\n       uisketch markdown [--output <doc.md> | --overwrite] [--format svg|ascii|source] [--asset-dir assets] [input]")
	}
	switch args[0] {
	case "render":
		return runRender(args[1:])
	case "markdown":
		return runMarkdown(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "", "render format")
	output := fs.String("output", "", "output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	renderFormat := inferRenderFormat(*format, *output)
	input, layoutIndex, err := renderArgs(fs)
	if err != nil {
		return err
	}
	body, sourcePath, err := readInput(input)
	if err != nil {
		return err
	}
	source, err := md.SelectSource(string(body), layoutIndex)
	if err != nil {
		return err
	}
	node, err := layout.ParseYAML(source.Source)
	if err != nil {
		return fmt.Errorf("%s source %d: %w", sourcePath, layoutIndex, err)
	}
	findings := validate.ValidateLayout(node, nil)
	for _, finding := range findings {
		fmt.Fprintf(stderr, "%s: %s\n", finding.Severity, finding.Message)
	}
	if validate.HasErrors(findings) {
		return fmt.Errorf("validation failed")
	}
	doc := sketch.FromLayout("", node)
	var result string
	switch renderFormat {
	case "ascii":
		result = ascii.Render(doc)
	case "svg":
		result = svg.Render(doc)
	default:
		return fmt.Errorf("format %q is not implemented; supported formats: ascii, svg", renderFormat)
	}
	if *output != "" {
		return os.WriteFile(*output, []byte(result+"\n"), 0644)
	}
	fmt.Fprintln(stdout, result)
	return nil
}

func runMarkdown(args []string) error {
	fs := flag.NewFlagSet("markdown", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("output", "", "markdown output file")
	overwrite := fs.Bool("overwrite", false, "overwrite input markdown file")
	format := fs.String("format", "source", "markdown output format")
	assetDir := fs.String("asset-dir", "assets", "relative directory for generated svg assets")
	if err := fs.Parse(args); err != nil {
		return err
	}
	input, err := inputArg(fs)
	if err != nil {
		return err
	}
	if (*output == "" && !*overwrite) || (*output != "" && *overwrite) {
		return fmt.Errorf("specify exactly one of --output or --overwrite")
	}
	if *overwrite && (input == "" || input == "-") {
		return fmt.Errorf("--overwrite requires a filesystem input path")
	}
	body, _, err := readInput(input)
	if err != nil {
		return err
	}
	result, err := md.Build(string(body), md.Options{AssetDir: *assetDir, Format: *format})
	if err != nil {
		return err
	}
	destination := *output
	if *overwrite {
		destination = input
	}
	if err := os.WriteFile(destination, []byte(result.Markdown+"\n"), 0644); err != nil {
		return err
	}
	baseDir := filepath.Dir(destination)
	for _, assetPath := range md.AssetPaths(result) {
		fullPath := filepath.Join(baseDir, filepath.FromSlash(assetPath))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(result.Assets[assetPath]), 0644); err != nil {
			return err
		}
	}
	return nil
}

func inputArg(fs *flag.FlagSet) (string, error) {
	switch fs.NArg() {
	case 0:
		return "", nil
	case 1:
		return fs.Arg(0), nil
	default:
		return "", fmt.Errorf("expected at most one input argument")
	}
}

func renderArgs(fs *flag.FlagSet) (string, int, error) {
	switch fs.NArg() {
	case 0:
		return "", 1, nil
	case 1:
		return fs.Arg(0), 1, nil
	case 2:
		index, err := strconv.Atoi(fs.Arg(1))
		if err != nil || index < 1 {
			return "", 0, fmt.Errorf("layout index must be a positive integer")
		}
		return fs.Arg(0), index, nil
	default:
		return "", 0, fmt.Errorf("expected input path and optional 1-origin layout index")
	}
}

func inferRenderFormat(format, output string) string {
	if format != "" {
		return format
	}
	if filepath.Ext(output) == ".svg" {
		return "svg"
	}
	return "svg"
}

func readInput(path string) ([]byte, string, error) {
	if path == "" || path == "-" {
		body, err := io.ReadAll(stdin)
		return body, "<stdin>", err
	}
	body, err := os.ReadFile(path)
	return body, path, err
}
