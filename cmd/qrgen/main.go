// Command qrgen renders a styled QR code to a PNG, JPEG or SVG file.
//
//	qrgen -out logo.png "https://example.com"
//	qrgen -out card.svg -width 512 -fg '#1f2937' "hello"
//
// The output format follows the -out extension unless -format says otherwise.
// The shapes qrgen accepts are read from the library's shape registry, so the
// help text always lists exactly what this build can draw.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/farizfadian/go-qrcode/qr"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the whole program, parameterised over its streams so tests can drive
// it without touching os.Exit or the real standard streams.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("qrgen", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		out         = fs.String("out", "qr.png", "output file; its extension selects the format unless -format is given")
		format      = fs.String("format", "", "output format: png, jpeg or svg (default: from the -out extension)")
		width       = fs.Int("width", 0, "image size in pixels")
		margin      = fs.Int("margin", 0, "quiet zone in modules")
		ecc         = fs.String("ecc", "auto", "error correction: auto, L, M, Q or H")
		fg          = fs.String("fg", "", "foreground colour, hex with or without '#'")
		bg          = fs.String("bg", "", "background colour; '#00000000' for transparent")
		dots        = fs.String("dots", "square", "dot shape: "+strings.Join(qr.DotTypeNames(), ", "))
		dotColor    = fs.String("dot-color", "", "dot colour; defaults to -fg")
		corners     = fs.String("corners", "square", "finder shape: "+strings.Join(qr.CornerTypeNames(), ", "))
		cornerColor = fs.String("corner-color", "", "finder colour; defaults to -fg")
		quality     = fs.Int("quality", 92, "JPEG quality, 1 to 100")
		showVersion = fs.Bool("version", false, "print the version and exit")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: qrgen [flags] <content>\n\n")
		fmt.Fprintf(stderr, "Renders a styled QR code to a PNG, JPEG or SVG file.\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if *showVersion {
		fmt.Fprintf(stdout, "qrgen %s\n", qr.Version)
		return 0
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	content := fs.Arg(0)

	opts, err := buildOptions(content, *width, *margin, *ecc, *fg, *bg,
		*dots, *dotColor, *corners, *cornerColor)
	if err != nil {
		fmt.Fprintf(stderr, "qrgen: %v\n", err)
		return 1
	}

	code, err := qr.New(opts)
	if err != nil {
		fmt.Fprintf(stderr, "qrgen: %v\n", err)
		return 1
	}

	if err := write(code, *out, *format, *quality); err != nil {
		fmt.Fprintf(stderr, "qrgen: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "wrote %s (%d modules, ECC %v)\n", *out, code.Modules(), code.ECC())
	return 0
}

func buildOptions(content string, width, margin int, ecc, fg, bg,
	dots, dotColor, corners, cornerColor string) (qr.Options, error) {

	level, err := qr.ParseECCLevel(ecc)
	if err != nil {
		return qr.Options{}, err
	}
	dotType, err := qr.ParseDotType(dots)
	if err != nil {
		return qr.Options{}, err
	}
	cornerType, err := qr.ParseCornerType(corners)
	if err != nil {
		return qr.Options{}, err
	}
	return qr.Options{
		Content:    content,
		Width:      width,
		Margin:     margin,
		ECC:        level,
		Foreground: fg,
		Background: bg,
		Dots:       qr.DotOptions{Type: dotType, Color: dotColor},
		Corners:    qr.CornerOptions{Type: cornerType, Color: cornerColor},
	}, nil
}

// write renders code into path. An explicit format wins; otherwise the file
// extension decides, defaulting to PNG.
func write(code *qr.QR, path, format string, quality int) error {
	if format == "" {
		format = strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	switch format {
	case "svg":
		err = code.SVG(f)
	case "jpg", "jpeg":
		err = code.JPEG(f, quality)
	case "png", "":
		err = code.PNG(f)
	default:
		return fmt.Errorf("unknown format %q; want png, jpeg or svg", format)
	}
	if err != nil {
		return err
	}
	return f.Close()
}
