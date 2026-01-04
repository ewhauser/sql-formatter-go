package sqlformatter

type FormatFn func(query string, cfg ...FormatOptions) string

type FormatErrFn func(query string, cfg ...FormatOptions) error
