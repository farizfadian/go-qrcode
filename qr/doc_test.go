package qr_test

import (
	"testing"

	"github.com/farizfadian/go-qrcode/qr"
)

func TestPackageBuilds(t *testing.T) {
	if qr.Version == "" {
		t.Fatal("qr.Version is empty; the package did not initialise")
	}
}
