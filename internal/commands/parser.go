package commands

import (
	"fmt"
	"strings"
)

// parseArgs: поддержка кавычек и экранирования
func parseArgs(input string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inQuotes := false
	escape := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if escape {
			cur.WriteByte(ch)
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == ' ' && !inQuotes {
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteByte(ch)
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	if inQuotes {
		return nil, fmt.Errorf("unmatched quote")
	}
	return args, nil
}
