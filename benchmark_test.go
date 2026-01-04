package sqlformatter

import (
	"strings"
	"testing"
)

var benchmarkSink string

func benchmarkFormat(b *testing.B, sql string, opts ...FormatOptions) {
	b.Helper()
	b.ReportAllocs()
	cfg := FormatOptionsWithLanguage{Language: LanguagePostgresql}
	if len(opts) > 0 {
		cfg.FormatOptions = opts[0]
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := Format(sql, cfg)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = out
	}
}

func BenchmarkFormat_PostgresComplex(b *testing.B) {
	sql := `
WITH RECURSIVE tree AS (
  SELECT id, parent_id, name, 1 AS depth
  FROM nodes
  WHERE parent_id IS NULL
  UNION ALL
  SELECT n.id, n.parent_id, n.name, t.depth + 1
  FROM nodes n
  JOIN tree t ON n.parent_id = t.id
),
ranked AS (
  SELECT
    t.*,
    row_number() OVER (PARTITION BY parent_id ORDER BY name) AS rn,
    lag(name) OVER (ORDER BY name) AS prev_name,
    lead(name) OVER (ORDER BY name) AS next_name
  FROM tree t
)
SELECT
  r.id,
  r.name,
  CASE
    WHEN r.depth = 1 THEN 'root'
    WHEN r.depth < 4 THEN 'branch'
    ELSE 'leaf'
  END AS node_type,
  jsonb_build_object(
    'path', array_agg(r.name) OVER (PARTITION BY r.id),
    'siblings', count(*) OVER (PARTITION BY r.parent_id)
  ) AS meta
FROM ranked r
LEFT JOIN LATERAL (
  SELECT sum(amount) AS total
  FROM payments p
  WHERE p.node_id = r.id
  GROUP BY p.node_id
) pay ON true
WHERE r.name ILIKE '%foo%'
  AND r.id = ANY (ARRAY[1, 2, 3, 4, 5]::int[])
ORDER BY r.depth, r.name
LIMIT 100 OFFSET 20;
`
	benchmarkFormat(b, sql)
}

func BenchmarkFormat_PostgresDDLAndFunctions(b *testing.B) {
	sql := `
CREATE OR REPLACE FUNCTION public.compute_score(
  user_id uuid,
  weights jsonb DEFAULT '{}'::jsonb
)
RETURNS numeric
LANGUAGE sql
AS $$
  WITH base AS (
    SELECT
      u.id,
      coalesce(sum(p.amount), 0) AS spent,
      count(o.id) AS orders
    FROM users u
    LEFT JOIN orders o ON o.user_id = u.id
    LEFT JOIN payments p ON p.order_id = o.id
    WHERE u.id = user_id
    GROUP BY u.id
  )
  SELECT
    (base.spent * coalesce((weights->>'spent')::numeric, 1)) +
    (base.orders * coalesce((weights->>'orders')::numeric, 1)) AS score
  FROM base;
$$;

CREATE TABLE IF NOT EXISTS public.audit_log (
  id bigserial PRIMARY KEY,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  created_at timestamp with time zone DEFAULT now()
);
`
	benchmarkFormat(b, sql)
}

func BenchmarkFormat_PostgresManyStatements(b *testing.B) {
	stmt := "SELECT id, name, created_at FROM users WHERE id = $1 AND name ILIKE $2;"
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		sb.WriteString(stmt)
		sb.WriteString("\n")
	}
	benchmarkFormat(b, sb.String())
}

func BenchmarkFormat_PostgresWideInserts(b *testing.B) {
	values := strings.Repeat("(1, 'foo', true, now(), jsonb_build_object('k', 'v')),", 200)
	sql := "INSERT INTO events (id, name, active, created_at, meta) VALUES " +
		strings.TrimSuffix(values, ",") + ";"
	benchmarkFormat(b, sql)
}
