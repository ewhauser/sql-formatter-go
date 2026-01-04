package sqlformatter

import (
	"sort"
	"strings"
)

type phraseEntry struct {
	words     []string
	tokenType TokenType
	text      string
}

type phraseIndex map[string][]phraseEntry

type tokenClassifier struct {
	keywordPhrases      phraseIndex
	dataTypePhrases     phraseIndex
	dataTypeWordPhrases phraseIndex
	clausePhrases       phraseIndex
	selectPhrases       phraseIndex
	setOpPhrases        phraseIndex
	joinPhrases         phraseIndex

	reservedClauses       map[string]bool
	reservedSelect        map[string]bool
	reservedSetOperations map[string]bool
	reservedJoins         map[string]bool
	reservedKeywords      map[string]bool
	reservedDataTypes     map[string]bool
	reservedFunctionNames map[string]bool

	hasLimit    bool
	supportsXor bool
}

func newTokenClassifier(cfg TokenizerOptions) *tokenClassifier {
	dataTypePhraseItems := cfg.ReservedDataTypePhrases
	dataTypeWordPhraseItems := multiWordItems(cfg.ReservedDataTypes)
	if len(dataTypeWordPhraseItems) > 0 {
		dataTypeWordPhraseItems = Dedupe(dataTypeWordPhraseItems)
	}
	return &tokenClassifier{
		keywordPhrases:        buildPhraseIndex(cfg.ReservedKeywordPhrases, TokenReservedKeywordPhrase),
		dataTypePhrases:       buildPhraseIndex(dataTypePhraseItems, TokenReservedDataTypePhrase),
		dataTypeWordPhrases:   buildPhraseIndex(dataTypeWordPhraseItems, TokenReservedDataType),
		clausePhrases:         buildPhraseIndex(cfg.ReservedClauses, TokenReservedClause),
		selectPhrases:         buildPhraseIndex(cfg.ReservedSelect, TokenReservedSelect),
		setOpPhrases:          buildPhraseIndex(cfg.ReservedSetOperations, TokenReservedSetOperation),
		joinPhrases:           buildPhraseIndex(cfg.ReservedJoins, TokenReservedJoin),
		reservedClauses:       buildWordSet(cfg.ReservedClauses),
		reservedSelect:        buildWordSet(cfg.ReservedSelect),
		reservedSetOperations: buildWordSet(cfg.ReservedSetOperations),
		reservedJoins:         buildWordSet(cfg.ReservedJoins),
		reservedKeywords:      buildWordSet(cfg.ReservedKeywords),
		reservedDataTypes:     buildWordSet(cfg.ReservedDataTypes),
		reservedFunctionNames: buildWordSet(cfg.ReservedFunctionNames),
		hasLimit:              containsString(cfg.ReservedClauses, "LIMIT"),
		supportsXor:           cfg.SupportsXor,
	}
}

func buildPhraseIndex(phrases []string, tokenType TokenType) phraseIndex {
	if len(phrases) == 0 {
		return nil
	}
	index := make(phraseIndex)
	for _, phrase := range phrases {
		words := strings.Fields(phrase)
		if len(words) < 2 {
			continue
		}
		upper := make([]string, len(words))
		for i, w := range words {
			upper[i] = asciiUpper(w)
		}
		entry := phraseEntry{
			words:     upper,
			tokenType: tokenType,
			text:      toCanonical(phrase),
		}
		index[upper[0]] = append(index[upper[0]], entry)
	}
	for key := range index {
		entries := index[key]
		sort.Slice(entries, func(i, j int) bool {
			return len(entries[i].words) > len(entries[j].words)
		})
		index[key] = entries
	}
	return index
}

func buildWordSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	set := make(map[string]bool, len(items))
	for _, item := range items {
		words := strings.Fields(item)
		if len(words) != 1 {
			continue
		}
		set[asciiUpper(words[0])] = true
	}
	return set
}

func (c *tokenClassifier) Classify(tokens []Token) []Token {
	if len(tokens) == 0 {
		return tokens
	}
	tokens = mergePhrases(tokens, c.keywordPhrases)
	tokens = mergePhrases(tokens, c.dataTypeWordPhrases)
	tokens = mergePhrases(tokens, c.dataTypePhrases)
	tokens = mergePhrases(tokens, c.clausePhrases)
	tokens = mergePhrases(tokens, c.selectPhrases)
	tokens = mergePhrases(tokens, c.setOpPhrases)
	tokens = mergePhrases(tokens, c.joinPhrases)

	for i, tok := range tokens {
		if tok.Type != TokenIdentifier {
			continue
		}
		word := asciiUpper(tok.Text)
		switch word {
		case "CASE":
			tokens[i] = promoteToken(tok, TokenCase)
			continue
		case "END":
			tokens[i] = promoteToken(tok, TokenEnd)
			continue
		case "BETWEEN":
			tokens[i] = promoteToken(tok, TokenBetween)
			continue
		case "WHEN":
			tokens[i] = promoteToken(tok, TokenWhen)
			continue
		case "ELSE":
			tokens[i] = promoteToken(tok, TokenElse)
			continue
		case "THEN":
			tokens[i] = promoteToken(tok, TokenThen)
			continue
		case "AND":
			tokens[i] = promoteToken(tok, TokenAnd)
			continue
		case "OR":
			tokens[i] = promoteToken(tok, TokenOr)
			continue
		case "XOR":
			if c.supportsXor {
				tokens[i] = promoteToken(tok, TokenXor)
			}
			continue
		case "LIMIT":
			if c.hasLimit {
				tokens[i] = promoteToken(tok, TokenLimit)
				continue
			}
		}

		if c.reservedClauses[word] {
			tokens[i] = promoteToken(tok, TokenReservedClause)
			continue
		}
		if c.reservedSelect[word] {
			tokens[i] = promoteToken(tok, TokenReservedSelect)
			continue
		}
		if c.reservedSetOperations[word] {
			tokens[i] = promoteToken(tok, TokenReservedSetOperation)
			continue
		}
		if c.reservedJoins[word] {
			tokens[i] = promoteToken(tok, TokenReservedJoin)
			continue
		}
		if c.reservedFunctionNames[word] {
			tokens[i] = promoteToken(tok, TokenReservedFunctionName)
			continue
		}
		if c.reservedDataTypes[word] {
			tokens[i] = promoteToken(tok, TokenReservedDataType)
			continue
		}
		if c.reservedKeywords[word] {
			tokens[i] = promoteToken(tok, TokenReservedKeyword)
			continue
		}
	}
	return tokens
}

func promoteToken(tok Token, tokenType TokenType) Token {
	tok.Type = tokenType
	tok.Text = toCanonical(tok.Raw)
	return tok
}

func mergePhrases(tokens []Token, index phraseIndex) []Token {
	if len(index) == 0 {
		return tokens
	}
	out := make([]Token, 0, len(tokens))
	for i := 0; i < len(tokens); {
		tok := tokens[i]
		if tok.Type != TokenIdentifier {
			out = append(out, tok)
			i++
			continue
		}
		entries := index[asciiUpper(tok.Text)]
		matched := false
		for _, entry := range entries {
			if matchPhrase(tokens, i, entry.words) {
				out = append(out, mergeTokens(tokens, i, entry))
				i += len(entry.words)
				matched = true
				break
			}
		}
		if !matched {
			out = append(out, tok)
			i++
		}
	}
	return out
}

func matchPhrase(tokens []Token, start int, words []string) bool {
	if start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		tok := tokens[start+i]
		if tok.Type != TokenIdentifier {
			return false
		}
		if !asciiEqualUpper(tok.Text, word) {
			return false
		}
	}
	return true
}

func mergeTokens(tokens []Token, start int, entry phraseEntry) Token {
	var b strings.Builder
	rawLen := 0
	for i := range entry.words {
		rawLen += len(tokens[start+i].Raw)
	}
	rawLen += len(entry.words) - 1
	if rawLen > 0 {
		b.Grow(rawLen)
	}
	for i := range entry.words {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(tokens[start+i].Raw)
	}
	return Token{
		Type:                entry.tokenType,
		Raw:                 b.String(),
		Text:                entry.text,
		Start:               tokens[start].Start,
		PrecedingWhitespace: tokens[start].PrecedingWhitespace,
	}
}

func asciiUpper(s string) string {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			buf := make([]byte, len(s))
			copy(buf, s)
			for j := i; j < len(buf); j++ {
				c = buf[j]
				if c >= 'a' && c <= 'z' {
					buf[j] = c - ('a' - 'A')
				}
			}
			return string(buf)
		}
	}
	return s
}

func asciiEqualUpper(s string, upper string) bool {
	if len(s) != len(upper) {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		if c != upper[i] {
			return false
		}
	}
	return true
}

func multiWordItems(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if len(strings.Fields(item)) > 1 {
			out = append(out, item)
		}
	}
	return out
}
