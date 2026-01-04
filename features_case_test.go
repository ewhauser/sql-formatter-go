package sqlformatter

import "testing"

func supportsCase(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats CASE WHEN with blank expression", func(t *testing.T) {
		result := format("CASE WHEN opt = 'foo' THEN 1 WHEN opt = 'bar' THEN 2 WHEN opt = 'baz' THEN 3 ELSE 4 END;")
		expected := dedent(`
			CASE
			  WHEN opt = 'foo' THEN 1
			  WHEN opt = 'bar' THEN 2
			  WHEN opt = 'baz' THEN 3
			  ELSE 4
			END;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CASE WHEN with expression", func(t *testing.T) {
		result := format("CASE trim(sqrt(2)) WHEN 'one' THEN 1 WHEN 'two' THEN 2 WHEN 'three' THEN 3 ELSE 4 END;")
		expected := dedent(`
			CASE trim(sqrt(2))
			  WHEN 'one' THEN 1
			  WHEN 'two' THEN 2
			  WHEN 'three' THEN 3
			  ELSE 4
			END;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CASE WHEN inside SELECT", func(t *testing.T) {
		result := format("SELECT foo, bar, CASE baz WHEN 'one' THEN 1 WHEN 'two' THEN 2 ELSE 3 END FROM tbl;")
		expected := dedent(`
			SELECT
			  foo,
			  bar,
			  CASE baz
			    WHEN 'one' THEN 1
			    WHEN 'two' THEN 2
			    ELSE 3
			  END
			FROM
			  tbl;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("recognizes lowercase CASE END", func(t *testing.T) {
		result := format("case when opt = 'foo' then 1 else 2 end;")
		expected := dedent(`
			case
			  when opt = 'foo' then 1
			  else 2
			end;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("ignores words CASE and END inside other strings", func(t *testing.T) {
		result := format("SELECT CASEDATE, ENDDATE FROM table1;")
		expected := dedent(`
			SELECT
			  CASEDATE,
			  ENDDATE
			FROM
			  table1;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("properly converts to uppercase in case statements", func(t *testing.T) {
		result := format("case trim(sqrt(my_field)) when 'one' then 1 when 'two' then 2 when 'three' then 3 else 4 end;", FormatOptions{KeywordCase: KeywordCaseUpper, FunctionCase: KeywordCaseUpper})
		expected := dedent(`
			CASE TRIM(SQRT(my_field))
			  WHEN 'one' THEN 1
			  WHEN 'two' THEN 2
			  WHEN 'three' THEN 3
			  ELSE 4
			END;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("handles edge case ending inline block with END", func(t *testing.T) {
		result := format(dedent("select sum(case a when foo then bar end) from quaz"))
		expected := dedent(`
			select
			  sum(
			    case a
			      when foo then bar
			    end
			  )
			from
			  quaz
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CASE with comments", func(t *testing.T) {
		result := format(`
      SELECT CASE /*c1*/ foo /*c2*/
      WHEN /*c3*/ 1 /*c4*/ THEN /*c5*/ 2 /*c6*/
      ELSE /*c7*/ 3 /*c8*/
      END;
    `)
		expected := dedent(`
			SELECT
			  CASE /*c1*/ foo /*c2*/
			    WHEN /*c3*/ 1 /*c4*/ THEN /*c5*/ 2 /*c6*/
			    ELSE /*c7*/ 3 /*c8*/
			  END;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CASE with comments inside sub-expressions", func(t *testing.T) {
		result := format(`
      SELECT CASE foo + /*c1*/ bar
      WHEN 1 /*c2*/ + 1 THEN 2 /*c2*/ * 2
      ELSE 3 - /*c3*/ 3
      END;
    `)
		expected := dedent(`
			SELECT
			  CASE foo + /*c1*/ bar
			    WHEN 1 /*c2*/ + 1 THEN 2 /*c2*/ * 2
			    ELSE 3 - /*c3*/ 3
			  END;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CASE with indentStyle tabularLeft", func(t *testing.T) {
		result := format("SELECT CASE foo WHEN 1 THEN bar ELSE baz END;", FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			SELECT    CASE foo
			                    WHEN 1 THEN bar
			                    ELSE baz
			          END;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats CASE with indentStyle tabularRight", func(t *testing.T) {
		result := format("SELECT CASE foo WHEN 1 THEN bar ELSE baz END;", FormatOptions{IndentStyle: IndentStyleTabularRight})
		expected := "   SELECT CASE foo\n                    WHEN 1 THEN bar\n                    ELSE baz\n          END;"
		assertEqual(t, result, expected)
	})

	t.Run("formats nested case expressions", func(t *testing.T) {
		result := format(`
      SELECT
        CASE
          CASE foo WHEN 1 THEN 11 ELSE 22 END
          WHEN 11 THEN 110
          WHEN 22 THEN 220
          ELSE 123
        END
      FROM
        tbl;
    `)
		expected := dedent(`
			SELECT
			  CASE CASE foo
			      WHEN 1 THEN 11
			      ELSE 22
			    END
			    WHEN 11 THEN 110
			    WHEN 22 THEN 220
			    ELSE 123
			  END
			FROM
			  tbl;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats between inside case expression", func(t *testing.T) {
		result := format(`
    SELECT CASE WHEN x1 BETWEEN 1 AND 12 THEN '' END c1;
  `)
		expected := dedent(`
    SELECT
      CASE
        WHEN x1 BETWEEN 1 AND 12  THEN ''
      END c1;
  `)
		assertEqual(t, result, expected)
	})
}
