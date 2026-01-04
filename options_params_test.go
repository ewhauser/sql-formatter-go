package sqlformatter

import "testing"

type paramConfig struct {
	Positional bool
	Numbered   []string
	Named      []string
	Quoted     []string
}

func supportsParams(t *testing.T, format FormatFn, params paramConfig) {
	t.Helper()
	t.Run("supports params", func(t *testing.T) {
		if params.Positional {
			t.Run("leaves ? positional placeholders when no params provided", func(t *testing.T) {
				result := format("SELECT ?, ?, ?;")
				expected := dedent(`
					SELECT
					  ?,
					  ?,
					  ?;
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces ? positional placeholders with params", func(t *testing.T) {
				result := format("SELECT ?, ?, ?;", FormatOptions{Params: []string{"first", "second", "third"}})
				expected := dedent(`
					SELECT
					  first,
					  second,
					  third;
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces ? positional placeholders inside BETWEEN", func(t *testing.T) {
				result := format("SELECT name WHERE age BETWEEN ? AND ?;", FormatOptions{Params: []string{"5", "10"}})
				expected := dedent(`
					SELECT
					  name
					WHERE
					  age BETWEEN 5 AND 10;
				`)
				assertEqual(t, result, expected)
			})
		}

		if containsString(params.Numbered, "?") {
			t.Run("recognizes ? numbered placeholders", func(t *testing.T) {
				result := format("SELECT ?1, ?25, ?2;")
				expected := dedent(`
					SELECT
					  ?1,
					  ?25,
					  ?2;
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces ? numbered placeholders with params", func(t *testing.T) {
				result := format("SELECT ?1, ?2, ?0;", FormatOptions{Params: ParamItems{"0": "first", "1": "second", "2": "third"}})
				expected := dedent(`
					SELECT
					  second,
					  third,
					  first;
				`)
				assertEqual(t, result, expected)
			})
		}

		if containsString(params.Numbered, "$") {
			t.Run("recognizes $n placeholders", func(t *testing.T) {
				result := format("SELECT $1, $2 FROM tbl")
				expected := dedent(`
					SELECT
					  $1,
					  $2
					FROM
					  tbl
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces $n placeholders with params", func(t *testing.T) {
				result := format("SELECT $1, $2 FROM tbl", FormatOptions{Params: ParamItems{"1": "\"variable value\"", "2": "\"blah\""}})
				expected := dedent(`
					SELECT
					  "variable value",
					  "blah"
					FROM
					  tbl
				`)
				assertEqual(t, result, expected)
			})
		}

		if containsString(params.Numbered, ":") {
			t.Run("recognizes :n placeholders", func(t *testing.T) {
				result := format("SELECT :1, :2 FROM tbl")
				expected := dedent(`
					SELECT
					  :1,
					  :2
					FROM
					  tbl
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces :n placeholders with params", func(t *testing.T) {
				result := format("SELECT :1, :2 FROM tbl", FormatOptions{Params: ParamItems{"1": "\"variable value\"", "2": "\"blah\""}})
				expected := dedent(`
					SELECT
					  "variable value",
					  "blah"
					FROM
					  tbl
				`)
				assertEqual(t, result, expected)
			})
		}

		if containsString(params.Named, ":") {
			t.Run("recognizes :name placeholders", func(t *testing.T) {
				result := format("SELECT :foo, :bar, :baz;")
				expected := dedent(`
					SELECT
					  :foo,
					  :bar,
					  :baz;
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces :name placeholders with params", func(t *testing.T) {
				result := format("WHERE name = :name AND age > :current_age;", FormatOptions{Params: ParamItems{"name": "'John'", "current_age": "10"}})
				expected := dedent(`
					WHERE
					  name = 'John'
					  AND age > 10;
				`)
				assertEqual(t, result, expected)
			})
		}

		if containsString(params.Named, "$" ) {
			t.Run("recognizes $name placeholders", func(t *testing.T) {
				result := format("SELECT $foo, $bar, $baz;")
				expected := dedent(`
					SELECT
					  $foo,
					  $bar,
					  $baz;
				`)
				assertEqual(t, result, expected)
			})

			t.Run("replaces $name placeholders with params", func(t *testing.T) {
				result := format("WHERE name = $name AND age > $current_age;", FormatOptions{Params: ParamItems{"name": "'John'", "current_age": "10"}})
				expected := dedent(`
					WHERE
					  name = 'John'
					  AND age > 10;
				`)
				assertEqual(t, result, expected)
			})
		}
	})

	if containsString(params.Named, "@") {
		t.Run("recognizes @name placeholders", func(t *testing.T) {
			result := format("SELECT @foo, @bar, @baz;")
			expected := dedent(`
				SELECT
				  @foo,
				  @bar,
				  @baz;
			`)
			assertEqual(t, result, expected)
		})

		t.Run("replaces @name placeholders with params", func(t *testing.T) {
			result := format("WHERE name = @name AND age > @current_age;", FormatOptions{Params: ParamItems{"name": "'John'", "current_age": "10"}})
			expected := dedent(`
				WHERE
				  name = 'John'
				  AND age > 10;
			`)
			assertEqual(t, result, expected)
		})
	}
}
