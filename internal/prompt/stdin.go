package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// StdinPrompter reads prompts from stdin with optional default display.
type StdinPrompter struct {
	In  io.Reader
	Out io.Writer
}

// NewStdinPrompter returns a prompter using os.Stdin/os.Stdout.
func NewStdinPrompter() *StdinPrompter {
	return &StdinPrompter{In: os.Stdin, Out: os.Stdout}
}

// AskString prompts for a line of text.
func (p *StdinPrompter) AskString(label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.Out, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(p.Out, "%s: ", label)
	}
	line, err := p.readLine()
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

// AskPassword prompts for a hidden password.
func (p *StdinPrompter) AskPassword(label string) (string, error) {
	fmt.Fprintf(p.Out, "%s: ", label)
	if f, ok := p.In.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.Out)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	line, err := p.readLine()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// AskBool prompts for yes/no.
func (p *StdinPrompter) AskBool(label string, defaultValue bool) (bool, error) {
	if defaultValue {
		fmt.Fprintf(p.Out, "%s [Y/n]: ", label)
	} else {
		fmt.Fprintf(p.Out, "%s [y/N]: ", label)
	}
	line, err := p.readLine()
	if err != nil {
		return false, err
	}
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultValue, nil
	}
	return line == "y" || line == "yes", nil
}

// AskSelect prompts for a numbered option.
func (p *StdinPrompter) AskSelect(label string, options []string, defaultIndex int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("no options for %s", label)
	}
	fmt.Fprintf(p.Out, "%s\n", label)
	for i, opt := range options {
		marker := " "
		if i == defaultIndex {
			marker = "*"
		}
		fmt.Fprintf(p.Out, "  %s %d) %s\n", marker, i+1, opt)
	}
	fmt.Fprintf(p.Out, "choice [%d]: ", defaultIndex+1)
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultIndex, nil
	}
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > len(options) {
		return 0, fmt.Errorf("invalid choice %q", line)
	}
	return n - 1, nil
}

// AskMultiSelect prompts for comma-separated option numbers.
func (p *StdinPrompter) AskMultiSelect(label string, options []string, defaultIndexes []int) ([]int, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("no options for %s", label)
	}
	defSet := map[int]struct{}{}
	for _, i := range defaultIndexes {
		defSet[i] = struct{}{}
	}
	fmt.Fprintf(p.Out, "%s (comma-separated numbers, empty for none)\n", label)
	for i, opt := range options {
		marker := " "
		if _, ok := defSet[i]; ok {
			marker = "*"
		}
		fmt.Fprintf(p.Out, "  %s %d) %s\n", marker, i+1, opt)
	}
	defStr := formatDefaultIndexes(defaultIndexes)
	if defStr != "" {
		fmt.Fprintf(p.Out, "choices [%s]: ", defStr)
	} else {
		fmt.Fprintf(p.Out, "choices: ")
	}
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return append([]int(nil), defaultIndexes...), nil
	}
	parts := strings.Split(line, ",")
	var idxs []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(options) {
			return nil, fmt.Errorf("invalid choice %q", part)
		}
		idxs = append(idxs, n-1)
	}
	return idxs, nil
}

func (p *StdinPrompter) readLine() (string, error) {
	sc := bufio.NewScanner(p.In)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return sc.Text(), nil
}

func formatDefaultIndexes(idxs []int) string {
	if len(idxs) == 0 {
		return ""
	}
	var parts []string
	for _, i := range idxs {
		parts = append(parts, strconv.Itoa(i+1))
	}
	return strings.Join(parts, ",")
}
