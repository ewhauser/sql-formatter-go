package sqlformatter

import "testing"

type commentsConfig struct {
	HashComments       bool
	DoubleSlashComments bool
	NestedBlockComments bool
}

func supportsComments(t *testing.T, format FormatFn, opts commentsConfig) {
	t.Helper()
	t.Run("formats SELECT query with different comments", func(t *testing.T) {
		result := format(dedent(`
      SELECT
      /*
       * This is a block comment
       */
      * FROM
      -- This is another comment
      MyTable -- One final comment
      WHERE 1 = 2;
    `))
		expected := dedent(`
			SELECT
			  /*
			   * This is a block comment
			   */
			  *
			FROM
			  -- This is another comment
			  MyTable -- One final comment
			WHERE
			  1 = 2;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("maintains block comment indentation", func(t *testing.T) {
		sql := dedent(`
      SELECT
        /*
         * This is a block comment
         */
        *
      FROM
        MyTable
      WHERE
        1 = 2;
    `)
		result := format(sql)
		assertEqual(t, result, sql)
	})

	t.Run("keeps block comment on separate line", func(t *testing.T) {
		sql := dedent(`
      SELECT
        /* separate-line block comment */
        foo,
        bar /* inline block comment */
      FROM
        tbl;
    `)
		result := format(sql)
		assertEqual(t, result, sql)
	})

	t.Run("formats tricky line comments", func(t *testing.T) {
		result := format("SELECT a--comment, here\nFROM b--comment")
		expected := dedent(`
			SELECT
			  a --comment, here
			FROM
			  b --comment
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats line comments followed by semicolon", func(t *testing.T) {
		result := format(`
      SELECT a FROM b --comment
      ;
    `)
		expected := dedent(`
			SELECT
			  a
			FROM
			  b --comment
			;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats line comments followed by comma", func(t *testing.T) {
		result := format(dedent(`
      SELECT a --comment
      , b
    `))
		expected := dedent(`
			SELECT
			  a --comment
			,
			  b
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats line comments followed by close-paren", func(t *testing.T) {
		result := format("SELECT ( a --comment\n )")
		expected := dedent(`
			SELECT
			  (
			    a --comment
			  )
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats line comments followed by open-paren", func(t *testing.T) {
		result := format("SELECT a --comment\n()")
		expected := dedent(`
			SELECT
			  a --comment
			  ()
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats first line comment in file", func(t *testing.T) {
		result := format("-- comment1\n-- comment2\n")
		expected := dedent(`
			-- comment1
			-- comment2
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats first block comment in file", func(t *testing.T) {
		result := format("/*comment1*/\n/*comment2*/\n")
		expected := dedent(`
			/*comment1*/
			/*comment2*/
		`)
		assertEqual(t, result, expected)
	})

	t.Run("preserves single-line comments at end of lines", func(t *testing.T) {
		result := format(`
        SELECT
          a, --comment1
          b --comment2
        FROM --comment3
          my_table;
      `)
		expected := dedent(`
			SELECT
			  a, --comment1
			  b --comment2
			FROM --comment3
			  my_table;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("preserves single-line comments on separate lines", func(t *testing.T) {
		result := format(`
        SELECT
          --comment1
          a,
          --comment2
          b
        FROM
          --comment3
          my_table;
      `)
		expected := dedent(`
			SELECT
			  --comment1
			  a,
			  --comment2
			  b
			FROM
			  --comment3
			  my_table;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("recognizes line-comments with Windows line-endings", func(t *testing.T) {
		result := format("SELECT * FROM\r\n-- line comment 1\r\nMyTable -- line comment 2\r\n")
		expected := "SELECT\n  *\nFROM\n  -- line comment 1\n  MyTable -- line comment 2"
		assertEqual(t, result, expected)
	})

	t.Run("does not detect unclosed comment as comment", func(t *testing.T) {
		result := format(`
      SELECT count(*)
      /*SomeComment
    `)
		expected := dedent(`
			SELECT
			  count(*) / * SomeComment
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats comments between function name and parenthesis", func(t *testing.T) {
		result := format(`
      SELECT count /* comment */ (*);
    `)
		expected := dedent(`
			SELECT
			  count/* comment */ (*);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats comments between qualified names (before dot)", func(t *testing.T) {
		result := format(`
      SELECT foo/* com1 */.bar, count()/* com2 */.bar, foo.bar/* com3 */.baz, (1, 2) /* com4 */.foo;
    `)
		expected := dedent(`
			SELECT
			  foo /* com1 */.bar,
			  count() /* com2 */.bar,
			  foo.bar /* com3 */.baz,
			  (1, 2) /* com4 */.foo;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("indents multiline block comment that is not a doc-comment", func(t *testing.T) {
		result := format(dedent(`
      SELECT 1
      /*
      comment line
      */
    `))
		expected := dedent(`
			SELECT
			  1
			  /*
			  comment line
			  */
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats comments between qualified names (after dot)", func(t *testing.T) {
		result := format(`
      SELECT foo. /* com1 */ bar, foo. /* com2 */ *;
    `)
		expected := dedent(`
			SELECT
			  foo./* com1 */ bar,
			  foo./* com2 */ *;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("handles block comments with /** and **/ patterns", func(t *testing.T) {
		sql := "/** This is a block comment **/"
		result := format(sql)
		assertEqual(t, result, sql)
	})

	if opts.HashComments {
		t.Run("supports # line comment", func(t *testing.T) {
			result := format("SELECT alpha # commment\nFROM beta")
			expected := dedent(`
				SELECT
				  alpha # commment
				FROM
				  beta
			`)
			assertEqual(t, result, expected)
		})
	}

	if opts.DoubleSlashComments {
		t.Run("supports // line comment", func(t *testing.T) {
			result := format("SELECT alpha // commment\nFROM beta")
			expected := dedent(`
				SELECT
				  alpha // commment
				FROM
				  beta
			`)
			assertEqual(t, result, expected)
		})
	}

	if opts.NestedBlockComments {
		t.Run("supports nested block comments", func(t *testing.T) {
			result := format("SELECT alpha /* /* commment */ */ FROM beta")
			expected := dedent(`
				SELECT
				  alpha /* /* commment */ */
				FROM
				  beta
			`)
			assertEqual(t, result, expected)
		})
	}
}
