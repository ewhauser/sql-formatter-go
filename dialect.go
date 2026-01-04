package sqlformatter

import (
	"fmt"
	"sync"
)

type DialectFormatOptions struct {
	AlwaysDenseOperators  []string
	OnelineClauses        []string
	TabularOnelineClauses []string
}

type ProcessedDialectFormatOptions struct {
	AlwaysDenseOperators  []string
	OnelineClauses        map[string]bool
	TabularOnelineClauses map[string]bool
}

type DialectOptions struct {
	Name             string
	TokenizerOptions TokenizerOptions
	FormatOptions    DialectFormatOptions
}

type Dialect struct {
	Tokenizer     *Tokenizer
	FormatOptions ProcessedDialectFormatOptions
}

var dialectCache sync.Map

func dialectCacheKey(options DialectOptions) string {
	if options.Name != "" {
		return options.Name
	}
	// Fallback to pointer identity for unnamed/custom dialects.
	return fmt.Sprintf("custom:%p", &options)
}

func CreateDialect(options DialectOptions) *Dialect {
	key := dialectCacheKey(options)
	if cached, ok := dialectCache.Load(key); ok {
		return cached.(*Dialect)
	}
	dialect := &Dialect{
		Tokenizer:     NewTokenizer(options.TokenizerOptions, options.Name),
		FormatOptions: processDialectFormatOptions(options.FormatOptions),
	}
	dialectCache.Store(key, dialect)
	return dialect
}

func processDialectFormatOptions(options DialectFormatOptions) ProcessedDialectFormatOptions {
	oneline := make(map[string]bool, len(options.OnelineClauses))
	for _, name := range options.OnelineClauses {
		oneline[name] = true
	}
	// if tabularOnelineClauses not provided, use oneline
	tabular := make(map[string]bool)
	if len(options.TabularOnelineClauses) == 0 {
		for k, v := range oneline {
			tabular[k] = v
		}
	} else {
		for _, name := range options.TabularOnelineClauses {
			tabular[name] = true
		}
	}
	return ProcessedDialectFormatOptions{
		AlwaysDenseOperators:  options.AlwaysDenseOperators,
		OnelineClauses:        oneline,
		TabularOnelineClauses: tabular,
	}
}
