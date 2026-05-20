// Package ui provides consistent terminal styling for lab commands.
package ui

import (
	"fmt"
	"io"
	"os"
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
	useColor := !noColor
	if f, ok := w.(*os.File); ok && !lipgloss.HasDarkBackground() {
		_ = f
	}
	if !useColor {
		plain := lipgloss.NewStyle()
		return Styles{
			Title: plain, Subtitle: plain, OK: plain, Warn: plain,
			Err: plain, Dim: plain, Accent: plain, Border: plain,
		}
	}
	return Styles{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		Subtitle: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		OK:       lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		Warn:     lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		Err:      lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Accent:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		Border:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
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

func Warn(w io.Writer, s Styles, msg string) {
	_, _ = fmt.Fprintln(w, s.Warn.Render("! "+msg))
}

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
		colW[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colW) && len(cell) > colW[i] {
				colW[i] = len(cell)
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
				b.WriteString(strings.Repeat(" ", colW[i]-len(c)))
			} else {
				b.WriteString(c)
			}
		}
		return b.String()
	}
	_, _ = fmt.Fprintln(w, s.Dim.Render(pad(headers)))
	for _, row := range rows {
		_, _ = fmt.Fprintln(w, pad(row))
	}
}

// Box wraps content in a rounded border.
func Box(s Styles, content string) string {
	return s.Border.Render(content)
}

// NoColorFromCmd reads --no-color from cobra root when available.
func NoColorFromCmd(noColorFlag bool) bool {
	return noColorFlag
}
