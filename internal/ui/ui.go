// Package ui provides consistent terminal styling for lab commands.
package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for lab CLI output.
type Styles struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	OK       lipgloss.Style
	Warn     lipgloss.Style
	Err      lipgloss.Style
	Dim      lipgloss.Style
	Accent   lipgloss.Style
	Border   lipgloss.Style
}

// NewStyles builds the palette; disable color when noColor or not a TTY.
func NewStyles(w io.Writer, noColor bool) Styles {
	if noColor {
		plain := lipgloss.NewStyle()
		bold := lipgloss.NewStyle().Bold(true)
		return Styles{
			Title: bold, Subtitle: plain, OK: bold, Warn: bold,
			Err: bold, Dim: plain, Accent: bold, Border: plain,
		}
	}
	dim := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "241"})
	return Styles{
		Title:    lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "255"}),
		Subtitle: dim,
		OK:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "42"}),
		Warn:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "172", Dark: "214"}),
		Err:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "124", Dark: "196"}),
		Dim:      dim,
		Accent:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "25", Dark: "39"}),
		Border:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Padding(0, 1),
	}
}

// Section prints a titled block.
func Section(w io.Writer, s Styles, title, subtitle string) {
	if title != "" {
		_, _ = fmt.Fprintln(w, s.Title.Render(title))
	}
	if subtitle != "" {
		_, _ = fmt.Fprintln(w, s.Subtitle.Render(subtitle))
	}
}

// Step prints a progress line: [1/5] message.
func Step(w io.Writer, s Styles, cur, total int, msg string) {
	prefix := s.Accent.Render(fmt.Sprintf("[%d/%d]", cur, total))
	_, _ = fmt.Fprintf(w, "%s %s\n", prefix, msg)
}

// OK / Warn / Fail lines.
func OK(w io.Writer, s Styles, msg string) {
	_, _ = fmt.Fprintln(w, s.OK.Render("✓ "+msg))
}

// Warn prints a warning line.
func Warn(w io.Writer, s Styles, msg string) {
	_, _ = fmt.Fprintln(w, s.Warn.Render("! "+msg))
}

// Fail prints an error line.
func Fail(w io.Writer, s Styles, msg string) {
	_, _ = fmt.Fprintln(w, s.Err.Render("✗ "+msg))
}

// Table renders a simple two-column table.
func Table(w io.Writer, s Styles, headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	colW := make([]int, len(headers))
	for i, h := range headers {
		colW[i] = lipgloss.Width(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colW) {
				w := lipgloss.Width(cell)
				if w > colW[i] {
					colW[i] = w
				}
			}
		}
	}
	pad := func(cols []string) string {
		var b strings.Builder
		for i, c := range cols {
			if i > 0 {
				b.WriteString("  ")
			}
			if i < len(colW) {
				b.WriteString(c)
				b.WriteString(strings.Repeat(" ", colW[i]-lipgloss.Width(c)))
			} else {
				b.WriteString(c)
			}
		}
		return b.String()
	}
	_, _ = fmt.Fprintln(w, s.Title.Render(pad(headers)))
	for _, row := range rows {
		_, _ = fmt.Fprintln(w, pad(row))
	}
}

// KeyValue prints aligned label: value lines (e.g. version info).
func KeyValue(w io.Writer, s Styles, title string, pairs [][2]string) {
	if title != "" {
		_, _ = fmt.Fprintln(w, s.Title.Render(title))
	}
	maxKey := 0
	for _, p := range pairs {
		if len(p[0]) > maxKey {
			maxKey = len(p[0])
		}
	}
	for _, p := range pairs {
		label := fmt.Sprintf("%-*s", maxKey, p[0]+":")
		_, _ = fmt.Fprintf(w, "  %s %s\n", s.Dim.Render(label), p[1])
	}
}

// SecurityWarning prints a high-visibility security banner (e.g. failed GPG verify).
func SecurityWarning(w io.Writer, s Styles, noColor bool, title string, lines ...string) {
	banner := s.Warn
	if !noColor {
		banner = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "178", Dark: "220"}).
			Background(lipgloss.AdaptiveColor{Light: "228", Dark: "236"})
	}

	sep := strings.Repeat("=", 72)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, banner.Render(sep))
	_, _ = fmt.Fprintln(w, banner.Render("!!!  "+strings.ToUpper(title)+"  !!!"))
	_, _ = fmt.Fprintln(w, banner.Render(sep))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		_, _ = fmt.Fprintln(w, banner.Render("  "+line))
	}
	_, _ = fmt.Fprintln(w, banner.Render(sep))
	_, _ = fmt.Fprintln(w)
}

// WarnLine prints a single warning using standard styles.
func WarnLine(w io.Writer, noColor bool, msg string) {
	Warn(w, NewStyles(w, noColor), msg)
}

// OKLine prints a single success line using standard styles.
func OKLine(w io.Writer, noColor bool, msg string) {
	OK(w, NewStyles(w, noColor), msg)
}

// NoColorFromCmd reads --no-color from cobra root when available.
func NoColorFromCmd(noColorFlag bool) bool {
	return noColorFlag
}
