package sqlformatter

import "testing"

func behavesLikePostgresqlFormatter(t *testing.T, format FormatFn) {
	t.Helper()
	behavesLikeSqlFormatter(t, format)
	supportsNumbers(t, format, numbersConfig{Underscore: true})
	supportsComments(t, format, commentsConfig{NestedBlockComments: true})
	supportsCommentOn(t, format)
	supportsArrayAndMapAccessors(t, format)
	supportsAlterTable(t, format, alterTableConfig{AddColumn: true, DropColumn: true, RenameTo: true, RenameColumn: true})
	supportsDeleteFrom(t, format)
	supportsInsertInto(t, format)
	supportsOnConflict(t, format)
	supportsBetween(t, format)
	supportsIsDistinctFrom(t, format)
	supportsReturning(t, format)
	supportsWindow(t, format)
	supportsDataTypeCase(t, format)

	t.Run("allows $ character as part of identifiers", func(t *testing.T) {
		result := format("SELECT foo$, some$$ident")
		expected := dedent(`
			SELECT
			  foo$,
			  some$$ident
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats type-cast operator without spaces", func(t *testing.T) {
		result := format("SELECT 2 :: numeric AS foo;")
		expected := dedent(`
			SELECT
			  2::numeric AS foo;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats SELECT DISTINCT ON syntax", func(t *testing.T) {
		result := format("SELECT DISTINCT ON (c1, c2) c1, c2 FROM tbl;")
		expected := dedent(`
			SELECT DISTINCT
			  ON (c1, c2) c1,
			  c2
			FROM
			  tbl;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats ALTER TABLE ... ALTER COLUMN", func(t *testing.T) {
		result := format(`ALTER TABLE t ALTER COLUMN foo SET DATA TYPE VARCHAR;
         ALTER TABLE t ALTER COLUMN foo SET DEFAULT 5;
         ALTER TABLE t ALTER COLUMN foo DROP DEFAULT;
         ALTER TABLE t ALTER COLUMN foo SET NOT NULL;
         ALTER TABLE t ALTER COLUMN foo DROP NOT NULL;`)
		expected := dedent(`
			ALTER TABLE t
			ALTER COLUMN foo
			SET DATA TYPE VARCHAR;

			ALTER TABLE t
			ALTER COLUMN foo
			SET DEFAULT 5;

			ALTER TABLE t
			ALTER COLUMN foo
			DROP DEFAULT;

			ALTER TABLE t
			ALTER COLUMN foo
			SET NOT NULL;

			ALTER TABLE t
			ALTER COLUMN foo
			DROP NOT NULL;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("allows TYPE to be used as identifier", func(t *testing.T) {
		result := format("SELECT type, modified_at FROM items;")
		expected := dedent(`
			SELECT
			  type,
			  modified_at
			FROM
			  items;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("does not recognize common field names as keywords", func(t *testing.T) {
		result := format("SELECT id, type, name, location, label, password FROM release;", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  id,
			  type,
			  name,
			  location,
			  label,
			  password
			FROM
			  release;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats DEFAULT VALUES clause", func(t *testing.T) {
		result := format("INSERT INTO items default values RETURNING id;", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			INSERT INTO
			  items
			DEFAULT VALUES
			RETURNING
			  id;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("treats TEXT as data-type", func(t *testing.T) {
		result := format("CREATE TABLE foo (items text);", FormatOptions{DataTypeCase: KeywordCaseUpper})
		expected := dedent(`
			CREATE TABLE foo (items TEXT);
		`)
		assertEqual(t, result, expected)

		result = format("CREATE TABLE foo (text VARCHAR(100));", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected = dedent(`
			CREATE TABLE foo (text VARCHAR(100));
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats TIMESTAMP WITH TIMEZONE as data type", func(t *testing.T) {
		result := format("create table time_table (id int primary key, created_at timestamp with time zone);", FormatOptions{DataTypeCase: KeywordCaseUpper})
		expected := dedent(`
			create table time_table (
			  id INT primary key,
			  created_at TIMESTAMP WITH TIME ZONE
			);
		`)
		assertEqual(t, result, expected)
	})
}
