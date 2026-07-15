package iso

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// VerifyGPG verifies the checksum file signature using upstream signing keys.
func VerifyGPG(ctx context.Context, rel Release) error {
	needsVerify := rel.ChecksumSigKind == SigClearsigned || rel.ChecksumSigURL != ""
	if !needsVerify {
		return nil
	}
	if rel.ChecksumSigKind == SigClearsigned && len(rel.SigningKeyIDs) == 0 && len(rel.SigningKeyURLs) == 0 {
		return fmt.Errorf("clearsigned checksum requires signing keys")
	}
	if rel.ChecksumSigKind != SigClearsigned && rel.ChecksumSigURL == "" {
		return fmt.Errorf("detached signature URL missing")
	}

	client := &http.Client{Timeout: 2 * time.Minute}
	tmp, err := os.MkdirTemp("", "lab-iso-gpg-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	gpgHome := filepath.Join(tmp, "gnupg")
	if err := os.MkdirAll(gpgHome, 0o700); err != nil {
		return err
	}

	var gpgOut bytes.Buffer
	runner := exec.NewOSRunner(io.Discard, &gpgOut)
	runner.Env = []string{"GNUPGHOME=" + gpgHome}

	if err := importSigningKeys(ctx, client, runner, tmp, rel); err != nil {
		return fmt.Errorf("import signing keys: %w", err)
	}

	sumPath := filepath.Join(tmp, "checksums")
	sumData, err := fetchBytes(ctx, client, rel.ChecksumURL)
	if err != nil {
		return fmt.Errorf("checksum file: %w", err)
	}
	if err := os.WriteFile(sumPath, sumData, 0o644); err != nil {
		return err
	}

	switch rel.ChecksumSigKind {
	case SigClearsigned:
		return verifyClearsigned(ctx, runner, &gpgOut, tmp, sumPath)
	default:
		sigPath := filepath.Join(tmp, "checksums.sig")
		sigData, err := fetchBytes(ctx, client, rel.ChecksumSigURL)
		if err != nil {
			return fmt.Errorf("signature file: %w", err)
		}
		if err := os.WriteFile(sigPath, sigData, 0o644); err != nil {
			return err
		}
		return verifyDetached(ctx, runner, &gpgOut, tmp, sigPath, sumPath)
	}
}

func importSigningKeys(ctx context.Context, client *http.Client, runner *exec.OSRunner, tmp string, rel Release) error {
	for i, url := range rel.SigningKeyURLs {
		keyData, err := fetchBytes(ctx, client, url)
		if err != nil {
			return fmt.Errorf("fetch key %s: %w", url, err)
		}
		keyPath := filepath.Join(tmp, fmt.Sprintf("keyring-%d.gpg", i))
		if err := os.WriteFile(keyPath, keyData, 0o600); err != nil {
			return err
		}
		if err := runner.Run(ctx, "gpg", "--batch", "--import", keyPath); err != nil {
			return fmt.Errorf("gpg import %s: %w", url, err)
		}
	}

	for _, rawID := range rel.SigningKeyIDs {
		keyID := normalizeKeyID(rawID)
		if keyID == "" {
			continue
		}
		if err := importKeyByID(ctx, client, runner, tmp, keyID); err != nil {
			return err
		}
	}
	return nil
}

func importKeyByID(ctx context.Context, client *http.Client, runner *exec.OSRunner, tmp, keyID string) error {
	keyURL := fmt.Sprintf("https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x%s", keyID)
	keyData, err := fetchBytes(ctx, client, keyURL)
	if err == nil && looksLikePGPKey(keyData) {
		keyPath := filepath.Join(tmp, keyID+".asc")
		if err := os.WriteFile(keyPath, keyData, 0o600); err != nil {
			return err
		}
		if err := runner.Run(ctx, "gpg", "--batch", "--import", keyPath); err != nil {
			return fmt.Errorf("gpg import 0x%s: %w", keyID, err)
		}
		return nil
	}

	servers := []string{
		"hkp://keyserver.ubuntu.com:80",
		"hkp://keys.openpgp.org:80",
	}
	var lastErr error
	for _, srv := range servers {
		if err := runner.Run(ctx, "gpg", "--batch", "--keyserver", srv, "--recv-keys", "0x"+keyID); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("import key 0x%s: %w", keyID, lastErr)
}

func verifyDetached(ctx context.Context, runner *exec.OSRunner, gpgOut *bytes.Buffer, tmp, sigPath, sumPath string) error {
	return runGPGVerify(ctx, runner, gpgOut, tmp, sigPath, sumPath)
}

func verifyClearsigned(ctx context.Context, runner *exec.OSRunner, gpgOut *bytes.Buffer, tmp, sumPath string) error {
	return runGPGVerify(ctx, runner, gpgOut, tmp, "", sumPath)
}

func runGPGVerify(ctx context.Context, runner *exec.OSRunner, gpgOut *bytes.Buffer, tmp, sigPath, sumPath string) error {
	gpgOut.Reset()
	statusPath := filepath.Join(tmp, "gpg.status")
	args := []string{"--batch", "--keyid-format", "long", "--status-file", statusPath, "--verify"}
	if sigPath != "" {
		args = append(args, sigPath, sumPath)
	} else {
		args = append(args, sumPath)
	}
	err := runner.Run(ctx, "gpg", args...)
	status, _ := os.ReadFile(statusPath)
	return interpretGPGResult(string(status), gpgOut.String(), err)
}

func interpretGPGResult(status, textOut string, err error) error {
	if strings.Contains(status, "BADSIG") || strings.Contains(status, "ERRSIG") {
		return fmt.Errorf("bad signature (checksum file may be tampered): %s", strings.TrimSpace(firstNonEmpty(textOut, status)))
	}
	if strings.Contains(status, "GOODSIG") || strings.Contains(status, "VALIDSIG") {
		return nil
	}
	out := strings.ToLower(textOut)
	if strings.Contains(out, "bad signature") || strings.Contains(out, "zła sygnatura") {
		return fmt.Errorf("bad signature (checksum file may be tampered): %s", strings.TrimSpace(textOut))
	}
	if strings.Contains(out, "good signature") || strings.Contains(out, "poprawny podpis") {
		return nil
	}
	if err != nil {
		detail := strings.TrimSpace(firstNonEmpty(textOut, status))
		if detail != "" {
			return fmt.Errorf("%w: %s", err, detail)
		}
		return err
	}
	return fmt.Errorf("gpg did not report a good signature: %s", strings.TrimSpace(firstNonEmpty(textOut, status)))
}

func firstNonEmpty(parts ...string) string {
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			return p
		}
	}
	return ""
}

func normalizeKeyID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.TrimPrefix(id, "0x")
	id = strings.TrimPrefix(id, "0X")
	return strings.ToUpper(id)
}

func looksLikePGPKey(b []byte) bool {
	return strings.Contains(string(b), "BEGIN PGP PUBLIC KEY BLOCK")
}

// ParseSHA256SUMS extracts hash for filename from Ubuntu-style SHA256SUMS.
func ParseSHA256SUMS(content, filename string) (string, bool) {
	return findSHA256(content, filename)
}

// ParseFedoraCHECKSUM extracts hash from Fedora CHECKSUM file.
func ParseFedoraCHECKSUM(content, filename string) (string, bool) {
	return findFedoraChecksum(content, filename)
}

// ReadISOSize returns human-readable size.
func ReadISOSize(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return formatBytes(st.Size()), nil
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// CopyWithProgress is a test hook placeholder.
func CopyWithProgress(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
