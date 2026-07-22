package podman

import "testing"

func TestUbuntuAtLeast(t *testing.T) {
	cases := []struct {
		ver   string
		ok    bool
		major int
		minor int
	}{
		{"24.04", true, 23, 10},
		{"23.10", true, 23, 10},
		{"23.04", false, 23, 10},
		{"22.04", false, 23, 10},
		{"bad", false, 23, 10},
	}
	for _, tc := range cases {
		if got := ubuntuAtLeast(tc.ver, tc.major, tc.minor); got != tc.ok {
			t.Fatalf("ubuntuAtLeast(%q)=%v want %v", tc.ver, got, tc.ok)
		}
	}
}

func TestParseLinger(t *testing.T) {
	if !parseLinger("Linger=yes") {
		t.Fatal("expected yes")
	}
	if parseLinger("Linger=no") {
		t.Fatal("expected no")
	}
}

func TestHasSubIDEntry(t *testing.T) {
	ok, err := hasSubIDEntry(func(string) ([]byte, error) {
		return []byte("root:0:1000\nalice:100000:65536\n"), nil
	}, "/etc/subuid", "alice")
	if err != nil || !ok {
		t.Fatalf("want alice present, ok=%v err=%v", ok, err)
	}
	ok, err = hasSubIDEntry(func(string) ([]byte, error) {
		return []byte("root:0:1000\n"), nil
	}, "/etc/subuid", "alice")
	if err != nil || ok {
		t.Fatalf("want alice absent, ok=%v err=%v", ok, err)
	}
}
