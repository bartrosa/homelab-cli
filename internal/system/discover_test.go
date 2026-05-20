package system

import (
	"testing"
)

func TestUbuntuDesktopISORegex(t *testing.T) {
	line := "bfd1cee02bc4f35db939e69b934ba49a39a378797ce9aee20f6e3e3e728fefbf *ubuntu-22.04.5-desktop-amd64.iso"
	m := ubuntuDesktopISO.FindStringSubmatch(line)
	if m == nil {
		t.Fatal("no match")
	}
	if m[2] != "ubuntu-22.04.5-desktop-amd64.iso" {
		t.Fatalf("got %q", m[2])
	}
}
