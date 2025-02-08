package template

import "strings"

func ReplaceAllPairs(str string, pairs ...string) string {
	if len(pairs) < 2 {
		return str
	}

	for i := 0; i < len(pairs)-1; i += 2 {
		replacer, subStr := pairs[i], pairs[i+1]
		str = strings.ReplaceAll(str, replacer, subStr)
	}

	return str
}
