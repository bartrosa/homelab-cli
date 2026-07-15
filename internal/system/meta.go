package system

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type metaEntry struct {
	Version   string
	Supported bool
	IsLTS     bool
}

func parseMetaRelease(data []byte) []metaEntry {
	var entries []metaEntry
	var cur metaEntry
	flush := func() {
		if cur.Version != "" {
			cur.IsLTS = strings.Contains(cur.Version, "LTS")
			entries = append(entries, cur)
		}
		cur = metaEntry{}
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, "Dist:") {
			flush()
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "Version":
			cur.Version = val
		case "Supported":
			cur.Supported = val == "1"
		}
	}
	flush()
	return entries
}

// ubuntuSeries returns "24.04" from "24.04.4 LTS" or "25.10".
func ubuntuSeries(version string) string {
	v := strings.TrimSuffix(version, " LTS")
	v = strings.TrimSpace(v)
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func ubuntuVersionTuple(series string) (major, minor int) {
	parts := strings.Split(series, ".")
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	return major, minor
}

// recentUbuntuSeries keeps desktop ISO targets users typically want (22.04+).
func recentUbuntuSeries(series string) bool {
	major, _ := ubuntuVersionTuple(series)
	return major >= 22
}

func compareUbuntuSeries(a, b string) int {
	am, an := ubuntuVersionTuple(a)
	bm, bn := ubuntuVersionTuple(b)
	if am != bm {
		return am - bm
	}
	return an - bn
}

func dedupeSeriesSortedDesc(series []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range series {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		return compareUbuntuSeries(out[i], out[j]) > 0
	})
	return out
}

func filterSupported(entries []metaEntry, ltsOnly, nonLTSOnly bool) []string {
	var series []string
	for _, e := range entries {
		if !e.Supported {
			continue
		}
		if ltsOnly && !e.IsLTS {
			continue
		}
		if nonLTSOnly && e.IsLTS {
			continue
		}
		series = append(series, ubuntuSeries(e.Version))
	}
	return dedupeSeriesSortedDesc(series)
}

func parseApacheIndex(data []byte) []string {
	var names []string
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.Index(line, `href="`); i >= 0 {
			rest := line[i+6:]
			if j := strings.Index(rest, `"`); j > 0 {
				names = append(names, rest[:j])
			}
		}
	}
	return names
}

func latestNumericDir(data []byte, minVersion int) (int, error) {
	highest := 0
	for _, name := range parseApacheIndex(data) {
		name = strings.TrimSuffix(name, "/")
		n, err := strconv.Atoi(name)
		if err != nil || n < minVersion {
			continue
		}
		if n > highest {
			highest = n
		}
	}
	if highest == 0 {
		return 0, fmt.Errorf("no release directories found")
	}
	return highest, nil
}
