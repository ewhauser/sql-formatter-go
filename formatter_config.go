package sqlformatter

import "strings"

func indentString(cfg FormatOptions) string {
	if cfg.IndentStyle == IndentStyleTabularLeft || cfg.IndentStyle == IndentStyleTabularRight {
		return strings.Repeat(" ", 10)
	}
	if cfg.UseTabs {
		return "\t"
	}
	return strings.Repeat(" ", cfg.TabWidth)
}

func isTabularStyle(cfg FormatOptions) bool {
	return cfg.IndentStyle == IndentStyleTabularLeft || cfg.IndentStyle == IndentStyleTabularRight
}
