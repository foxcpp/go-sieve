# AGENTS Guide for `go-sieve`

## Big picture
- Public API is intentionally thin in `sieve.go`: `Load` wires `lexer.Lex` -> `parser.Parse` -> `interp.LoadScript`; runtime is `Script.Execute(ctx, *RuntimeData)`.
- The interpreter is split into compile-time loaders (`interp/load*.go`) and runtime executors (`interp/*.go` command/test types with `Execute`/`Check`).
- Command and test dispatch is table-driven in `interp/load.go` (`commands`/`tests` maps), keyed by lowercased Sieve identifiers.
- `cmd/sieve-run/main.go` is the fastest end-to-end reference for wiring envelope/message/policy and observing actions.

## Key boundaries and data flow
- Lexer (`lexer/`) produces token stream with limits (`lexer.Options.MaxTokens`); parser (`parser/`) builds a permissive AST (`parser.Cmd`, `parser.Test`) with nesting guards.
- `interp.LoadSpec` in `interp/load_generic.go` is the core argument validator/decoder used by most loaders; reuse it instead of hand-parsing args. **It panics (not errors) on programmer-error misconfigurations of `Spec` structs** (missing `MatchStr`/`MatchNum` when `NeedsValue: true`).
- Runtime side effects live in `RuntimeData` (`interp/runtime.go`). The current mechanism is `OnAction func(...)` callback + `AppliedActions []AppliedAction` slice. Fields `RedirectAddr`, `Mailboxes`, `Keep`, `ImplicitKeep` are **deprecated** but still mutated alongside `OnAction`; new code should use `AppliedActions`/`OnAction`.
- External integration is interface-based: `PolicyReader`, `Envelope`, `Message` (`interp/runtime.go`), with in-memory defaults in `interp/message_static.go`. Use `interp.DummyPolicy{}` and `interp.EnvelopeStatic{}` as zero-config defaults in tests.
- Implicit keep is applied by `Script.Execute` after all commands run, not by any individual command. An action cancels implicit keep if `cancelsImplicitKeep() == true` (`ActionKeep`, `ActionDiscard`, `ActionFileInto`, `ActionRedirect`).
- `Script`, `Options`, `Cmd` interface, and `ErrStop` are defined in `interp/script.go`, not `interp/load.go`.

## Project-specific conventions
- Extension gating is strict: loaders must reject extension-specific syntax unless required (e.g., `loadFileInto`/`loadSetFlag` check `s.RequiresExtension(...)`).
- Additions usually require both loader registration and runtime type registration:
  - Add factory in `interp/load.go` maps.
  - Implement loader in `interp/load_*.go` + runtime type in `interp/*.go`.
  - Register new concrete command/test with `gob.Register(...)` in an `init()` for `Script.Save`/`Restore` compatibility. These `init()` calls are **scattered across multiple files** (`action.go`, `control.go`, `variables.go`, `dovecot_testsuite.go`, etc.), not centralized in `interp/binary.go`. Forgetting to register causes a runtime gob error, not a compile error.
- Variable handling is centralized in `RuntimeData.Var`/`SetVar`; namespace behavior (e.g., `envelope.*`) is enforced there.
- `require` command is compile-time only and returns `nil` command (`interp/load_control.go`), so do not expect a runtime `CmdRequire`.
- `sieve.go` exports `ActionFileInfo` as a public alias for `interp.ActionFileInto` — the "Info" vs "Into" inconsistency is a known API wart; do not rename without considering API breakage.
- New interpeter Options should also be added to `savedOptions`, see `Script.Save` (`interp/binary.go`).
- Message body and other potentially large byte streams should be accessed via streaming API only (`io.Reader` and similar). `io.ReadAll` is allowed only for test fixtures or when size is guaranteed to be small; production code should not read entire message bodies into memory.
- List of supported extensions is provided in README and should be kept up-to-date.

## Workflows that matter
```bash
go build -v ./...
go test -v ./...
```
- CI runs exactly those two commands on Go 1.20 (`.github/workflows/go.yml`).
- Run a single test: `go test -v -run TestName ./tests/` or `go test -v -run TestName .`
- Dovecot tests rely on git submodule content under `tests/pigeonhole`; run `git submodule update --init` before diagnosing missing test files. CI handles this automatically.
- End-to-end local run example:
```bash
go run ./cmd/sieve-run -scriptPath ./cmd/sieve-run/test.sieve -eml ./cmd/sieve-run/msg.eml -from from@test.com -to to@test.com
```

## Test layout and expected behavior
- Root-level `execute_test.go` validates API-level behavior via `sieve.Load` + `Script.Execute`; no submodule dependency.
- `tests/` package drives Dovecot `.svtest` scripts via `RunDovecotTest*` helpers (`tests/run.go`); requires `tests/pigeonhole` submodule.
- To activate Dovecot testsuite commands, set `opts.Interp.T = t` before loading — this enables `vnd.dovecot.testsuite` extension and wires the Go test framework.
- Some upstream tests are intentionally disabled with documented reasons (`tests/base_test.go`); preserve these expectations unless fixing underlying parser/address behavior.
- `interp/dovecot_testsuite.go` intentionally treats `test_error` checks as no-op pass (go-sieve stops on first error; Pigeonhole collects multiple). `test_imap_metadata_set` is unimplemented and will error if used in `.svtest` files.
- `CmdDovecotTest.Execute` copies `RuntimeData` via `d.Copy()` for isolation between test blocks — commands inside `test { ... }` do not affect outer state.
- New upstream tests should be added to `tests/` as `.svtest` files, not as Go unit tests, to preserve the Dovecot test suite as the canonical source of accepted-invalid-script cases.
- New features should be added with both Go unit tests (for API-level validation) and `.svtest` cases (to preserve Dovecot test suite coverage).
- Tests that expect certain MTA behavior should be added to `tests/execute.go`. `ExecuteTestEnvironment` is an interface to mock MTA environment that applies actions and provides access to results (e.g. sent messages); use it for tests that care about action semantics. For tests that only care about script acceptance/rejection, `RunDovecotTest` without extra parameters is sufficient. 
- Tests added to `tests/execute.go` should pass with the simple `simpleExecuteRuntime` implemented in `tests/execute_test.go`. This ensures that go-sieve is able to provide enough data to the MTA to execute correct decisions.
- Other than for sourcing .svtest files, Pigeonhole source code is not used for anything and is never to be modified.

## Known caveats to preserve
- `README.md` lists accepted-invalid-script gaps and address parsing caveats; avoid changes that silently alter these behaviors without targeted tests.
- `tests/compile_test.go` is the canonical record of accepted-invalid-script cases; parser changes that reject previously-accepted scripts will break it.
- `interp/options.go` fields mutated at runtime by `CmdDovecotConfigSet` (e.g., `MaxVariableLen`) — options are not immutable after load.
