package sqlformatter

import (
	"regexp"
	"strings"
)

type Tokenizer struct {
	cfg         TokenizerOptions
	dialectName string
	rulesBefore []TokenRule
	rulesAfter  []TokenRule
	classifier  *tokenClassifier
}

func NewTokenizer(cfg TokenizerOptions, dialectName string) *Tokenizer {
	t := &Tokenizer{cfg: cfg, dialectName: dialectName, classifier: newTokenClassifier(cfg)}
	t.rulesBefore = t.buildRulesBeforeParams(cfg)
	t.rulesAfter = t.buildRulesAfterParams(cfg)
	return t
}

func (t *Tokenizer) Tokenize(input string, paramTypesOverrides *ParamTypes) ([]Token, error) {
	rules := make([]TokenRule, 0, len(t.rulesBefore)+len(t.rulesAfter)+5)
	rules = append(rules, t.rulesBefore...)
	rules = append(rules, t.buildParamRules(t.cfg, paramTypesOverrides)...)
	rules = append(rules, t.rulesAfter...)
	engine := NewTokenizerEngine(rules, t.dialectName)
	tokens, err := engine.Tokenize(input)
	if err != nil {
		return nil, err
	}
	if t.classifier != nil {
		tokens = t.classifier.Classify(tokens)
	}
	if t.cfg.PostProcess != nil {
		return t.cfg.PostProcess(tokens), nil
	}
	return tokens, nil
}

func (t *Tokenizer) buildRulesBeforeParams(cfg TokenizerOptions) []TokenRule {
	lineComments := cfg.LineCommentTypes
	if len(lineComments) == 0 {
		lineComments = []string{"--"}
	}

	rules := []TokenRule{}
	// disable comment
	disableRe := regexp.MustCompile(`(?s)^/\* *sql-formatter-disable *\*/.*?(?:/\* *sql-formatter-enable *\*/|$)`) // dot matches newline
	rules = append(rules, TokenRule{Type: TokenDisableComment, Regex: &RegexMatcher{re: newRegexpWrapper(disableRe)}})

	// block comment
	if cfg.NestedBlockComments {
		rules = append(rules, TokenRule{Type: TokenBlockComment, Regex: NestedCommentMatcher{}})
	} else {
		re := regexp.MustCompile(`(?s)^/\*.*?\*/`)
		rules = append(rules, TokenRule{Type: TokenBlockComment, Regex: &RegexMatcher{re: newRegexpWrapper(re)}})
	}

	rules = append(rules, TokenRule{Type: TokenLineComment, Regex: NewLineCommentMatcher(lineComments)})

	rules = append(rules, TokenRule{Type: TokenQuotedIdentifier, Regex: NewQuoteMatcher(cfg.IdentTypes)})

	rules = append(rules, TokenRule{Type: TokenNumber, Regex: &NumberMatcher{AllowUnderscore: cfg.UnderscoresInNumbers}})
	if cfg.OperatorKeyword {
		re := regexp.MustCompile(`(?i)^OPERATOR *\([^)]+\)`) // case-insensitive
		rules = append(rules, TokenRule{Type: TokenOperator, Regex: &RegexMatcher{re: newRegexpWrapper(re)}})
	}

	return rules
}

func (t *Tokenizer) buildRulesAfterParams(cfg TokenizerOptions) []TokenRule {
	rules := []TokenRule{}
	if len(cfg.VariableTypes) > 0 {
		quoteTypes := make([]QuoteType, len(cfg.VariableTypes))
		for i, v := range cfg.VariableTypes {
			quoteTypes[i] = v
		}
		rules = append(rules, TokenRule{Type: TokenVariable, Regex: NewQuoteMatcher(quoteTypes)})
	}
	rules = append(rules, TokenRule{Type: TokenString, Regex: NewQuoteMatcher(cfg.StringTypes)})
	rules = append(rules, TokenRule{Type: TokenIdentifier, Regex: NewIdentifierMatcher(cfg.IdentChars)})
	rules = append(rules, TokenRule{Type: TokenDelimiter, Regex: &RegexMatcher{re: newRegexpWrapper(regexp.MustCompile(`^;`))}})
	rules = append(rules, TokenRule{Type: TokenComma, Regex: &RegexMatcher{re: newRegexpWrapper(regexp.MustCompile(`^,`))}})
	rules = append(rules, TokenRule{Type: TokenOpenParen, Regex: NewParenMatcher(true, cfg.ExtraParens)})
	rules = append(rules, TokenRule{Type: TokenCloseParen, Regex: NewParenMatcher(false, cfg.ExtraParens)})
	ops := []string{"+", "-", "/", ">", "<", "=", "<>", "<=", ">=", "!="}
	if len(cfg.Operators) > 0 {
		ops = append(ops, cfg.Operators...)
	}
	rules = append(rules, TokenRule{Type: TokenOperator, Regex: NewOperatorMatcher(ops)})
	rules = append(rules, TokenRule{Type: TokenAsterisk, Regex: &RegexMatcher{re: newRegexpWrapper(regexp.MustCompile(`^\*`))}})
	propertyOps := []string{"."}
	if len(cfg.PropertyAccessOperators) > 0 {
		propertyOps = append(propertyOps, cfg.PropertyAccessOperators...)
	}
	rules = append(rules, TokenRule{Type: TokenPropertyAccessOperator, Regex: NewOperatorMatcher(propertyOps)})
	return rules
}

func (t *Tokenizer) buildParamRules(cfg TokenizerOptions, overrides *ParamTypes) []TokenRule {
	paramTypes := mergeParamTypes(cfg.ParamTypes, overrides)
	if paramTypes == nil {
		return nil
	}

	paramChars := cfg.ParamChars
	if paramChars == nil {
		paramChars = cfg.IdentChars
	}

	rules := []TokenRule{}

	if len(paramTypes.Named) > 0 {
		matcher := NewParameterMatcher(paramTypes.Named, paramChars, cfg.IdentTypes)
		rules = append(rules, TokenRule{Type: TokenNamedParameter, Regex: matcher, Key: func(raw string) string { return raw[1:] }})
	}
	if len(paramTypes.Quoted) > 0 {
		matcher := NewQuotedParameterMatcher(paramTypes.Quoted, cfg.IdentTypes)
		rules = append(rules, TokenRule{Type: TokenQuotedParameter, Regex: matcher, Key: func(raw string) string {
			if len(raw) < 2 {
				return ""
			}
			// remove prefix and quotes
			return raw[2 : len(raw)-1]
		}})
	}
	if len(paramTypes.Numbered) > 0 {
		matcher := NewNumberedParameterMatcher(paramTypes.Numbered)
		rules = append(rules, TokenRule{Type: TokenNumberedParameter, Regex: matcher, Key: func(raw string) string { return raw[1:] }})
	}
	if paramTypes.Positional {
		re := regexp.MustCompile(`^\?`)
		rules = append(rules, TokenRule{Type: TokenPositionalParameter, Regex: &RegexMatcher{re: newRegexpWrapper(re)}})
	}
	for _, custom := range paramTypes.Custom {
		pattern := custom.Regex
		if pattern == "" {
			continue
		}
		re := PatternToRegex(pattern, false)
		keyFn := custom.Key
		if keyFn == nil {
			keyFn = func(v string) string { return v }
		}
		rules = append(rules, TokenRule{Type: TokenCustomParameter, Regex: &RegexMatcher{re: newRegexpWrapper(re)}, Key: keyFn})
	}
	return rules
}

func mergeParamTypes(defaults *ParamTypes, overrides *ParamTypes) *ParamTypes {
	if defaults == nil && overrides == nil {
		return nil
	}
	out := &ParamTypes{}
	if defaults != nil {
		*out = *defaults
	}
	if overrides == nil {
		return out
	}
	if overrides.Named != nil {
		out.Named = overrides.Named
	}
	if overrides.Quoted != nil {
		out.Quoted = overrides.Quoted
	}
	if overrides.Numbered != nil {
		out.Numbered = overrides.Numbered
	}
	if overrides.Positional {
		out.Positional = true
	}
	if overrides.Custom != nil {
		out.Custom = overrides.Custom
	}
	return out
}

func toCanonical(v string) string {
	return EqualizeWhitespace(strings.ToUpper(v))
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
