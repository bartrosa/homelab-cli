//go:build integration

package system

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestResolveUbuntuDesktop_2204(t *testing.T) {
	client := &http.Client{Timeout: 2 * time.Minute}
	img, err := resolveUbuntuDesktop(context.Background(), client, "22.04", "ubuntu-lts")
	if err != nil {
		t.Fatal(err)
	}
	if img.spec.isoFile == "" {
		t.Fatal("empty iso")
	}
	t.Log(img.spec.isoURL)
}
