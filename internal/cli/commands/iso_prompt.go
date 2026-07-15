package commands

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func promptChoice(r io.Reader, w io.Writer, title string, options []string, defaultIdx int) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no options for %s", title)
	}
	if defaultIdx < 0 || defaultIdx >= len(options) {
		defaultIdx = 0
	}
	fmt.Fprintln(w, title)
	for i, o := range options {
		marker := " "
		if i == defaultIdx {
			marker = "*"
		}
		fmt.Fprintf(w, "  %s [%d] %s\n", marker, i+1, o)
	}
	fmt.Fprintf(w, "Choice [%d]: ", defaultIdx+1)

	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return -1, err
		}
		return defaultIdx, nil
	}
	line := strings.TrimSpace(sc.Text())
	if line == "" {
		return defaultIdx, nil
	}
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > len(options) {
		return -1, fmt.Errorf("invalid choice %q (enter 1-%d)", line, len(options))
	}
	return n - 1, nil
}

func promptLine(r io.Reader, w io.Writer, prompt, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(w, "%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Fprint(w, prompt)
	}
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return defaultVal, nil
	}
	line := strings.TrimSpace(sc.Text())
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}
