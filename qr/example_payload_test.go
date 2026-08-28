package qr_test

// A QR code stores text and nothing more. What makes a phone offer to join a
// network, save a contact or open a map is the *shape* of that text: a
// convention the scanner recognises. This library therefore needs to know
// nothing about these formats — you build the string, it encodes it.
//
// These examples are the formats worth knowing, each one compiled and run by
// `go test`, so they cannot rot.

import (
	"fmt"
	"log"
	"strings"

	"github.com/farizfadian/go-qrcode/qr"
)

// A plain URL is the most common payload. Scanners offer to open it.
func Example_payloadURL() {
	q, err := qr.New(qr.Options{Content: "https://example.com/promo?ref=poster"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(q.Modules(), "modules")
	// Output: 33 modules
}

// Anything that is not a recognised format is shown as text.
func Example_payloadText() {
	q, _ := qr.New(qr.Options{Content: "Meja 12 — Warung Kopi Senja"})
	fmt.Println(q.Content())
	// Output: Meja 12 — Warung Kopi Senja
}

// wifiEscape escapes the characters the WiFi format treats as syntax. Getting
// this wrong is the usual reason a WiFi QR code fails: a password containing a
// semicolon silently truncates the field.
func wifiEscape(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`;`, `\;`,
		`,`, `\,`,
		`:`, `\:`,
		`"`, `\"`,
	)
	return r.Replace(s)
}

// WIFI:T:<WPA|WEP|nopass>;S:<ssid>;P:<password>;H:<true if hidden>;;
// Scanning it offers to join the network.
func Example_payloadWiFi() {
	ssid, password := "Kopi Senja", "rahasia;123"

	content := fmt.Sprintf("WIFI:T:WPA;S:%s;P:%s;H:false;;",
		wifiEscape(ssid), wifiEscape(password))

	q, err := qr.New(qr.Options{Content: content, ECC: qr.ECCQuartile})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(content)
	fmt.Println(q.Modules(), "modules")
	// Output:
	// WIFI:T:WPA;S:Kopi Senja;P:rahasia\;123;H:false;;
	// 37 modules
}

// vCard is the portable contact format. Keep it short: every field makes the
// symbol bigger, and a bigger symbol at the same pixel width is harder to scan.
func Example_payloadVCard() {
	content := strings.Join([]string{
		"BEGIN:VCARD",
		"VERSION:3.0",
		"N:Fadian;Fariz;;;",
		"FN:Fariz Fadian",
		"ORG:Contoh Teknologi",
		"TITLE:Software Engineer",
		"TEL;TYPE=CELL:+6281234567890",
		"EMAIL:fariz@example.com",
		"URL:https://github.com/farizfadian",
		"END:VCARD",
	}, "\n")

	// Contact cards are long, so give the symbol room: raise Width rather than
	// letting the modules shrink.
	q, err := qr.New(qr.Options{Content: content, Width: 640})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(q.Modules(), "modules, ECC", q.ECC())
	// Output: 57 modules, ECC M
}

// MeCard says the same thing in far fewer bytes, which means a smaller and
// easier-to-scan symbol. Android reads it natively; iOS support is patchier.
func Example_payloadMeCard() {
	content := "MECARD:N:Fadian,Fariz;TEL:+6281234567890;EMAIL:fariz@example.com;;"

	q, _ := qr.New(qr.Options{Content: content})
	fmt.Println(q.Modules(), "modules — compare with the vCard above")
	// Output: 37 modules — compare with the vCard above
}

// mailto: opens a pre-filled email. Spaces must be percent-encoded.
func Example_payloadEmail() {
	content := "mailto:halo@example.com" +
		"?subject=Pesanan%20baru" +
		"&body=Halo%2C%20saya%20ingin%20memesan"

	q, _ := qr.New(qr.Options{Content: content})
	fmt.Println(q.Modules(), "modules")
	// Output: 37 modules
}

// SMSTO opens the messaging app with the number and text ready to send.
func Example_payloadSMS() {
	q, _ := qr.New(qr.Options{Content: "SMSTO:+6281234567890:Halo, saya tertarik"})
	fmt.Println(q.Modules(), "modules")
	// Output: 29 modules
}

// tel: starts a call. Use the full international form so it works abroad.
func Example_payloadPhone() {
	q, _ := qr.New(qr.Options{Content: "tel:+6281234567890"})
	fmt.Println(q.Modules(), "modules")
	// Output: 25 modules
}

// wa.me is just a URL, which is why it works on every scanner. The number
// carries no plus sign and no leading zero.
func Example_payloadWhatsApp() {
	content := "https://wa.me/6281234567890?text=Halo%2C%20saya%20ingin%20bertanya"

	q, _ := qr.New(qr.Options{Content: content})
	fmt.Println(q.Modules(), "modules")
	// Output: 37 modules
}

// geo: opens a map at the given latitude and longitude.
func Example_payloadLocation() {
	q, _ := qr.New(qr.Options{Content: "geo:-6.175392,106.827153"})
	fmt.Println(q.Modules(), "modules")
	// Output: 29 modules
}

// A calendar event, in the same iCalendar syntax a .ics file uses. Times are
// UTC when they end in Z.
func Example_payloadCalendarEvent() {
	content := strings.Join([]string{
		"BEGIN:VEVENT",
		"SUMMARY:Rapat Peluncuran",
		"LOCATION:Kantor Pusat",
		"DTSTART:20260915T020000Z",
		"DTEND:20260915T030000Z",
		"END:VEVENT",
	}, "\n")

	q, err := qr.New(qr.Options{Content: content, Width: 512})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(q.Modules(), "modules")
	// Output: 45 modules
}

// Longer payloads need a bigger image, not a higher error-correction level.
// More redundancy grows the symbol, which shrinks every module at a fixed
// width and makes scanning harder — the opposite of the intent.
func Example_payloadSizing() {
	long := strings.Repeat("data ", 120)

	tight, err := qr.New(qr.Options{Content: long, Width: 300})
	if err != nil {
		log.Fatal(err)
	}
	roomy, err := qr.New(qr.Options{Content: long, Width: 900})
	if err != nil {
		log.Fatal(err)
	}

	// Same symbol, very different pixels per module.
	fmt.Printf("300px: %d modules, %.1f px per module\n",
		tight.Modules(), 300.0/float64(tight.Modules()+8))
	fmt.Printf("900px: %d modules, %.1f px per module\n",
		roomy.Modules(), 900.0/float64(roomy.Modules()+8))
	// Output:
	// 300px: 93 modules, 3.0 px per module
	// 900px: 93 modules, 8.9 px per module
}
