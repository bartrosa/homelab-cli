package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

const (
	downloadBarWidth   = 28
	downloadRefresh    = 100 * time.Millisecond
	connectSpinEvery   = 120 * time.Millisecond
	nonTTYLogEvery     = 5 * time.Second
)

// DownloadReporter renders HTTP download progress (bar, speed, ETA).
type DownloadReporter struct {
	W       io.Writer
	NoColor bool

	mu    sync.Mutex
	total int64
	read  int64
	start time.Time

	connectCancel context.CancelFunc
	tick          *downloadWriter
}

// NewDownloadReporter builds a progress renderer for stdout.
func NewDownloadReporter(w io.Writer, noColor bool) *DownloadReporter {
	return &DownloadReporter{W: w, NoColor: noColor}
}

func (d *DownloadReporter) interactive() bool {
	f, ok := d.W.(*os.File)
	return ok && isatty.IsTerminal(f.Fd())
}

// BeginConnect shows a spinner until EndConnect is called.
func (d *DownloadReporter) BeginConnect(phase string) {
	if !d.interactive() {
		_, _ = fmt.Fprintf(d.W, "%s...\n", phase)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.connectCancel = cancel
	go d.spinConnect(ctx, phase)
}

// EndConnect stops the connecting spinner and clears its line.
func (d *DownloadReporter) EndConnect() {
	if d.connectCancel != nil {
		d.connectCancel()
		d.connectCancel = nil
		if d.interactive() {
			_, _ = fmt.Fprint(d.W, "\r\033[K")
		}
	}
}

func (d *DownloadReporter) spinConnect(ctx context.Context, phase string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if d.NoColor {
		frames = []string{"|", "/", "-", "\\"}
	}
	i := 0
	t := time.NewTicker(connectSpinEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			frame := frames[i%len(frames)]
			i++
			_, _ = fmt.Fprintf(d.W, "\r  %s %s", frame, phase)
		}
	}
}

// SetTotal sets expected size (-1 if unknown).
func (d *DownloadReporter) SetTotal(total int64) {
	d.mu.Lock()
	d.total = total
	d.read = 0
	d.start = time.Now()
	d.mu.Unlock()
}

// UpdateBytes sets absolute progress (for dd status=progress parsing).
func (d *DownloadReporter) UpdateBytes(read int64) {
	d.mu.Lock()
	if read < d.read {
		d.mu.Unlock()
		return
	}
	if d.start.IsZero() {
		d.start = time.Now()
	}
	d.read = read
	total := d.total
	start := d.start
	d.mu.Unlock()

	now := time.Now()
	if d.tick == nil {
		d.tick = &downloadWriter{d: d, first: true}
	}
	dw := d.tick
	if dw.first || now.Sub(dw.lastRender) >= downloadRefresh {
		dw.render(now, read, total, start, false)
		dw.lastRender = now
		dw.first = false
	}
}

// Writer returns an io.Writer that tracks bytes and refreshes the display.
func (d *DownloadReporter) Writer() io.Writer {
	if d.tick == nil {
		d.tick = &downloadWriter{d: d, first: true}
	}
	return d.tick
}

type downloadWriter struct {
	d            *DownloadReporter
	lastRender   time.Time
	lastPlainLog time.Time
	lastRead     int64
	lastSpeed    float64
	first        bool
}

func (dw *downloadWriter) Write(p []byte) (int, error) {
	n := len(p)
	dw.d.mu.Lock()
	dw.d.read += int64(n)
	read := dw.d.read
	total := dw.d.total
	start := dw.d.start
	dw.d.mu.Unlock()

	now := time.Now()
	if dw.first || now.Sub(dw.lastRender) >= downloadRefresh {
		dw.render(now, read, total, start, false)
		dw.lastRender = now
		dw.first = false
	}
	return n, nil
}

func (dw *downloadWriter) render(now time.Time, read, total int64, start time.Time, final bool) {
	elapsed := now.Sub(start)
	if elapsed <= 0 {
		elapsed = time.Millisecond
	}
	var instant float64
	if dw.lastRender.IsZero() {
		instant = float64(read) / elapsed.Seconds()
	} else {
		dt := now.Sub(dw.lastRender).Seconds()
		if dt <= 0 {
			dt = 0.001
		}
		instant = float64(read-dw.lastRead) / dt
	}
	if dw.lastSpeed == 0 {
		dw.lastSpeed = instant
	} else {
		dw.lastSpeed = dw.lastSpeed*0.7 + instant*0.3
	}
	dw.lastRead = read

	if !dw.d.interactive() {
		if final || dw.lastPlainLog.IsZero() || now.Sub(dw.lastPlainLog) >= nonTTYLogEvery {
			dw.d.printPlain(read, total, dw.lastSpeed, final)
			dw.lastPlainLog = now
		}
		return
	}

	line := dw.d.formatLine(read, total, dw.lastSpeed, elapsed, final)
	if final {
		clearProgressLine(dw.d.W)
		_, _ = fmt.Fprintln(dw.d.W, line)
		return
	}
	_, _ = fmt.Fprintf(dw.d.W, "\033[2K\r%s", line)
}

func clearProgressLine(w io.Writer) {
	_, _ = fmt.Fprint(w, "\033[2K\r")
}

func (d *DownloadReporter) formatLine(read, total int64, speed float64, elapsed time.Duration, final bool) string {
	var parts []string

	if total > 0 {
		pct := float64(read) * 100 / float64(total)
		if pct > 100 {
			pct = 100
		}
		parts = append(parts, progressBar(pct, downloadBarWidth, d.NoColor))
		parts = append(parts, fmt.Sprintf("%5.1f%%", pct))
		parts = append(parts, fmt.Sprintf("%s / %s", FormatBytes(read), FormatBytes(total)))
	} else {
		parts = append(parts, progressBarIndeterminate(d.NoColor))
		parts = append(parts, FormatBytes(read))
	}

	if speed > 0 {
		parts = append(parts, FormatBytes(int64(speed))+"/s")
		if total > 0 && read < total {
			eta := time.Duration(float64(total-read)/speed) * time.Second
			parts = append(parts, "ETA "+FormatETA(eta))
		}
	} else if !final {
		parts = append(parts, "starting…")
	}

	if final && elapsed > 0 {
		parts = append(parts, "in "+FormatETA(elapsed))
	}

	return "  " + strings.Join(parts, "  ")
}

func (d *DownloadReporter) printPlain(read, total int64, speed float64, final bool) {
	if total > 0 {
		pct := int(read * 100 / total)
		_, _ = fmt.Fprintf(d.W, "  %d%%  %s / %s", pct, FormatBytes(read), FormatBytes(total))
	} else {
		_, _ = fmt.Fprintf(d.W, "  %s downloaded", FormatBytes(read))
	}
	if speed > 0 {
		_, _ = fmt.Fprintf(d.W, "  at %s/s", FormatBytes(int64(speed)))
	}
	if final {
		_, _ = fmt.Fprintln(d.W, "  (complete)")
	} else {
		_, _ = fmt.Fprintln(d.W)
	}
}

// Finish prints the final progress line and a short summary.
func (d *DownloadReporter) Finish() {
	d.mu.Lock()
	read, _, start := d.read, d.total, d.start
	d.mu.Unlock()
	if start.IsZero() {
		return
	}
	elapsed := time.Since(start)
	if d.interactive() {
		clearProgressLine(d.W)
	}
	if read > 0 {
		_, _ = fmt.Fprintf(d.W, "  done: %s", FormatBytes(read))
		if elapsed >= time.Second {
			_, _ = fmt.Fprintf(d.W, " in %s", FormatETA(elapsed))
		}
		_, _ = fmt.Fprintln(d.W)
	}
}

// RunWithSpinner runs fn while showing a spinner (checksum, GPG, etc.).
func RunWithSpinner(w io.Writer, noColor bool, phase string, fn func() error) error {
	rep := NewDownloadReporter(w, noColor)
	rep.BeginConnect(phase)
	defer rep.EndConnect()
	err := fn()
	if err != nil {
		return err
	}
	if !rep.interactive() {
		_, _ = fmt.Fprintf(w, "%s done\n", phase)
	}
	return nil
}

// FormatBytes renders a human-readable size.
func FormatBytes(b int64) string {
	if b < 0 {
		return "0 B"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div := int64(unit)
	exp := 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := float64(b) / float64(div)
	return fmt.Sprintf("%.1f %ciB", val, "KMGTPE"[exp])
}

// FormatETA renders a duration for progress display.
func FormatETA(d time.Duration) string {
	if d < 0 || d >= 24*time.Hour {
		return "—"
	}
	d = d.Round(time.Second)
	sec := int(d.Seconds())
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	if sec < 3600 {
		return fmt.Sprintf("%dm%02ds", sec/60, sec%60)
	}
	return fmt.Sprintf("%dh%02dm", sec/3600, (sec%3600)/60)
}

func progressBar(pct float64, width int, noColor bool) string {
	if width < 4 {
		width = 4
	}
	filled := int(pct * float64(width) / 100)
	if filled > width {
		filled = width
	}
	if filled == width && pct >= 100 {
		inner := strings.Repeat("=", width)
		return bracket(inner, noColor)
	}
	var inner strings.Builder
	if filled > 0 {
		inner.WriteString(strings.Repeat("=", filled))
	}
	if filled < width {
		inner.WriteByte('>')
		inner.WriteString(strings.Repeat(" ", width-filled-1))
	}
	return bracket(inner.String(), noColor)
}

func progressBarIndeterminate(noColor bool) string {
	return bracket(strings.Repeat("·", downloadBarWidth), noColor)
}

func bracket(inner string, noColor bool) string {
	if noColor {
		return "[" + inner + "]"
	}
	return "[" + inner + "]"
}
