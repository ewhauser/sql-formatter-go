package sqlformatter

import (
	"regexp"
	"strings"
	"testing"
)

type joinOptions struct {
	Without        []string
	Additionally   []string
	SupportsUsing  bool
	SupportsApply  bool
}

func supportsJoin(t *testing.T, format FormatFn, opts ...joinOptions) {
	t.Helper()
	cfg := joinOptions{SupportsUsing: true}
	if len(opts) > 0 {
		cfg = opts[0]
	}

	unsupported := regexp.MustCompile(`^whateve_!%&$`)
	if len(cfg.Without) > 0 {
		unsupported = regexp.MustCompile(strings.Join(cfg.Without, "|"))
	}
	isSupported := func(join string) bool { return !unsupported.MatchString(join) }

	joins := []string{
		"JOIN",
		"INNER JOIN",
		"CROSS JOIN",
		"LEFT JOIN",
		"LEFT OUTER JOIN",
		"RIGHT JOIN",
		"RIGHT OUTER JOIN",
		"FULL JOIN",
		"FULL OUTER JOIN",
		"NATURAL JOIN",
		"NATURAL INNER JOIN",
		"NATURAL LEFT JOIN",
		"NATURAL LEFT OUTER JOIN",
		"NATURAL RIGHT JOIN",
		"NATURAL RIGHT OUTER JOIN",
		"NATURAL FULL JOIN",
		"NATURAL FULL OUTER JOIN",
	}
	if len(cfg.Additionally) > 0 {
		joins = append(joins, cfg.Additionally...)
	}
	for _, join := range joins {
		if !isSupported(join) {
			continue
		}
		join := join
		t.Run("supports "+join, func(t *testing.T) {
			result := format(`
          SELECT * FROM customers
          `+join+` orders ON customers.customer_id = orders.customer_id
          `+join+` items ON items.id = orders.id;
        `)
			expected := dedent(`
				SELECT
				  *
				FROM
				  customers
				  ` + join + ` orders ON customers.customer_id = orders.customer_id
				  ` + join + ` items ON items.id = orders.id;
			`)
			assertEqual(t, result, expected)
		})
	}

	t.Run("properly uppercases JOIN ON", func(t *testing.T) {
		result := format("select * from customers join foo on foo.id = customers.id;", FormatOptions{KeywordCase: KeywordCaseUpper})
		expected := dedent(`
			SELECT
			  *
			FROM
			  customers
			  JOIN foo ON foo.id = customers.id;
		`)
		assertEqual(t, result, expected)
	})

	if cfg.SupportsUsing {
		t.Run("properly uppercases JOIN USING", func(t *testing.T) {
			result := format("select * from customers join foo using (id);", FormatOptions{KeywordCase: KeywordCaseUpper})
			expected := dedent(`
				SELECT
				  *
				FROM
				  customers
				  JOIN foo USING (id);
			`)
			assertEqual(t, result, expected)
		})
	}

	if cfg.SupportsApply {
		for _, apply := range []string{"CROSS APPLY", "OUTER APPLY"} {
			apply := apply
			t.Run("supports "+apply, func(t *testing.T) {
				result := format("SELECT * FROM customers " + apply + " fn(customers.id)")
				expected := dedent(`
					SELECT
					  *
					FROM
					  customers
					  ` + apply + ` fn (customers.id)
				`)
				assertEqual(t, result, expected)
			})
		}
	}
}
