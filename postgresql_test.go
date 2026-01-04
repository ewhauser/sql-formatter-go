package sqlformatter

import "testing"

func TestPostgresqlFormatter(t *testing.T) {
	format := func(query string, cfg ...FormatOptions) string {
		if len(cfg) > 0 {
			return formatPostgres(t, query, cfg[0])
		}
		return formatPostgres(t, query)
	}
	formatErr := func(query string, cfg ...FormatOptions) error {
		if len(cfg) > 0 {
			return formatPostgresErr(t, query, cfg[0])
		}
		return formatPostgresErr(t, query)
	}

	behavesLikePostgresqlFormatter(t, format)
	supportsCreateView(t, format, createViewConfig{OrReplace: true, Materialized: true, IfNotExists: true})
	supportsCreateTable(t, format, createTableConfig{IfNotExists: true})
	supportsDropTable(t, format, dropTableConfig{IfExists: true})
	supportsConstraints(t, format, []string{"NO ACTION", "RESTRICT", "CASCADE", "SET NULL", "SET DEFAULT"})
	supportsArrayLiterals(t, format, arrayLiteralConfig{WithArrayPrefix: true})
	supportsUpdate(t, format, updateConfig{WhereCurrentOf: true})
	supportsTruncateTable(t, format, truncateTableConfig{WithoutTable: true})
	supportsStrings(t, format, formatErr, []string{"''-qq", "U&''", "X''", "B''", "E''", "$$"})
	supportsIdentifiers(t, format, formatErr, []string{"\"\"-qq", "U&\"\""})
	supportsSchema(t, format)
	supportsOperators(t, format, []string{
		"%", "^", "|/", "||/", "@",
		":=",
		"&", "|", "#", "~", "<<", ">>",
		"~>~", "~<~", "~>=~", "~<=~",
		"@-@", "@@", "##", "<->", "&&", "&<", "&>", "<<|", "&<|", "|>>", "|&>", "<^", "^>", "?#", "?-", "?|", "?-|", "?||", "@>", "<@", "~=",
		"?", "@?", "?&", "->", "->>", "#>", "#>>", "#-",
		"=>",
		">>=", "<<=",
		"~~", "~~*", "!~~", "!~~*",
		"~", "~*", "!~", "!~*",
		"-|-",
		"||",
		"@@@", "!!", "^@",
		"<%", "<<%", "%>", "%>>", "<<->", "<->>", "<<<->", "<->>>",
		"<#>", "<=>", "<+>", "<~>", "<%>",
	}, operatorConfig{Any: true})
	supportsIsDistinctFrom(t, format)
	supportsJoin(t, format)
	supportsSetOperations(t, format)
	supportsParams(t, format, paramConfig{Numbered: []string{"$"}})
	supportsLimiting(t, format, limitingConfig{Limit: true, Offset: true, FetchFirst: true, FetchNext: true})
	supportsDataTypeCase(t, format)

	// Regression test for issue #624
	t.Run("supports array slice operator", func(t *testing.T) {
		result := format("SELECT foo[:5], bar[1:], baz[1:5], zap[:];")
		expected := dedent(`
			SELECT
			  foo[:5],
			  bar[1:],
			  baz[1:5],
			  zap[:];
		`)
		assertEqual(t, result, expected)
	})

	// Regression test for issue #447
	t.Run("formats empty SELECT", func(t *testing.T) {
		result := format("SELECT;")
		expected := dedent(`
			SELECT;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats TIMESTAMP WITH TIME ZONE syntax", func(t *testing.T) {
		result := format("create table time_table (id int,\n          created_at timestamp without time zone,\n          deleted_at time with time zone,\n          modified_at timestamp(0) with time zone);", FormatOptions{DataTypeCase: KeywordCaseUpper})
		expected := dedent(`
			create table time_table (
			  id INT,
			  created_at TIMESTAMP WITHOUT TIME ZONE,
			  deleted_at TIME WITH TIME ZONE,
			  modified_at TIMESTAMP(0) WITH TIME ZONE
			);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats FOR UPDATE clause", func(t *testing.T) {
		result := format(`
        SELECT * FROM tbl FOR UPDATE;
        SELECT * FROM tbl FOR UPDATE OF tbl.salary;
      `)
		expected := dedent(`
			SELECT
			  *
			FROM
			  tbl
			FOR UPDATE;

			SELECT
			  *
			FROM
			  tbl
			FOR UPDATE OF
			  tbl.salary;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports OPERATOR() syntax", func(t *testing.T) {
		result := format("SELECT foo OPERATOR(public.===) bar;")
		expected := dedent(`
			SELECT
			  foo OPERATOR(public.===) bar;
		`)
		assertEqual(t, result, expected)
		result = format("SELECT foo operator ( !== ) bar;")
		expected = dedent(`
			SELECT
			  foo operator ( !== ) bar;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("supports OR REPLACE in CREATE FUNCTION", func(t *testing.T) {
		result := format("CREATE OR REPLACE FUNCTION foo ();")
		expected := dedent(`
			CREATE OR REPLACE FUNCTION foo ();
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats JSON and JSONB data types", func(t *testing.T) {
		result := format("CREATE TABLE foo (bar json, baz jsonb);", FormatOptions{DataTypeCase: KeywordCaseUpper})
		expected := "CREATE TABLE foo (bar JSON, baz JSONB);"
		assertEqual(t, result, expected)
	})

	t.Run("supports OR REPLACE in CREATE PROCEDURE", func(t *testing.T) {
		result := format("CREATE OR REPLACE PROCEDURE foo () LANGUAGE sql AS $$ BEGIN END $$;")
		expected := dedent(`CREATE OR REPLACE PROCEDURE foo () LANGUAGE sql AS $$ BEGIN END $$;`)
		assertEqual(t, result, expected)
	})

	t.Run("supports UUID type and functions", func(t *testing.T) {
		result := format("CREATE TABLE foo (id uuid DEFAULT Gen_Random_Uuid());", FormatOptions{DataTypeCase: KeywordCaseUpper, FunctionCase: KeywordCaseLower})
		expected := dedent(`CREATE TABLE foo (id UUID DEFAULT gen_random_uuid());`)
		assertEqual(t, result, expected)
	})

	t.Run("formats keywords in COMMENT ON", func(t *testing.T) {
		result := format("comment on table foo is 'Hello my table';", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`COMMENT ON TABLE foo IS 'Hello my table';`)
		assertEqual(t, result, expected)
	})
}
