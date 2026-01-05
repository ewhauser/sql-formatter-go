package sqlformatter

import "testing"

func TestPropertyAccessWithParenthesis(t *testing.T) {
	// Schema-qualified table names should preserve space before column definitions
	t.Run("schema-qualified table in CREATE TABLE preserves space before parenthesis", func(t *testing.T) {
		result := formatPostgres(t, "CREATE TABLE public.foo (id int);")
		expected := dedent(`
			CREATE TABLE public.foo (id int);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("schema-qualified table in CREATE TABLE with multiple columns preserves space", func(t *testing.T) {
		result := formatPostgres(t, "CREATE TABLE public.auth_record (service_name text NOT NULL, token text NOT NULL);")
		expected := dedent(`
			CREATE TABLE public.auth_record (service_name text NOT NULL, token text NOT NULL);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("multi-level schema-qualified table preserves space", func(t *testing.T) {
		result := formatPostgres(t, "CREATE TABLE mydb.public.foo (id int);")
		expected := dedent(`
			CREATE TABLE mydb.public.foo (id int);
		`)
		assertEqual(t, result, expected)
	})

	// sqlc helper functions should NOT have space before parenthesis
	t.Run("sqlc.arg function call has no space before parenthesis", func(t *testing.T) {
		result := formatPostgres(t, "SELECT * FROM foo WHERE bar = sqlc.arg(baz);")
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			WHERE
			  bar = sqlc.arg(baz);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("sqlc.arg with space in input is normalized to no space", func(t *testing.T) {
		result := formatPostgres(t, "SELECT * FROM foo WHERE bar = sqlc.arg (baz);")
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			WHERE
			  bar = sqlc.arg(baz);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("sqlc.narg function call has no space before parenthesis", func(t *testing.T) {
		result := formatPostgres(t, "SELECT * FROM foo WHERE bar = sqlc.narg(baz);")
		expected := dedent(`
			SELECT
			  *
			FROM
			  foo
			WHERE
			  bar = sqlc.narg(baz);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("sqlc.embed function call has no space before parenthesis", func(t *testing.T) {
		result := formatPostgres(t, "SELECT sqlc.embed(foo) FROM bar;")
		expected := dedent(`
			SELECT
			  sqlc.embed(foo)
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("SQLC uppercase also works", func(t *testing.T) {
		result := formatPostgres(t, "SELECT SQLC.arg(foo) FROM bar;")
		expected := dedent(`
			SELECT
			  SQLC.arg(foo)
			FROM
			  bar;
		`)
		assertEqual(t, result, expected)
	})

	// Non-sqlc property access with parenthesis should preserve space
	t.Run("generic property access followed by parenthesis preserves space", func(t *testing.T) {
		result := formatPostgres(t, "SELECT obj.method (arg1, arg2) FROM foo;")
		expected := dedent(`
			SELECT
			  obj.method (arg1, arg2)
			FROM
			  foo;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("schema-qualified function call preserves space", func(t *testing.T) {
		result := formatPostgres(t, "SELECT public.my_function (arg) FROM foo;")
		expected := dedent(`
			SELECT
			  public.my_function (arg)
			FROM
			  foo;
		`)
		assertEqual(t, result, expected)
	})

	// Other DDL statements with schema-qualified names
	t.Run("INSERT INTO schema-qualified table preserves space", func(t *testing.T) {
		result := formatPostgres(t, "INSERT INTO public.foo (id, name) VALUES (1, 'bar');")
		expected := dedent(`
			INSERT INTO
			  public.foo (id, name)
			VALUES
			  (1, 'bar');
		`)
		assertEqual(t, result, expected)
	})

	t.Run("UPDATE schema-qualified table", func(t *testing.T) {
		result := formatPostgres(t, "UPDATE public.foo SET bar = 1 WHERE id = 2;")
		expected := dedent(`
			UPDATE public.foo
			SET
			  bar = 1
			WHERE
			  id = 2;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("DELETE FROM schema-qualified table", func(t *testing.T) {
		result := formatPostgres(t, "DELETE FROM public.foo WHERE id = 1;")
		expected := dedent(`
			DELETE FROM public.foo
			WHERE
			  id = 1;
		`)
		assertEqual(t, result, expected)
	})

	t.Run("ALTER TABLE schema-qualified table", func(t *testing.T) {
		result := formatPostgres(t, "ALTER TABLE public.foo ADD COLUMN bar int;")
		expected := dedent(`
			ALTER TABLE public.foo
			ADD COLUMN bar int;
		`)
		assertEqual(t, result, expected)
	})
}
