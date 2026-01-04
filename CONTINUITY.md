Goal (incl. success criteria):
- Port existing /Users/ewhauser/working/sql-formatter (Postgres formatting) to Go in this repo; ensure compatibility with existing sql-formatter Postgres behavior and port all Postgres formatting tests so they pass.
Constraints/Assumptions:
- Follow provided AGENTS.md instructions in prompt (no file found in repo).
- Maintain and update CONTINUITY.md at start of each turn.
- approval_policy: never; sandbox_mode: danger-full-access; network_access: enabled.
- Tests may use testify require/assert (user-approved).
Key decisions:
- Using a native Go tokenizer/parser/formatter port rather than embedding JS.
- Postgres dialect data generated from TS sources into Go files.
State:
- Core formatter/tokenizer/parser implemented in Go; Postgres dialect and keywords/functions generated.
- All Postgres tests ported; go test ./... passes.
Done:
- Created Go module and core formatter/tokenizer/parser/AST/layout/indentation.
- Added Postgres dialect options and keyword/data/function lists.
- Added initial Postgres test wrapper and case feature tests.
- Fixed inline layout handling (LayoutWriter interface + InlineLayout usage) to avoid over-inlining.
- Preserved disable comment blocks, fixed nested comment matching for unterminated comments.
- Fixed comment attachment for property access and BETWEEN chains; updated tests for unicode and tabular CASE.
Now:
- All tests passing; ready for user validation or further requests.
Next:
- (Optional) User review/validation or additional dialect support if requested.
Open questions (UNCONFIRMED if needed):
- None.
Working set (files/ids/commands):
- go.mod
- sql_formatter.go
- formatter.go
- parser.go
- tokenizer.go
- postgresql_dialect.go
- languages/postgresql/keywords.go
- helpers_test.go
- postgresql_test.go
- features_case_test.go
- cmd: go test ./...
