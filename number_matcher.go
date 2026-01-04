package sqlformatter

import "unicode"

type NumberMatcher struct {
	AllowUnderscore bool
}

func (m *NumberMatcher) Match(input string, index int) (string, bool) {
	if index >= len(input) {
		return "", false
	}
	start := index
	i := index
	if input[i] == '-' {
		i++
		// optional whitespace after minus
		for i < len(input) && (input[i] == ' ' || input[i] == '\t' || input[i] == '\n' || input[i] == '\r') {
			i++
		}
	}
	if i >= len(input) {
		return "", false
	}
	// hex or binary
	if hasPrefixFold(input[i:], "0x") {
		j := i + 2
		if !m.scanDigits(&j, input, func(r rune) bool { return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') }, m.AllowUnderscore) {
			return "", false
		}
		if !m.validNumberBoundary(input, j) {
			return "", false
		}
		return input[start:j], true
	}
	if hasPrefixFold(input[i:], "0b") {
		j := i + 2
		if !m.scanDigits(&j, input, func(r rune) bool { return r == '0' || r == '1' }, m.AllowUnderscore) {
			return "", false
		}
		if !m.validNumberBoundary(input, j) {
			return "", false
		}
		return input[start:j], true
	}

	j := i
	readDigits := m.scanDigits(&j, input, func(r rune) bool { return unicode.IsDigit(r) }, m.AllowUnderscore)
	if j < len(input) && input[j] == '.' {
		j++
		readDigitsAfter := m.scanDigits(&j, input, func(r rune) bool { return unicode.IsDigit(r) }, m.AllowUnderscore)
		if !readDigits && !readDigitsAfter {
			return "", false
		}
		readDigits = true
	} else if !readDigits {
		// allow leading .
		if i < len(input) && input[i] == '.' {
			j = i + 1
			if !m.scanDigits(&j, input, func(r rune) bool { return unicode.IsDigit(r) }, m.AllowUnderscore) {
				return "", false
			}
			readDigits = true
		} else {
			return "", false
		}
	}

	// exponent
	if j < len(input) && (input[j] == 'e' || input[j] == 'E') {
		k := j + 1
		if k < len(input) && (input[k] == '+' || input[k] == '-') {
			k++
		}
		if !m.scanDigits(&k, input, func(r rune) bool { return unicode.IsDigit(r) }, m.AllowUnderscore) {
			return "", false
		}
		j = k
	}

	if !readDigits {
		return "", false
	}
	if !m.validNumberBoundary(input, j) {
		return "", false
	}
	return input[start:j], true
}

func (m *NumberMatcher) scanDigits(idx *int, input string, isDigit func(rune) bool, allowUnderscore bool) bool {
	start := *idx
	for *idx < len(input) {
		r, size := utf8DecodeRuneInString(input[*idx:])
		if r == '_' && allowUnderscore {
			*idx += size
			continue
		}
		if !isDigit(r) {
			break
		}
		*idx += size
	}
	return *idx > start
}

func (m *NumberMatcher) validNumberBoundary(input string, index int) bool {
	if index >= len(input) {
		return true
	}
	r, _ := utf8DecodeRuneInString(input[index:])
	return !(unicode.IsLetter(r) || r == '_')
}
