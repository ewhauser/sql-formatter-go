package sqlformatter

import "testing"

func supportsCommentOn(t *testing.T, format FormatFn) {
	t.Helper()
	t.Run("formats COMMENT ON", func(t *testing.T) {
		result := format("COMMENT ON TABLE my_table IS 'This is an awesome table.';")
		expected := dedent(`
			COMMENT ON TABLE my_table IS 'This is an awesome table.';
		`)
		assertEqual(t, result, expected)

		result = format("COMMENT ON COLUMN my_table.ssn IS 'Social Security Number';")
		expected = dedent(`
			COMMENT ON COLUMN my_table.ssn IS 'Social Security Number';
		`)
		assertEqual(t, result, expected)
	})
}
