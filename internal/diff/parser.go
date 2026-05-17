package diff

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/aybykovskii/gitlab-tui/internal/mr"
)

var hunkHeaderPattern = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

func Parse(raw string) []mr.DiffRow {
	rows := []mr.DiffRow{}
	oldLine := 0
	newLine := 0

	for _, line := range strings.Split(raw, "\n") {
		if line == "" {
			continue
		}
		if matches := hunkHeaderPattern.FindStringSubmatch(line); matches != nil {
			oldLine = atoi(matches[1])
			newLine = atoi(matches[2])
			continue
		}
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "diff --git") {
			continue
		}
		if oldLine == 0 && newLine == 0 {
			continue
		}

		prefix := line[0]
		text := ""
		if len(line) > 1 {
			text = line[1:]
		}

		switch prefix {
		case ' ':
			rows = append(rows, mr.DiffRow{OldLine: oldLine, NewLine: newLine, OldText: text, NewText: text})
			oldLine++
			newLine++
		case '-':
			rows = append(rows, mr.DiffRow{OldLine: oldLine, OldText: text})
			oldLine++
		case '+':
			rows = append(rows, mr.DiffRow{NewLine: newLine, NewText: text})
			newLine++
		}
	}

	return rows
}

func atoi(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
