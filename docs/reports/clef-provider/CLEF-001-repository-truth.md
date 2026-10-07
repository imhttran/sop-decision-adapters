# CLEF-001 — Repository Truth: sop-decision-adapters

> Documentation-only reconnaissance. Source is authoritative; where documentation
disagrees with source, the divergence is recorded and resolved in favor of source.
> No production code was changed to produce this report. Only this report file was
created/updated.

## 1. Repository state

| Fact | Value |
|---|---|
| Branch | `main` (`git status --short --branch` reports `## main...origin/main`) |
| HEAD | `9718dad57240a2532b3d8b169c053dae4338b35f` (`git rev-parse HEAD`) |
| Upstream | `origin/main` (tracking; ahead/behind not reported by `git status`) |
| Module path | `github.com/imhttran/sop-decision-adapters` (`go.mod`) |
| Go directive | `go 1.22` (`go.mod`) |
| Working tree | pre-existing user-owned changes present — see §1.1 |
| Build / test baseline | see §2 |

### 1.1 Working-tree state

`git status --short --branch` reports:

```
## main...origin/main
 M docs/plans/PLAN-Clef-Provider-Integration.md
?? docs/reports/clef-provider/CLEF-001-repository-truth.md
```

- `docs/plans/PLAN-Clef-Provider-Integration.md` is modified (user-owned; preserved,
  not reverted or edited by this task).
- `docs/reports/clef-provider/CLEF-001-repository-truth.md` is untracked and is this
  task's own deliverable.
- Nothing else is modified, added, or deleted. No cleanup, reset, or discard was
  performed.

> Note: the SOP state directory `.agent-sdlc/` is out of scope and was not modified.

## 2. Go toolchain and build/test baseline

The Go module identity and language directive are taken from `go.mod` (authoritative):

- module `github.com/imhttran/sop-decision-adapters`
- `go 1.22`

The concrete toolchain binary version (`go version`) was **not captured** in this
pass because that command is not permitted under the allowed command set; it is
recorded as UNKNOWN rather than inferred from documentation.

Baseline commands and results (run from the repository root):

| Command | Result |
|---|---|
| `go build ./...` | exit 0 (clean build, no output) |
| `go vet ./...` | exit 0 (no findings) |
| `go test ./...` | exit 0 — see per-package output below |

`go test ./...` output:

```
?   	github.com/imhttran/sop-decision-adapters/cmd/sop-decision-adapter	[no test files]
ok  	github.com/imhttran/sop-decision-adapters/decision	(cached)
ok  	github.com/imhttran/sop-decision-adapters/internal/providers/julia	(cached)
ok  	github.com/imhttran/sop-decision-adapters/internal/providers/nimble	(cached)
ok  	github.com/imhttran/sop-decision-adapters/tests	(cached)
```

No failing or skipped tests were observed. Results were cached at capture time, which
means the packages were previously built and tested successfully with the same inputs.

## 3. Module layout (observed)

```
sop-decision-adapters/
├── go.mod                 module github.com/imhttran/sop-decision-adapters, go 1.22
├── .env.example           environment template
├── cmd/
│   └── sop-decision-adapter/
│       └── main.go        CLI entry point (subcommand "decide")
├── decision/              provider-neutral public contract
├── internal/
│   ├── decision/          present but empty in listing (no Go sources observed)
│   └── providers/
│       ├── julia/
│       └── nimble/
├── tests/                 top-level test assets
├── tools/                 repository-owned helpers (e.g. tools/julia/infer.py)
└── docs/reports/clef-provider/
```

## 4. Provider interface and request/result types

File/symbol evidence from the `decision` package (the authoritative public contract):

- `decision/provider.go` — `type Provider interface` with exactly three methods:
  - `Name() string` — stable adapter identifier (e.g. `"nimble"`).
  - `Available(ctx context.Context) bool` — must not perform an inference request.
  - `Decide(ctx context.Context, req DecisionRequest) (DecisionResult, error)` —
    evaluates every question in `req` against `req.State` and returns a normalized,
    provider-neutral result.
  Source comment: "Callers such as agentic-sop depend only on this contract, never
  on provider-specific types."
- `decision/request.go` — `DecisionRequest`, `Question`, `QuestionType`
  (choice/boolean/score), with validation and `Validate`.
- `decision/result.go` — `DecisionResult`, `Answer`, `AnswerType`, `Usage`, with `Validate`.
- `decision/errors.go` — error taxonomy (`ErrorKind`, `ProviderError`, sentinel kinds
  such as invalid-request, unavailable, malformed-response, provider-failure) and
  sentinel errors used by callers (`errors.Is(err, decision.ErrUnavailable)`,
  `errors.Is(err, decision.ErrInvalidRequest)`).

### 4.1 Capability representation

Capability is represented provider-neutrally by `QuestionType` / `AnswerType`
(choice, boolean, score) in the `decision` package. Adapter-native representations
(e.g. Julia's `qtype` integer, Nimble/SystemOne's "noul" boolean) are confined to their
adapter packages and never appear in the public `decision` package.

## 5. Existing providers

Both adapters live under `internal/providers/` and each implements `decision.Provider`.

### 5.1 Nimble — `internal/providers/nimble`

- `internal/providers/nimble/nimble.go` — `type Provider struct`, `New(cfg Config, transport Transport)`,
  `NewFromEnv()` / `ConfigFromEnv()`, `Name() string`, `Available(ctx) bool`, `Decide(ctx, req)`.
- Request path: `decision.Provider -> internal/providers/nimble -> HTTP -> Ollama -> Nimble`.
- Native translation: `internal/providers/nimble/systemone.go` (`BuildSystemOneRequest`,
  `NormalizeSystemOneResponse`).
- Transport: `internal/providers/nimble/transport.go` — `Transport` interface
  (`SystemOne`, `ListModels`), `HTTPTransport` over Go's `net/http`.
- Noul boolean representation is a SystemOne implementation detail confined to this package.
- Config: `internal/providers/nimble/config.go`; tests: `config_test.go`, `nimble_test.go`,
  `translate_test.go`, `transport_test.go`, `golden_test.go`.

### 5.2 Julia — `internal/providers/julia`

- `internal/providers/julia/julia.go` — `type Provider struct`, `New(cfg Config, runner Runner)`,
  `NewFromEnv()` / `ConfigFromEnv()`, `Name() string`, `Available(ctx) bool`, `Decide(ctx, req)`.
- Request path: `decision.Provider -> internal/providers/julia -> Runner -> ONNX inference -> Julia`.
- Runner seam: `internal/providers/julia/runner.go` — `Runner` interface,
  `RunnerFunc`, `CommandRunner` (external helper process; JSON-in/JSON-out).
- Native translation: `internal/providers/julia/translate.go` (`BuildInputs`, `NormalizeOutputs`);
  `semantics.go` holds adapter-native semantics.
- Default helper: `tools/julia/infer.py`, invoked via `JULIA_PYTHON`.
- Tests: `config_test.go`, `julia_test.go`, `runner_test.go`, `translate_test.go`.

Both adapters translate and normalize only; neither defines or enforces SOP policy.

## 6. Provider selection / registration

The `decision` package contains **no global registry or factory map**. Selection is by
construction: each adapter exposes a package-level `New(cfg, dep)` constructor plus a
`NewFromEnv()`/`ConfigFromEnv()` helper, and the caller dispatches on a provider name.

CLI wiring lives in `cmd/sop-decision-adapter/main.go` (package `main`), not in a
`cmd/decision` path and not inside the provider-neutral contract:

- `availableProviders = "nimble, julia"` — the set of selectable provider names.
- `newProvider(name, baseURL, model string, timeout time.Duration) (decision.Provider, error)`
  — the factory/selection seam. A `switch` on the lower-cased/trimmed `-provider` value:
  - `""` or `"nimble"` → `nimble.ConfigFromEnv()`, flag overrides (`-base-url`, `-model`,
    `-timeout`), then `nimble.New(cfg, nil)`.
  - `"julia"` → `julia.ConfigFromEnv()`, flag overrides (`-model`, `-timeout`), then
    `julia.New(cfg, nil)`.
  - anything else → error `unsupported provider %q (available: nimble, julia)`.
  Passing `nil` as the transport/runner argument selects the adapter's default
  (HTTP / default command runner). The comment marks the `switch` as "the seam for
  additional adapters".
- `main()` dispatches the subcommand: `decide` → `runDecide`, `help`/`-h`/`--help` →
  `usage`, anything else → error + exit 2.

### 6.1 CLI selection flags (`runDecide`)

| Flag | Default | Purpose |
|---|---|---|
| `-provider` | `nimble` | decision provider to use (`nimble`, `julia`) |
| `-base-url` | `""` | Nimble base URL (defaults to `OLLAMA_BASE_URL`) |
| `-model` | `""` | model name (defaults to `NIMBLE_MODEL` / `JULIA_MODEL`) |
| `-file` | `""` | path to a JSON `DecisionRequest` file (multi-question) |
| `-state` | `""` | free-form state/context (single-question form) |
| `-choices` | `""` | comma-separated allowed choices (single-question form) |
| `-decision-id` | `example` | question id for the single-question form |
| `-timeout` | `nimble.DefaultTimeout` (30s) | request timeout |
| `-json` | `false` | print the `DecisionResult` as JSON |

Where a registration symbol is absent from source it is recorded as absent rather
than assumed. Per the source comment, the CLI is "a debugging and testing surface,
not the primary integration mechanism for agentic-sop."

## 7. Configuration and enablement

Each adapter reads configuration from the environment, with defaults applied via
`ConfigFromEnv()` / `WithDefaults()`.

### 7.1 Nimble configuration (`internal/providers/nimble/config.go`)

| Env var | Default | Purpose |
|---|---|---|
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama-compatible backend root |
| `NIMBLE_MODEL` | `nimble` | model name |
| `NIMBLE_TIMEOUT` | `30s` | request timeout (Go duration syntax) |

### 7.2 Julia configuration (`internal/providers/julia/config.go`)

| Env var | Default | Purpose |
|---|---|---|
| `JULIA_MODEL` | `julia` | reported model identifier |
| `JULIA_MODEL_PATH` | (empty) | ONNX model path; forwarded as `JULIA_MODEL_PATH` to the child |
| `JULIA_PYTHON` | `python3` | interpreter for default helper |
| `JULIA_INFERENCE_CMD` | (empty) | override inference command argv (whitespace-split) |
| `JULIA_TIMEOUT` | `30s` | inference timeout |

`.env.example` exists at the repository root as the environment template. The CLI's
`usage()` text repeats these env vars and defaults, matching the `config.go` defaults.

### 7.3 Enablement / availability

- **Nimble** `Available(ctx)`: performs a cheap model-list check (`GET /api/tags`), never an
  inference call. Returns `false` when the backend is unreachable/times out, or when the
  configured model is absent (exact or tag-prefixed match, e.g. `nimble:latest`).
- **Julia** `Available(ctx)`: delegates to the `Runner`. `CommandRunner.Available` is a
  side-effect-free prerequisite check — true only when a command is configured and its
  program resolves on `PATH`, plus (for the built-in helper) the helper script and a real
  model path exist.

## 8. Timeout / cancellation conventions

**Contract level:** every `Provider` method takes a `context.Context`, so context
propagation is part of the contract (`decision/provider.go`).

**Per-adapter behavior:**

- Nimble: `DefaultTimeout = 30s`; the `HTTPTransport` uses an `http.Client{Timeout: timeout}`
  and builds requests with `http.NewRequestWithContext(ctx, ...)`. `isUnavailable` maps
  `context.DeadlineExceeded`, `context.Canceled`, `net.Error.Timeout`, DNS errors, and
  connection-refused/reset/unreachable to the unavailable error kind.
- Julia: `DefaultTimeout = 30s`; `CommandRunner.Run` wraps the call in
  `context.WithTimeout(ctx, timeout)` and uses `exec.CommandContext`. A timeout maps to
  the unavailable kind; undecodable stdout maps to the malformed-response kind; a
  helper-reported `kind: "failure"` maps to the provider-failure kind.
- CLI: `runDecide` derives a `context.WithTimeout(context.Background(), *timeout)` and
  cancels it on return, passing the same context to `Available` and `Decide`. The
  `-timeout` flag is applied to both provider configs.

## 9. Process / HTTP abstractions

- **HTTP (Nimble):** `internal/providers/nimble/transport.go` — `Transport` interface +
  `HTTPTransport` (Go `net/http`). Endpoints observed: the SystemOne inference endpoint
  and the model-list endpoint (`GET /api/tags`). Typed errors include HTTP-status and
  malformed-response errors.
- **Process (Julia):** `internal/providers/julia/runner.go` — `Runner` interface +
  `CommandRunner` (JSON on stdin/stdout). Sentinel errors for unavailable/malformed;
  optional error envelope on stdout for precise classification.
- There is **no shared transport abstraction across adapters**: each adapter owns its own
  transport/runner seam. This is recorded as an observed boundary, not a shared layer.

## 10. Test structure

Test files observed per package:

- `decision/` — `request_test.go`, `result_test.go` (contract validation tests).
- `internal/providers/nimble/` — `config_test.go`, `nimble_test.go`, `translate_test.go`,
  `transport_test.go`, `golden_test.go` (golden fixtures).
- `internal/providers/julia/` — `config_test.go`, `julia_test.go`, `runner_test.go`,
  `translate_test.go`.
- `tests/` — top-level package (`go test` reports
  `ok .../tests`, so it contains at least one test file and may hold shared assets such
  as request fixtures referenced by the CLI usage examples).
- `cmd/sop-decision-adapter/` — no test files (`[no test files]`).

Patterns: tests use stubbed `Transport` / `Runner` seams so adapters can be exercised
without a live Ollama instance or ONNX runtime; `golden_test.go` indicates golden-file
comparison for translation output. This maps to the §2 baseline: every package with
tests reports `ok`, and the CLI package reports `[no test files]`.

## 11. Relationship to agentic-sop

Dependency direction: the `decision` package is a **provider-neutral contract** that
callers (including agentic-sop) depend on. Adapters (`internal/providers/*`) implement
that contract. Source evidence:

- `decision/provider.go` — "Callers such as agentic-sop depend only on this contract,
  never on provider-specific types."
- `decision/request.go` — comment stating execution constraints and safety gates belong
  to agentic-sop, not here.
- `decision/result.go` — comment stating an approval gate is decided by agentic-sop,
  not by this adapter.
- `cmd/sop-decision-adapter/main.go` — the CLI is "a debugging and testing surface, not
  the primary integration mechanism for agentic-sop."

No `agentic-sop` module require line, submodule, or directory was found in this checkout;
the relationship is expressed in source comments describing ownership boundaries, not in
build metadata. The concrete module-level integration (how the agentic-sop repository
imports `sop-decision-adapters`) was **not present in this repository** and is recorded
as UNKNOWN rather than assumed.

## 12. Current-state architecture diagram

```mermaid
flowchart TD
    A[agentic-sop\nowns SOP policy: approval gates,<br/>execution thresholds, safety] -->|depends only on| C[decision package<br/>Provider / DecisionRequest / Question<br/>DecisionResult / Answer / Usage]
    CLI[cmd/sop-decision-adapter<br/>main.go: newProvider() switch<br/>-provider nimble|julia] --> C
    C --> N[internal/providers/nimble<br/>Provider, New(cfg, Transport)]
    C --> J[internal/providers/julia<br/>Provider, New(cfg, Runner)]
    N --> NT[Transport / HTTPTransport<br/>net/http; SystemOne endpoint, GET /api/tags]
    NT --> OL[Ollama-compatible backend -> Nimble]
    J --> JR[Runner / CommandRunner<br/>exec.CommandContext; JSON stdin->stdout]
    JR --> JP[tools/julia/infer.py -> ONNX -> Julia]
    ENV[.env.example / env vars<br/>OLLAMA_BASE_URL, NIMBLE_*<br/>JULIA_*] -.-> N
    ENV -.-> J
    ENV -.-> CLI
```

## 13. Documentation / source divergence

| Topic | Documentation claim | Source (authoritative) | Resolution |
|---|---|---|---|
| CLI path | Documentation/plans refer to `cmd/decision` | Actual entry point is `cmd/sop-decision-adapter/main.go` (package `main`) | Source wins: CLI lives under `cmd/sop-decision-adapter`. |
| Provider selection | (implied registry) | No registry/factory map in `decision` or adapters; selection is a `switch` in `newProvider` inside the CLI | Source wins: selection is construction-time, not registry-based. |
| Shared transport | (implied shared layer) | Each adapter owns its own transport/runner seam; no shared transport | Source wins: transport seams are per-adapter. |
| agentic-sop integration | (described conceptually) | Only contract comments and CLI comments; no build-metadata require/submodule in this repo | Recorded as UNKNOWN; not asserted. |
| Defaults | README/env doc | Defaults in `config.go` (`30s`, `http://localhost:11434`, `nimble`, `julia`, `python3`) | Source wins. |

## 14. Explicit UNKNOWN / not-fully-inspected items

- Concrete toolchain binary version (`go version`) — command not permitted under the
  allowed command set; module/directive recorded from `go.mod` instead.
- Concrete module-level integration/dependency direction with the `agentic-sop`
  repository (no in-repo build metadata).
- Exhaustive contents of `tests/` beyond the fact that the package has tests and reports `ok`.
- `internal/decision/` was empty in the directory listing; no Go sources were observed
  there.

## 15. Change discipline

This report is the only file created or modified by this task. No production code, tests,
or build files were changed. The pre-existing user-owned modification to
`docs/plans/PLAN-Clef-Provider-Integration.md` was preserved untouched, and SOP state
(`.agent-sdlc/`) was not modified.
