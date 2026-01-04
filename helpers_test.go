package sqlformatter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sql-formatter-go/internal/testutil"
)

func formatPostgres(t *testing.T, query string, opts ...FormatOptions) string {
	t.Helper()
	cfg := FormatOptionsWithLanguage{Language: LanguagePostgresql}
	if len(opts) > 0 {
		cfg.FormatOptions = opts[0]
	}
	out, err := Format(query, cfg)
	require.NoError(t, err)
	return out
}

func formatPostgresErr(t *testing.T, query string, opts ...FormatOptions) error {
	t.Helper()
	cfg := FormatOptionsWithLanguage{Language: LanguagePostgresql}
	if len(opts) > 0 {
		cfg.FormatOptions = opts[0]
	}
	_, err := Format(query, cfg)
	return err
}

func dedent(text string) string {
	return testutil.Dedent(text)
}

func assertEqual(t *testing.T, got, expected string) {
	t.Helper()
	require.Equal(t, expected, got)
}
