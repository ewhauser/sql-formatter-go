package sqlformatter

import "strings"

type Formatter struct {
	dialect *Dialect
	cfg     FormatOptions
	params  *Params
}

func NewFormatter(dialect *Dialect, cfg FormatOptions) *Formatter {
	return &Formatter{dialect: dialect, cfg: cfg, params: NewParams(cfg.Params)}
}

func (f *Formatter) Format(query string) (string, error) {
	parser := NewParser(f.dialect.Tokenizer)
	ast, err := parser.Parse(query, f.dialect.Tokenizer, f.cfg.ParamTypes)
	if err != nil {
		return "", err
	}
	formatted := f.formatAst(ast)
	return strings.TrimRight(formatted, " \t\n\r"), nil
}

func (f *Formatter) formatAst(statements []*StatementNode) string {
	parts := make([]string, 0, len(statements))
	for _, stmt := range statements {
		parts = append(parts, f.formatStatement(stmt))
	}
	if len(parts) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		if startsWithLineComment(parts[i]) {
			out.WriteString("\n")
		} else {
			out.WriteString(strings.Repeat("\n", f.cfg.LinesBetweenQueries+1))
		}
		out.WriteString(parts[i])
	}
	return out.String()
}

func startsWithLineComment(formatted string) bool {
	return strings.HasPrefix(strings.TrimLeft(formatted, " \t"), "--")
}

func (f *Formatter) formatStatement(statement *StatementNode) string {
	layout := NewExpressionFormatter(ExpressionFormatterParams{
		Cfg:        f.cfg,
		DialectCfg: f.dialect.FormatOptions,
		Params:     f.params,
		Layout:     NewLayout(NewIndentation(indentString(f.cfg))),
	}).Format(statement.Children)

	if statement.HasSemicolon {
		if f.cfg.NewlineBeforeSemicolon {
			layout.Add(Newline, ";")
		} else {
			layout.Add(NoNewline, ";")
		}
	}
	return layout.ToString()
}
