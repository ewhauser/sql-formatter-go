package sqlformatter

import "testing"

type createTableConfig struct {
	OrReplace    bool
	IfNotExists  bool
	ColumnComment bool
	TableComment  bool
}

func supportsCreateTable(t *testing.T, format FormatFn, cfg createTableConfig) {
	t.Helper()
	t.Run("formats short CREATE TABLE", func(t *testing.T) {
		result := format("CREATE TABLE tbl (a INT PRIMARY KEY, b TEXT);")
		expected := dedent(`
			CREATE TABLE tbl (a INT PRIMARY KEY, b TEXT);
		`)
		assertEqual(t, result, expected)
	})

	t.Run("formats long CREATE TABLE", func(t *testing.T) {
		result := format("CREATE TABLE tbl (a INT PRIMARY KEY, b TEXT, c INT NOT NULL, doggie INT NOT NULL);")
		expected := dedent(`
			CREATE TABLE tbl (
			  a INT PRIMARY KEY,
			  b TEXT,
			  c INT NOT NULL,
			  doggie INT NOT NULL
			);
		`)
		assertEqual(t, result, expected)
	})

	if cfg.OrReplace {
		t.Run("formats short CREATE OR REPLACE TABLE", func(t *testing.T) {
			result := format("CREATE OR REPLACE TABLE tbl (a INT PRIMARY KEY, b TEXT);")
			expected := dedent(`
				CREATE OR REPLACE TABLE tbl (a INT PRIMARY KEY, b TEXT);
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.IfNotExists {
		t.Run("formats short CREATE TABLE IF NOT EXISTS", func(t *testing.T) {
			result := format("CREATE TABLE IF NOT EXISTS tbl (a INT PRIMARY KEY, b TEXT);")
			expected := dedent(`
				CREATE TABLE IF NOT EXISTS tbl (a INT PRIMARY KEY, b TEXT);
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.ColumnComment {
		t.Run("formats CREATE TABLE with column comments", func(t *testing.T) {
			result := format("CREATE TABLE tbl (a INT COMMENT 'Hello world!', b TEXT COMMENT 'Here we are!');")
			expected := dedent(`
				CREATE TABLE tbl (
				  a INT COMMENT 'Hello world!',
				  b TEXT COMMENT 'Here we are!'
				);
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.TableComment {
		t.Run("formats CREATE TABLE with comment", func(t *testing.T) {
			result := format("CREATE TABLE tbl (a INT, b TEXT) COMMENT = 'Hello, world!';")
			expected := dedent(`
				CREATE TABLE tbl (a INT, b TEXT) COMMENT = 'Hello, world!';
			`)
			assertEqual(t, result, expected)
		})
	}

	t.Run("correctly indents CREATE TABLE in tabular style", func(t *testing.T) {
		result := format(`CREATE TABLE foo (
          id INT PRIMARY KEY NOT NULL,
          fname VARCHAR NOT NULL
        );`, FormatOptions{IndentStyle: IndentStyleTabularLeft})
		expected := dedent(`
			CREATE    TABLE foo (
			          id INT PRIMARY KEY NOT NULL,
			          fname VARCHAR NOT NULL
			          );
		`)
		assertEqual(t, result, expected)
	})
}
