package iso

import "testing"

func TestBlockDevicePath(t *testing.T) {
	if got := BlockDevicePath("sda"); got != "/dev/sda" {
		t.Fatalf("got %q", got)
	}
	if got := BlockDevicePath("/dev/sda"); got != "/dev/sda" {
		t.Fatalf("got %q", got)
	}
}
