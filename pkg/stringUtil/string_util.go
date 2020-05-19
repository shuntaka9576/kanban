package stringUtil

import "strings"

func TrimSuffixLine(str string) string {
	return strings.TrimSuffix(str, "\n")
}
