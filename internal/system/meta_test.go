package system

import "testing"

func TestParseMetaRelease_LTS(t *testing.T) {
	data := []byte(`Dist: jammy
Version: 22.04.5 LTS
Supported: 1

Dist: noble
Version: 24.04.4 LTS
Supported: 1

Dist: old
Version: 20.04.6 LTS
Supported: 0
`)
	entries := parseMetaRelease(data)
	lts := filterSupported(entries, true, false)
	if len(lts) != 2 {
		t.Fatalf("got %v want 2 LTS", lts)
	}
	if lts[0] != "24.04" || lts[1] != "22.04" {
		t.Fatalf("order: %v", lts)
	}
}

func TestCompareUbuntuSeries(t *testing.T) {
	if compareUbuntuSeries("25.10", "24.04") <= 0 {
		t.Fatal("25.10 should be > 24.04")
	}
}
