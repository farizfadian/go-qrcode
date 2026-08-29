package main

import (
	"bytes"
	"image"
	"image/color"
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

// writeTestLogo puts a small PNG on disk so the -logo flag has something real
// to load.
func writeTestLogo(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			c := color.RGBA{0x0f, 0x76, 0x6e, 0xff}
			if (x/16+y/16)%2 == 0 {
				c = color.RGBA{0xff, 0xff, 0xff, 0xff}
			}
			img.Set(x, y, c)
		}
	}
	path := filepath.Join(t.TempDir(), "logo.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLogoFlagProducesADecodableCode(t *testing.T) {
	logo := writeTestLogo(t)
	out := filepath.Join(t.TempDir(), "qr.png")

	code, stdout, errOut := runArgs(t,
		"-logo", logo, "-width", "900", "-out", out, sample)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	if !strings.Contains(stdout, "logo hides") {
		t.Errorf("stdout does not report the logo budget: %s", stdout)
	}
	if got := decodeFile(t, out); got != sample {
		t.Errorf("decoded %q, want %q", got, sample)
	}
}

func TestLogoFlagsAreAllWiredUp(t *testing.T) {
	logo := writeTestLogo(t)
	out := filepath.Join(t.TempDir(), "qr.png")

	code, _, errOut := runArgs(t,
		"-logo", logo,
		"-logo-size", "0.18",
		"-logo-radius", "10",
		"-logo-border", "12",
		"-logo-border-radius", "16",
		"-logo-border-color", "#ffffff",
		"-logo-bg", "#ffffff",
		"-width", "900", "-out", out, sample)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	if got := decodeFile(t, out); got != sample {
		t.Errorf("decoded %q, want %q", got, sample)
	}
}

// An oversized logo must be refused with the library's explanation, not
// silently shrunk or written out unscannable.
func TestOversizedLogoFlagIsRejected(t *testing.T) {
	logo := writeTestLogo(t)
	out := filepath.Join(t.TempDir(), "qr.png")

	code, _, errOut := runArgs(t,
		"-logo", logo, "-logo-size", "0.8", "-width", "900", "-out", out, sample)
	if code == 0 {
		t.Fatal("exit 0; an oversized logo must fail")
	}
	if !strings.Contains(errOut, "error correction") {
		t.Errorf("stderr does not explain the budget: %s", errOut)
	}
}

// The contrast rules must reach the command line too.
func TestInvertedColoursAreRejectedByTheCLI(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	code, _, errOut := runArgs(t, "-fg", "#ffffff", "-bg", "#000000", "-out", out, sample)
	if code == 0 {
		t.Fatal("exit 0; an inverted scheme must fail")
	}
	if !strings.Contains(errOut, "darker") {
		t.Errorf("stderr does not explain the polarity rule: %s", errOut)
	}
}

func TestWebPOutputIsWrittenAndReadable(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.webp")
	if code, _, errOut := runArgs(t, "-out", out, "-width", "500", sample); code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("RIFF")) {
		t.Fatalf("not a RIFF container: %.8q", b)
	}
	// The scan path must read back what the write path produced.
	code, stdout, errOut := runArgs(t, "-scan", out)
	if code != 0 {
		t.Fatalf("scan exit %d, stderr: %s", code, errOut)
	}
	if strings.TrimSpace(stdout) != sample {
		t.Errorf("scanned %q, want %q", strings.TrimSpace(stdout), sample)
	}
}

func TestWebPIsSmallerThanPNG(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "qr.png")
	webpPath := filepath.Join(dir, "qr.webp")

	for _, p := range []string{pngPath, webpPath} {
		if code, _, errOut := runArgs(t, "-out", p, "-width", "600", sample); code != 0 {
			t.Fatalf("%s: exit %d, stderr: %s", p, code, errOut)
		}
	}
	pngInfo, err := os.Stat(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	webpInfo, err := os.Stat(webpPath)
	if err != nil {
		t.Fatal(err)
	}
	if webpInfo.Size() >= pngInfo.Size() {
		t.Errorf("WebP is %d bytes, PNG is %d; the smaller format is the reason it exists",
			webpInfo.Size(), pngInfo.Size())
	}
}

func TestScanFlagPrintsTheContent(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	if code, _, errOut := runArgs(t, "-out", out, sample); code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	code, stdout, errOut := runArgs(t, "-scan", out)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	// Content alone on stdout, so the command composes with a shell pipeline.
	if strings.TrimSpace(stdout) != sample {
		t.Errorf("stdout = %q, want just the content", stdout)
	}
}

func TestScanDetailsPrintsTheMetadata(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	if code, _, errOut := runArgs(t, "-out", out, "-ecc", "H", sample); code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	code, stdout, errOut := runArgs(t, "-scan", out, "-scan-details")
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	for _, want := range []string{sample, "version:", "ecc:", "mask:", "segments:", "byte"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q:\n%s", want, stdout)
		}
	}
	if !strings.Contains(stdout, "ecc:      H") {
		t.Errorf("the reported ECC does not match what was requested:\n%s", stdout)
	}
}

func TestScanReportsAnUnreadableImage(t *testing.T) {
	blank := filepath.Join(t.TempDir(), "blank.png")
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
		}
	}
	f, err := os.Create(blank)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()

	if code, _, errOut := runArgs(t, "-scan", blank); code == 0 {
		t.Fatalf("exit 0 on an image with no code; stderr: %s", errOut)
	}
	if code, _, _ := runArgs(t, "-scan", "no-such-file.png"); code == 0 {
		t.Error("exit 0 on a missing file")
	}
}

func TestScanRejectsBeingGivenContentToo(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	if code, _, _ := runArgs(t, "-out", out, sample); code != 0 {
		t.Fatal("setup failed")
	}
	if code, _, errOut := runArgs(t, "-scan", out, "some content"); code != 2 {
		t.Errorf("exit = %d, want 2; stderr: %s", code, errOut)
	}
}

const testLogoSVGMarkup = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
	`<rect x="8" y="8" width="84" height="84" rx="16" fill="#0f766e"/></svg>`

func writeTestLogoSVG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "logo.svg")
	if err := os.WriteFile(path, []byte(testLogoSVGMarkup), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// -logo-svg upgrades SVG output without changing anything else.
func TestLogoSVGFlagEmbedsVectorArtInSVGOutput(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "qr.svg")

	code, _, errOut := runArgs(t,
		"-logo", writeTestLogo(t), "-logo-svg", writeTestLogoSVG(t),
		"-width", "900", "-out", out, sample)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`rx="16"`)) {
		t.Error("the vector logo was not embedded verbatim")
	}
	if bytes.Contains(b, []byte("data:image/png")) {
		t.Error("a bitmap was embedded even though vector art was given")
	}
}

// Raster output still needs the raster logo, so -logo-svg alone is refused up
// front rather than producing a code with no logo on it.
func TestLogoSVGFlagRequiresTheRasterLogoToo(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	code, _, errOut := runArgs(t,
		"-logo-svg", writeTestLogoSVG(t), "-out", out, sample)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "-logo") {
		t.Errorf("the error does not say what is missing: %s", errOut)
	}
}

func TestLogoSVGFlagReportsAMissingFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "qr.png")
	code, _, errOut := runArgs(t,
		"-logo", writeTestLogo(t), "-logo-svg", "no-such-file.svg",
		"-out", out, sample)
	if code == 0 {
		t.Fatal("exit 0 with a missing -logo-svg file")
	}
	if errOut == "" {
		t.Error("nothing written to stderr")
	}
}

// The usage text drifted out of date once already, claiming only PNG, JPEG and
// SVG after WebP and scanning had landed. Help that lies is worse than terse
// help, so it is pinned.
func TestUsageMentionsEveryCapability(t *testing.T) {
	_, _, errOut := runArgs(t, "-h")
	for _, want := range []string{"PNG", "JPEG", "WebP", "SVG", "-scan"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("usage text does not mention %q:\n%s", want, errOut)
		}
	}
}

// Every output format the writer accepts must be named in the -format help,
// and nothing else should be.
func TestFormatFlagHelpListsTheSupportedFormats(t *testing.T) {
	_, _, errOut := runArgs(t, "-h")
	for _, want := range []string{"png", "jpeg", "webp", "svg"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("-format help does not mention %q:\n%s", want, errOut)
		}
	}
}
