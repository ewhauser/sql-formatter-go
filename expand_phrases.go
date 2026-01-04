package sqlformatter

import "strings"

func ExpandPhrases(phrases []string) []string {
	out := []string{}
	for _, phrase := range phrases {
		out = append(out, ExpandSinglePhrase(phrase)...)
	}
	return out
}

func ExpandSinglePhrase(phrase string) []string {
	combos := buildCombinations(parsePhrase(phrase))
	out := make([]string, 0, len(combos))
	for _, combo := range combos {
		out = append(out, stripExtraWhitespace(combo))
	}
	return out
}

func stripExtraWhitespace(text string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(text), " "))
}

type phraseNode interface{}

type conc struct{ items []phraseNode }

type mandatory struct{ items []phraseNode }

type optional struct{ items []phraseNode }

func parsePhrase(text string) phraseNode {
	items, _ := parseAlteration(text, 0, 0)
	return mandatory{items: items}
}

func parseAlteration(text string, index int, expectClosing rune) ([]phraseNode, int) {
	alterations := []phraseNode{}
	for index < len(text) {
		term, newIndex := parseConcatenation(text, index)
		alterations = append(alterations, term)
		index = newIndex
		if index < len(text) && text[index] == '|' {
			index++
			continue
		}
		if index < len(text) && (text[index] == '}' || text[index] == ']') {
			if expectClosing != 0 && rune(text[index]) != expectClosing {
				panic("Unbalanced parenthesis in: " + text)
			}
			index++
			return alterations, index
		}
		if index == len(text) {
			if expectClosing != 0 {
				panic("Unbalanced parenthesis in: " + text)
			}
			return alterations, index
		}
		panic("Unexpected \"" + string(text[index]) + "\"")
	}
	return alterations, index
}

func parseConcatenation(text string, index int) (phraseNode, int) {
	items := []phraseNode{}
	for {
		term, newIndex := parseTerm(text, index)
		if term == nil {
			break
		}
		items = append(items, term)
		index = newIndex
	}
	if len(items) == 1 {
		return items[0], index
	}
	return conc{items: items}, index
}

func parseTerm(text string, index int) (phraseNode, int) {
	if index >= len(text) {
		return nil, index
	}
	if text[index] == '{' {
		return parseMandatoryBlock(text, index+1)
	}
	if text[index] == '[' {
		return parseOptionalBlock(text, index+1)
	}
	word := strings.Builder{}
	for index < len(text) {
		ch := text[index]
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == ' ' {
			word.WriteByte(ch)
			index++
			continue
		}
		break
	}
	if word.Len() == 0 {
		return nil, index
	}
	return word.String(), index
}

func parseMandatoryBlock(text string, index int) (phraseNode, int) {
	items, newIndex := parseAlteration(text, index, '}')
	return mandatory{items: items}, newIndex
}

func parseOptionalBlock(text string, index int) (phraseNode, int) {
	items, newIndex := parseAlteration(text, index, ']')
	return optional{items: items}, newIndex
}

func buildCombinations(node phraseNode) []string {
	switch n := node.(type) {
	case string:
		return []string{n}
	case conc:
		out := []string{""}
		for _, item := range n.items {
			out = stringCombinations(out, buildCombinations(item))
		}
		return out
	case mandatory:
		out := []string{}
		for _, item := range n.items {
			out = append(out, buildCombinations(item)...)
		}
		return out
	case optional:
		out := []string{""}
		for _, item := range n.items {
			out = append(out, buildCombinations(item)...)
		}
		return out
	default:
		return []string{}
	}
}

func stringCombinations(xs, ys []string) []string {
	out := make([]string, 0, len(xs)*len(ys))
	for _, x := range xs {
		for _, y := range ys {
			out = append(out, x+y)
		}
	}
	return out
}
