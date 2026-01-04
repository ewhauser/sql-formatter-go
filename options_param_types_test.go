package sqlformatter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func supportsParamTypes(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("when paramTypes.positional=true", func(t *testing.T) {
		result := format("SELECT ?, ?, ?;", FormatOptions{ParamTypes: &ParamTypes{Positional: true}, Params: []string{"first", "second", "third"}})
		expected := dedent(`
			SELECT
			  first,
			  second,
			  third;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("when paramTypes.named=[\":\"]", func(t *testing.T) {
		result := format("SELECT :a, :b, :c;", FormatOptions{ParamTypes: &ParamTypes{Named: []string{":"}}, Params: ParamItems{"a": "first", "b": "second", "c": "third"}})
		expected := dedent(`
			SELECT
			  first,
			  second,
			  third;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("when paramTypes.numbered=[\"?\"]", func(t *testing.T) {
		result := format("SELECT ?1, ?2, ?3;", FormatOptions{ParamTypes: &ParamTypes{Numbered: []string{"?"}}, Params: ParamItems{"1": "first", "2": "second", "3": "third"}})
		expected := dedent(`
			SELECT
			  first,
			  second,
			  third;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("when paramTypes.custom provided", func(t *testing.T) {
		result := format("SELECT %1%, %2%, %3%;", FormatOptions{ParamTypes: &ParamTypes{Custom: []CustomParameter{{Regex: "%[0-9]+%"}}}, Params: ParamItems{"%1%": "first", "%2%": "second", "%3%": "third"}})
		expected := dedent(`
			SELECT
			  first,
			  second,
			  third;
		`)
		assertEqual(t, result, expected)

		result = format("SELECT %1%, %2%, %3%;", FormatOptions{ParamTypes: &ParamTypes{Custom: []CustomParameter{{Regex: "%[0-9]+%", Key: func(v string) string { return v[1 : len(v)-1] }}}}, Params: ParamItems{"1": "first", "2": "second", "3": "third"}})
		expected = dedent(`
			SELECT
			  first,
			  second,
			  third;
		`)
		assertEqual(t, result, expected)

		result = format("SELECT %1%, {2};", FormatOptions{ParamTypes: &ParamTypes{Custom: []CustomParameter{{Regex: "%[0-9]+%"}, {Regex: `\{[0-9]\}`}}}, Params: ParamItems{"%1%": "first", "{2}": "second"}})
		expected = dedent(`
			SELECT
			  first,
			  second;
		`)
		assertEqual(t, result, expected)

		result = format("SELECT {schema}.{table}.{column} FROM {schema}.{table}", FormatOptions{ParamTypes: &ParamTypes{Custom: []CustomParameter{{Regex: `\{\w+\}`}}}})
		expected = dedent(`
			SELECT
			  {schema}.{table}.{column}
			FROM
			  {schema}.{table}
		`)
		assertEqual(t, result, expected)

		err := formatPostgresErr(t, "SELECT foo FROM bar", FormatOptions{ParamTypes: &ParamTypes{Custom: []CustomParameter{{Regex: ""}}}})
		require.Error(t, err)
		require.Equal(t, "Empty regex given in custom paramTypes. That would result in matching infinite amount of parameters.", err.Error())
	})
}
