package sqlformatter

import (
	"sort"
	"strings"
)

func Dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func Last[T any](items []T) (T, bool) {
	var zero T
	if len(items) == 0 {
		return zero, false
	}
	return items[len(items)-1], true
}

func SortByLengthDesc(items []string) []string {
	sort.SliceStable(items, func(i, j int) bool {
		if len(items[i]) != len(items[j]) {
			return len(items[i]) > len(items[j])
		}
		return items[i] < items[j]
	})
	return items
}

func MaxLength(items []string) int {
	max := 0
	for _, item := range items {
		if len(item) > max {
			max = len(item)
		}
	}
	return max
}

func EqualizeWhitespace(text string) string {
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}

func IsMultiline(text string) bool {
	return strings.Contains(text, "\n")
}
