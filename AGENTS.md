# AGENTS Guide for `go-sieve`

## Big picture
- Public API is intentionally thin in `sieve.go`: `Load` wires `lexer.Lex` -> `parser.Parse` -> `interp.LoadScript`; runtime is `Script.Execute(ctx, *RuntimeData)`.
- The interpreter is split into compile-time loaders (`interp/load*.go`) and runtime executors (`interp/*.go` command/test types with `Execute`/`Check`).
- Command and test dispatch is table-driven in `interp/load.go` (`commands`/`tests` maps), keyed by lowercased Sieve identifiers.
- `cmd/sieve-run/main.go` is the fastest end-to-end reference for wiring envelope/message/policy and observing actions (`RedirectAddr`, `Mailboxes`, `Flags`, `Keep`).

## Key boundaries and data flow
- Lexer (`lexer/`) produces token stream with limits (`lexer.Options.MaxTokens`); parser (`parser/`) builds a permissive AST (`parser.Cmd`, `parser.Test`) with nesting guards.
- `interp.LoadSpec` in `interp/load_generic.go` is the core argument validator/decoder used by most loaders; reuse it instead of hand-parsing args.
- Runtime side effects live in `RuntimeData` (`interp/runtime.go`): actions mutate `RedirectAddr`, `Mailboxes`, flags, keep state, and variables.
- External integration is interface-based: `PolicyReader`, `Envelope`, `Message` (`interp/runtime.go`), with in-memory defaults in `interp/message_static.go`.

## Project-specific conventions
- Extension gating is strict: loaders must reject extension-specific syntax unless required (e.g., `loadFileInto`/`loadSetFlag` check `s.RequiresExtension(...)`).
- Additions usually require both loader registration and runtime type registration:
  - Add factory in `interp/load.go` maps.
  - Implement loader in `interp/load_*.go` + runtime type in `interp/*.go`.
  - Register new concrete command/test in `init()` with `gob.Register(...)` for `Script.Save`/`Restore` compatibility (`interp/binary.go`).
- Variable handling is centralized in `RuntimeData.Var`/`SetVar`; namespace behavior (e.g., `envelope.*`) is enforced there.
- `require` command is compile-time only and returns `nil` command (`interp/load_control.go`), so do not expect a runtime `CmdRequire`.

## Workflows that matter
```bash
go build ./...
go test -v ./...
```
- CI runs exactly those two commands on Go 1.20 (`.github/workflows/go.yml`).
- Dovecot tests rely on git submodule content under `tests/pigeonhole` (`.gitmodules`); ensure submodules are present before diagnosing missing test files.
- End-to-end local run example:
```bash
go run ./cmd/sieve-run -scriptPath ./cmd/sieve-run/test.sieve -eml ./cmd/sieve-run/msg.eml -from from@test.com -to to@test.com
```

## Test layout and expected behavior
- Root-level tests (e.g., `execute_test.go`) validate API-level behavior via `sieve.Load` + `Script.Execute`.
- `tests/` package drives Dovecot `.svtest` scripts via `RunDovecotTest*` helpers (`tests/run.go`) and sets `RuntimeData.Namespace` for include/compile/run test operations.
- Some upstream tests are intentionally disabled with documented reasons (`tests/base_test.go`); preserve these expectations unless fixing underlying parser/address behavior.

## Known caveats to preserve
- `README.md` lists accepted-invalid-script gaps and address parsing caveats; avoid changes that silently alter these behaviors without targeted tests.
- `interp/dovecot_testsuite.go` intentionally treats `test_error` checks as no-op pass due to format mismatch with Pigeonhole behavior.

