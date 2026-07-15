//go:build linux

package iso

import (
	"context"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bartrosa/homelab-cli/internal/ui"
)

const ddPollInterval = 250 * time.Millisecond

var ddBytesRe = regexp.MustCompile(`(?i)(\d+)\s+(?:bytes|bajt)`)

func parseDDBytes(line string) (int64, bool) {
	lower := strings.ToLower(line)
	if !strings.Contains(lower, "bytes") && !strings.Contains(lower, "bajt") {
		return 0, false
	}
	m := ddBytesRe.FindStringSubmatch(line)
	if len(m) < 2 {
		return 0, false
	}
	n, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func readDDStatus(r io.Reader) error {
	buf := make([]byte, 4096)
	var rem []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			rem = append(rem, buf[:n]...)
			for {
				split := -1
				for i, b := range rem {
					if b == '\r' || b == '\n' {
						split = i
						break
					}
				}
				if split < 0 {
					break
				}
				rem = rem[split+1:]
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func pollBlockWriteProgress(ctx context.Context, device string, baseline, total int64, rep *ui.DownloadReporter, spinner *sync.Once) {
	t := time.NewTicker(ddPollInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w, err := blockWriteBytes(device)
			if err != nil {
				continue
			}
			written := w - baseline
			if written < 0 {
				continue
			}
			if total > 0 && written > total {
				written = total
			}
			spinner.Do(func() { rep.EndConnect() })
			rep.UpdateBytes(written)
		}
	}
}

func buildDDCommand(ddArgs []string) (string, []string) {
	useSudo := os.Geteuid() != 0
	if _, err := osexec.LookPath("stdbuf"); err == nil {
		argv := append([]string{"-oL", "-eL", "dd"}, ddArgs...)
		if useSudo {
			return "sudo", append([]string{"stdbuf"}, argv...)
		}
		return "stdbuf", argv
	}
	if useSudo {
		return "sudo", append([]string{"dd"}, ddArgs...)
	}
	return "dd", ddArgs
}

func runDDWithProgress(ctx context.Context, out io.Writer, noColor bool, totalBytes int64, device string, ddArgs []string) error {
	baseline, err := blockWriteBytes(device)
	if err != nil {
		return fmt.Errorf("read block stats for %s: %w", device, err)
	}

	rep := ui.NewDownloadReporter(out, noColor)
	rep.SetTotal(totalBytes)
	rep.BeginConnect("Writing ISO to USB")
	defer rep.EndConnect()

	name, argv := buildDDCommand(ddArgs)
	cmd := osexec.CommandContext(ctx, name, argv...)
	cmd.Stdout = io.Discard
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	pollCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var spinner sync.Once
	go pollBlockWriteProgress(pollCtx, device, baseline, totalBytes, rep, &spinner)

	readDone := make(chan error, 1)
	go func() {
		readDone <- readDDStatus(stderrPipe)
	}()

	waitErr := cmd.Wait()
	cancel()
	readErr := <-readDone
	rep.EndConnect()
	if waitErr == nil && totalBytes > 0 {
		rep.UpdateBytes(totalBytes)
	}
	rep.Finish()
	if waitErr != nil {
		return fmt.Errorf("dd: %w", waitErr)
	}
	if readErr != nil {
		return readErr
	}
	return nil
}
