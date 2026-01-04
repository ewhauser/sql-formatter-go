# sql-formatter-go

A Go port of sql-formatter with PostgreSQL support.
The formatter output is byte-for-byte compatible with the upstream JavaScript `sql-formatter` for PostgreSQL.

## Install

```sh
go install ./cmd/sql-formatter
```

## CLI (drop-in for `sql-formatter`)

The CLI mirrors the upstream `sql-formatter` options and defaults, with multi-file and multithreaded support.
Only `postgresql` (and `sql`, as an alias) are supported in this port.
Other dialects will return an error.

```sh
sql-formatter -h
```

```
usage: sql-formatter [-h] [-o OUTPUT] [-l {postgresql,sql}] [-c CONFIG] [--version] [FILE...]

SQL Formatter

positional arguments:
  FILE            Input SQL file(s) (defaults to stdin)

optional arguments:
  -h, --help      show this help message and exit
  -o, --output    OUTPUT
                    File to write SQL output (defaults to stdout)
  --fix           Update the file in-place
  -l, --language  {postgresql,sql}
                    SQL dialect (defaults to basic sql)
  -c, --config    CONFIG
                    Path to config JSON file or json string (will find a file named '.sql-formatter.json' or use default configs if unspecified)
  --workers       number of concurrent workers for multiple files (0 = NumCPU)
  --version       show program's version number and exit
```

Examples:

```sh
echo 'select * from tbl where id = 3' | sql-formatter
sql-formatter --fix path/to/query.sql
sql-formatter -o formatted.sql path/to/query.sql
sql-formatter --fix path/to/one.sql path/to/two.sql
sql-formatter --workers 8 --fix path/to/**/*.sql
```

### Config file

The CLI reads `.sql-formatter.json` from the current directory (or any parent), or accepts a JSON string/file via `--config`.

```json
{
  "language": "postgresql",
  "tabWidth": 2,
  "keywordCase": "upper",
  "linesBetweenQueries": 2
}
```

Supported fields (PostgreSQL-only in this port):

- `language`
- `tabWidth`
- `useTabs`
- `keywordCase`
- `identifierCase`
- `dataTypeCase`
- `functionCase`
- `indentStyle`
- `logicalOperatorNewline`
- `expressionWidth`
- `linesBetweenQueries`
- `denseOperators`
- `newlineBeforeSemicolon`
- `params`
- `paramTypes`

## Go API

```go
formatted, err := sqlformatter.Format(query, sqlformatter.FormatOptionsWithLanguage{
    Language: sqlformatter.LanguagePostgresql,
})
```

## Benchmarks

Large repository run over `/Users/ewhauser/working/cadencerpm/monorepo/go/**/*.sql` (3400 files, 24.47MB).
Formatting only (read + format, no write).

Environment: Apple M3 Max (darwin/arm64), Go 1.25.0, Node v25.2.1, sql-formatter v15.6.12.

| Implementation | Workers | Time | Throughput |
| --- | --- | --- | --- |
| Go (this repo) | 1 | 2.211s | 1537.6 files/s (11.1 MB/s) |
| Go (this repo) | 16 | 0.494s | 6878.8 files/s (49.5 MB/s) |
| JS (sql-formatter) | 1 | 17.706s | 192.0 files/s (1.4 MB/s) |

## Notes

- Only PostgreSQL is supported (alias `sql` uses PostgreSQL rules).
- The CLI is compatible with the upstream `sql-formatter` option set; unsupported dialects will return an error.
