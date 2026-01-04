package sqlformatter

import (
	"fmt"
)

type SqlLanguage string

const (
	LanguagePostgresql SqlLanguage = "postgresql"
)

var dialectNameMap = map[SqlLanguage]DialectOptions{
	LanguagePostgresql: PostgresqlDialect,
}

var supportedDialects = []string{"postgresql"}

var defaultOptions = FormatOptions{
	TabWidth:               2,
	UseTabs:                false,
	KeywordCase:            KeywordCasePreserve,
	IdentifierCase:         KeywordCasePreserve,
	DataTypeCase:           KeywordCasePreserve,
	FunctionCase:           KeywordCasePreserve,
	IndentStyle:            IndentStyleStandard,
	LogicalOperatorNewline: LogicalOperatorNewlineBefore,
	ExpressionWidth:        50,
	LinesBetweenQueries:    1,
	DenseOperators:         false,
	NewlineBeforeSemicolon: false,
}

func Format(query string, cfg FormatOptionsWithLanguage) (string, error) {
	if cfg.Language != "" {
		if _, ok := dialectNameMap[cfg.Language]; !ok {
			return "", ConfigError{Message: fmt.Sprintf("Unsupported SQL dialect: %s", cfg.Language)}
		}
	} else {
		cfg.Language = LanguagePostgresql
	}
	options := cfg.FormatOptions
	options = mergeOptions(defaultOptions, options)
	validated, err := validateConfig(options)
	if err != nil {
		return "", err
	}
	dialect := CreateDialect(dialectNameMap[cfg.Language])
	formatter := NewFormatter(dialect, validated)
	return formatter.Format(query)
}

func FormatDialect(query string, cfg FormatOptionsWithDialect) (string, error) {
	options := cfg.FormatOptions
	options = mergeOptions(defaultOptions, options)
	validated, err := validateConfig(options)
	if err != nil {
		return "", err
	}
	dialect := CreateDialect(cfg.Dialect)
	formatter := NewFormatter(dialect, validated)
	return formatter.Format(query)
}

func mergeOptions(base FormatOptions, override FormatOptions) FormatOptions {
	if override.TabWidth != 0 {
		base.TabWidth = override.TabWidth
	}
	if override.UseTabs {
		base.UseTabs = true
	}
	if override.KeywordCase != "" {
		base.KeywordCase = override.KeywordCase
	}
	if override.IdentifierCase != "" {
		base.IdentifierCase = override.IdentifierCase
	}
	if override.DataTypeCase != "" {
		base.DataTypeCase = override.DataTypeCase
	}
	if override.FunctionCase != "" {
		base.FunctionCase = override.FunctionCase
	}
	if override.IndentStyle != "" {
		base.IndentStyle = override.IndentStyle
	}
	if override.LogicalOperatorNewline != "" {
		base.LogicalOperatorNewline = override.LogicalOperatorNewline
	}
	if override.ExpressionWidthSet {
		base.ExpressionWidth = override.ExpressionWidth
	}
	if override.LinesBetweenQueriesSet {
		base.LinesBetweenQueries = override.LinesBetweenQueries
	}
	if override.DenseOperators {
		base.DenseOperators = true
	}
	if override.NewlineBeforeSemicolon {
		base.NewlineBeforeSemicolon = true
	}
	if override.Params != nil {
		base.Params = override.Params
	}
	if override.ParamTypes != nil {
		base.ParamTypes = override.ParamTypes
	}
	return base
}
