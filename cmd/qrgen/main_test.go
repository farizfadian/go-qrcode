package main

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
	zxqr "github.com/makiuchi-d/gozxing/qrcode"
)

const sample = "https://github.com/farizfadian/go-qrcode"

func runArgs(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func decodeFile(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		t.Fatalf("bitmap: %v", err)
	}
	res, err := zxqr.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return res.GetText()
}

func TestWritesADecodablePNG(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	code, _, errOut := runArgs(t, "-out", out, sample)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	if got := decodeFile(t, out); got != sample {
		t.Errorf("decoded %q, want %q", got, sample)
	}
}

func TestFormatIsInferredFromTheOutputExtension(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.png", "b.svg", "c.jpg", "d.jpeg"} {
		out := filepath.Join(dir, name)
		if code, _, errOut := runArgs(t, "-out", out, sample); code != 0 {
			t.Fatalf("%s: exit %d, stderr: %s", name, code, errOut)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(b) == 0 {
			t.Errorf("%s is empty", name)
		}
		if strings.HasSuffix(name, ".svg") && !bytes.HasPrefix(b, []byte("<svg ")) {
			t.Errorf("%s is not SVG markup", name)
		}
	}
}

func TestExplicitFormatOverridesTheExtension(t *testing.T) {
	out := filepath.Join(t.TempDir(), "weird.dat")
	if code, _, errOut := runArgs(t, "-format", "svg", "-out", out, sample); code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("<svg ")) {
		t.Errorf("expected SVG markup, got %.40s", b)
	}
}

func TestMissingContentIsAUsageError(t *testing.T) {
	code, _, errOut := runArgs(t)
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "usage") {
		t.Errorf("stderr does not show usage: %s", errOut)
	}
}

func TestTooManyPositionalArgsIsAUsageError(t *testing.T) {
	if code, _, _ := runArgs(t, "one", "two"); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}

// The CLI must reject a shape this build cannot draw, and must say what it can
// draw. It reads that list from the library's registry, so the message stays
// correct as shapes are added without any edit here.
func TestUnknownShapeIsRejectedAndListsWhatIsAvailable(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	code, _, errOut := runArgs(t, "-dots", "sparkle", "-out", out, sample)
	if code == 0 {
		t.Fatal("exit 0; a shape this build cannot draw must fail")
	}
	if !strings.Contains(errOut, "square") {
		t.Errorf("stderr does not list available shapes: %s", errOut)
	}
}

// A shape that this build does draw must be accepted without any edit to the
// CLI, because the CLI reads the list from the library.
func TestNewlyLandedShapesAreAcceptedWithoutCLIChanges(t *testing.T) {
	for _, shape := range []string{"fluid", "fluid-line", "stripe", "star", "diamond"} {
		out := filepath.Join(t.TempDir(), shape+".png")
		if code, _, errOut := runArgs(t, "-dots", shape, "-out", out, sample); code != 0 {
			t.Errorf("-dots %s: exit %d, stderr: %s", shape, code, errOut)
		}
	}
}

func TestBadColourIsReported(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	code, _, errOut := runArgs(t, "-fg", "not-a-colour", "-out", out, sample)
	if code == 0 {
		t.Fatal("exit 0; an unparseable colour must fail")
	}
	if errOut == "" {
		t.Error("no error message written to stderr")
	}
}

func TestVersionFlagPrintsTheLibraryVersion(t *testing.T) {
	code, stdout, _ := runArgs(t, "-version")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !strings.Contains(stdout, "qrgen") {
		t.Errorf("stdout = %q, want it to name the tool", stdout)
	}
}

func TestHelpListsTheShapesThisBuildSupports(t *testing.T) {
	code, _, errOut := runArgs(t, "-h")
	if code != 0 {
		t.Errorf("exit = %d, want 0 for -h", code)
	}
	if !strings.Contains(errOut, "square") {
		t.Errorf("help does not list the dot shapes: %s", errOut)
	}
}
