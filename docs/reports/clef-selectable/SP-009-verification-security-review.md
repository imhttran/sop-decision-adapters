# SP-009 — Full Verification and Security Review

**Status:** verification/security-review report (documentation only; no production code changed)
**Scope:** `sop-decision-adapters`, stage SP-009 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline; SP-002 contract; SP-003 hardening; SP-004 runtime-readiness; SP-005 selection; SP-006 selection tests; SP-007 governance/neutrality; SP-008 runtime verification)
**Audience:** SP-010 (judging)

## 0. Purpose

This stage runs the mandated repository-wide Go toolchain gates and performs a
security/robustness review of the selectable-provider **selection surface** and
its **architecture guards**. It is a **read-only** verification stage: the only
in-scope write is this report. No product code is modified. Findings, if any,
are recorded rather than fixed.

## 1. Working-tree basis

- Branch: `main` (ahead of `origin/main` by 1 commit), per `git status --short --branch`.
- Uncommitted working-tree state at review time:
  - ` M cmd/sop-decision-adapter/clef_selection_test.go` (pre-existing, user-owned; reviewed as-is).
  - `?? docs/reports/clef-selectable/SP-004-runtime-readiness-decision.md`
  - `?? docs/reports/clef-selectable/SP-005-selection-implementation.md`
  - `?? docs/reports/clef-selectable/SP-006-selection-tests.md`
  - `?? docs/reports/clef-selectable/SP-007-governance-neutrality-verification.md`
  - `?? docs/reports/clef-selectable/SP-008-runtime-verification.md`
- Pre-existing user-owned changes are **reviewed as-is** and were **not reverted,
  discarded, or "fixed"** by this stage.
- Reviewed revision line: HEAD `2b0909b` plus the working-tree changes above.
- Module: `github.com/imhttran/sop-decision-adapters` (`go.mod`).

## 2. Verification gates (SP-009-S1)

All commands are run from the repository root. Exit codes and outcomes are
recorded verbatim below.

| # | Command | Exit code | Outcome |
|---|---------|-----------|---------|
| 1 | `gofmt -l .` | 0 | PASS |
| 2 | `go vet ./...` | 0 | PASS |
| 3 | `go test -count=1 ./...` | 0 | PASS |
| 4 | `go test -race -count=1 ./...` | 0 | PASS |
| 5 | `go build ./...` | 0 | PASS |
| 6 | `git diff --check` | 0 | PASS |

### 2.1 `gofmt -l .`
- Outcome: **PASS** — no files reported. `gofmt -l` lists files whose formatting
  differs from `gofmt`; an empty list is a pass. Exit code 0.

### 2.2 `go vet ./...`
- Outcome: **PASS** — exit code 0, no diagnostics.

### 2.3 `go test -count=1 ./...`
- Outcome: **PASS** — exit code 0; all packages report `ok`.
- Note: with `CLEF_LIVE_TEST` unset (the default), the opt-in live check in
  `internal/providers/clef/live_systemone_test.go` skips cleanly, so the
  `internal/providers/clef` package reports `ok` rather than failing or hanging.

### 2.4 `go test -race -count=1 ./...`
- Outcome: **PASS** — exit code 0; all packages report `ok` under the race detector.

### 2.5 `go build ./...`
- Outcome: **PASS** — exit code 0.

### 2.6 `git diff --check`
- Outcome: **PASS** — exit code 0; no whitespace errors or conflict markers.

### 2.7 Gate classification

No gate failed. If any gate had failed, it would be recorded here as an explicit
finding with the exact failing command and output. **No required gate failed.**

## 3. Selection-surface security and robustness review (SP-009-S2)

The selection surface is the CLI subcommand `decide` and its provider resolver in
`cmd/sop-decision-adapter/main.go`. The provider seam is `decision.Provider`
plus the per-provider adapters under `internal/providers/`.

### 3.1 Command and argument handling

- **Entry point** — `cmd/sop-decision-adapter/main.go`:
  - `main()` requires at least one argument (`len(os.Args) < 2` → usage + exit 2).
  - Subcommand dispatch: `"decide"` → `runDecide(os.Args[2:])`; `"help"`/`"-h"`/
    `"--help"` → usage on stdout; `default:` → `unknown command %q` on stderr +
    usage + exit 2. Unknown commands **fail closed**.
- **Flag parsing** — `runDecide` uses `flag.NewFlagSet("decide", flag.ContinueOnError)`;
  a parse error returns exit code 2. Flags: `-provider` (default `"nimble"`),
  `-base-url`, `-model`, `-file`, `-state`, `-choices`, `-decision-id` (default
  `"example"`), `-timeout` (default `nimble.DefaultTimeout`), `-json`.
- **Concrete malformed/edge-case scenario — empty `-choices` with no `-file`:**
  `buildRequest` returns `errors.New("-choices is required (or use -file)")`,
  `runDecide` prints `invalid request: ...` and returns exit code 2. An empty or
  whitespace-only `-choices` is filtered by `splitChoices`, which trims and drops
  empty entries, so `","` yields zero choices and hits the same fail-closed path.
- **Concrete malformed/edge-case scenario — extra/unknown flags:** `flag`'s
  `ContinueOnError` mode causes `fs.Parse` to return an error, and `runDecide`
  returns exit code 2 without proceeding to provider selection.
- **Concrete malformed/edge-case scenario — unknown provider token:**
  `newProvider` lowercases and trims the name and switches; the `default` branch
  returns `fmt.Errorf("unsupported provider %q (available: %s)", name, availableProviders)`.
  `runDecide` prints `error: ...` and returns exit code 2. Selection does not
  silently fall back to another provider.
- **Injection/quoting:** arguments are consumed via the `flag` package as typed
  values (`string`, `time.Duration`, `bool`); no argument is interpolated into a
  shell command. The Clef adapter builds an HTTP request with
  `http.NewRequestWithContext`, so no shell interpolation or command injection
  surface exists on the selection path.
- **Request validation:** before selection is *used*, `req.Validate()` is called
  in `runDecide`; an invalid request exits 2.

### 3.2 Secret and environment handling

- **Environment reads** are confined to `ConfigFromEnv` in
  `internal/providers/clef/config.go` and the analogous per-provider config
  helpers (`nimble.ConfigFromEnv`, `julia.ConfigFromEnv`). Clef reads
  `OLLAMA_BASE_URL`, `CLEF_MODEL`, `CLEF_TIMEOUT`, `CLEF_ENABLED`.
- **No secret material is read or echoed.** Clef's configuration surface reads
  only a base URL, a model name, a timeout, and a boolean enable flag. There is no
  API-key/token/credential variable in the Clef config: `Config` struct fields are
  `BaseURL`, `Model`, `Timeout`, `Enable`. There is therefore no credential that
  can leak.
- **Leak surfaces checked:**
  - `printResult` prints only `result.Provider`, `result.Model`, question IDs,
    answer kinds, choice/probability/score/confidence, and token usage. It does
    **not** print configuration, base URLs, or environment values.
  - Error paths (`runDecide`, `Decide`) format `provider.Name()`, `err`, and the
    requested `-provider` token — none of which include secrets.
  - `ConfigFromEnv` does not log; it only assembles a `Config`.
- **Safe defaults:** `Config{}.Enable == false`; `WithDefaults` never changes
  `Enable`; `ConfigFromEnv` reads `CLEF_ENABLED` with a default of `false`. Only
  `CLEF_ENABLED` in `{"1","true","yes"}` (case-insensitive) enables Clef, via
  `envBool`. `envDurationOr` falls back to the default on parse error or a
  non-positive duration, so a malformed `CLEF_TIMEOUT` cannot produce an unsafe
  (e.g., zero/negative) timeout.
- **Default endpoint** is `DefaultBaseURL = "http://localhost:11434"`; the code
  makes no assumption that the server is local (callers can override via
  `OLLAMA_BASE_URL`), and the package comment states this explicitly.

### 3.3 Malformed input, error, and timeout handling

- **Malformed `-file` JSON:** `buildRequest` calls `os.ReadFile` then
  `json.Unmarshal`; a read error is returned directly and a decode error is
  wrapped as `decode %s: %w`. `runDecide` prints `invalid request: ...` and
  returns exit code 2. No partial/untrusted request is dispatched.
- **Malformed/undecodable provider response:** `transport.SystemOne` returns a
  `malformedResponseError` for an undecodable body; `Decide` classifies it via
  `classifyTransportError` → `decision.KindMalformedResponse`, wrapped in a
  `decision.ProviderError{Provider, Kind, Err}`. `runDecide`'s error switch maps a
  generic decision error to exit code 1 (a failure — never a success).
- **Indeterminate response:** `NormalizeSystemOneResponse` returning an error is
  normalized to `decision.KindMalformedResponse`; it is **never** counted as a
  successful decision (`clef.go` `Decide` doc + body).
- **Invalid request:** `req.Validate()` failure and `BuildSystemOneRequest`
  failure both become `decision.KindInvalidRequest`; `runDecide` maps
  `decision.ErrInvalidRequest` to exit code 2.
- **Timeout:** `runDecide` sets `context.WithTimeout(context.Background(), *timeout)`
  and passes `ctx` to `provider.Available`, `provider.Decide`, and the underlying
  `transport.SystemOne` (`http.NewRequestWithContext`). A timeout or connectivity
  failure surfaces as `decision.KindUnavailable` via `isUnavailable`, mapped to a
  `decision.ProviderError{Kind: KindUnavailable}`; `runDecide` maps
  `decision.ErrUnavailable` to exit code 3. An unresponsive endpoint therefore
  fails after the bounded timeout rather than hanging indefinitely.

### 3.4 Fail-closed behavior

- **Disabled provider never serves:** `Provider.Decide` returns
  `providerErr(decision.KindUnavailable, ErrDisabled)` when `!p.cfg.Enabled()`
  (`clef.go`). `Provider.Available` returns `false` when disabled. `New` never
  enables Clef.
- **Selection does not enable:** `newProvider`'s `"clef"` case constructs the
  provider with `clef.ConfigFromEnv()` and explicitly notes "Clef is OFF unless
  explicitly enabled via CLEF_ENABLED; selection constructs the provider but does
  not enable it."
- **Unknown/unavailable provider does not fall through:** `newProvider`'s
  `default` branch returns an error; there is no implicit fallback to `nimble`
  or any other provider. An empty provider name is treated as `nimble` by
  explicit `case "", "nimble":`, which is the documented default, not a silent
  fallback on error.
- **Failures are never successes:** `classifyTransportError` maps connectivity →
  `KindUnavailable`, undecodable → `KindMalformedResponse`, otherwise →
  `KindProviderFailure`. `Decide` returns a non-nil error in every failure branch;
  it never returns a zero-value result as a success.
- **No ambiguous selection proceeds:** an unknown command, unknown flag, empty
  `-choices`, invalid request, or unknown provider each exits non-zero before any
  decision is served.

### 3.5 Ollama SystemOne decision endpoint

- **Status: UNKNOWN — NOT_EXERCISED / UNAVAILABLE.** No live Clef-serving Ollama
  backend was reachable in this environment (consistent with SP-004 §2 and SP-008
  §4). No live probe is required by this stage.
- **Non-gating:** no gate or acceptance criterion in SP-009 depends on the live
  endpoint. Fail-closed behavior was assessed entirely from source/call-path
  inspection and unit-test evidence (`internal/providers/clef/*_test.go`), not
  from a live call. No live result is fabricated. Any optional observational probe
  would be labeled non-gating; none was run.

## 4. Architecture-guard verification (SP-009-S3)

### 4.1 Neutral seam — VERIFIED

- The public seam is the `decision` package: `decision.Provider` (interface) and
  the `decision.DecisionRequest` / `decision.DecisionResult` contract. Provider
  adapters implement that interface.
- Clef-specific wire shapes are **unexported** and confined to
  `internal/providers/clef/systemone.go` (`systemOneRequest`, `systemOneQuestion`,
  `systemOneResponse`, `systemOneAnswer`, `systemOneUsage`); the `clef.go` package
  comment states: "Clef-specific concepts (SystemOne wire shapes, the `noul`
  boolean representation, Ollama request/response shapes) are implementation
  details confined to this package and never appear in the public decision package
  or the provider-neutral wire layer."
- The seam in `newProvider` is provider-neutral: it constructs a `decision.Provider`
  and returns it as the interface type; no consumer branches on concrete provider
  types.
- **Evidence:** `decision` package (public contract),
  `internal/providers/clef/systemone.go`, `internal/providers/clef/clef.go`
  (package comment), `cmd/sop-decision-adapter/main.go` (`newProvider`).
  **Verdict: VERIFIED.**

### 4.2 No policy — VERIFIED

- The seam carries no policy decisions: the Clef adapter "only translates and
  normalizes. It never defines or enforces SOP policy; approval gates, execution
  thresholds, and safety enforcement belong to agentic-sop" (`clef.go` package
  comment).
- Inspecting `clef.go`, `translate.go`, and `config.go`, the only decisions made
  are transport/validation/normalization concerns (build request, classify error,
  normalize response). No approval, threshold, authorization, or lifecycle-policy
  logic is embedded in the seam. The `decision` package exposes data contracts,
  not policy.
- **Evidence:** `internal/providers/clef/clef.go` (package doc + `Decide`),
  `internal/providers/clef/translate.go`, `internal/providers/clef/config.go`, and
  the SP-007 governance/neutrality verification
  (`docs/reports/clef-selectable/SP-007-governance-neutrality-verification.md`).
  **Verdict: VERIFIED.**

### 4.3 No routing coupling — VERIFIED

- Selection does not couple Clef's core to any particular provider's routing or
  transport. `newProvider` returns a `decision.Provider` interface; the consumer
  (`runDecide`) calls only `Name()`, `Available(ctx)`, and `Decide(ctx, req)`
  (`decision.Provider` surface). No concrete provider type or provider-specific
  transport type is referenced by the consumer.
- Transport selection is internal to the Clef package: a `transport` interface with
  a `newHTTPTransport` default, and only the CLEF-002-selected `POST /v1/systemone`
  transport is implemented (`clef.go` package comment: the unverified MLX/oMLX
  `clef-4bit` runtime "is NOT wired here and cannot be enabled").
- The `decision` public package has no import of any provider package, so there is
  no routing coupling from the core outward.
- **Evidence:** `cmd/sop-decision-adapter/main.go` (`newProvider`, `runDecide`),
  `internal/providers/clef/transport.go`, `internal/providers/clef/clef.go`
  (package comment), `decision` package imports. **Verdict: VERIFIED.**

### 4.4 Fail-closed — VERIFIED

- Cross-referenced with §3.4. An unknown provider token, an unknown command/flag,
  an empty `-choices`, an invalid request, a disabled Clef, an unreachable backend,
  a timed-out call, an undecodable response, or an indeterminate response all
  produce a non-zero exit and a non-nil error; none falls through to an unintended
  provider and none is counted as a successful decision.
- **Evidence:** `internal/providers/clef/clef.go` (`Decide`, `Available`,
  `classifyTransportError`), `internal/providers/clef/config.go` (`Enabled`,
  `WithDefaults`, `ConfigFromEnv`), `internal/providers/clef/transport.go`, and
  `cmd/sop-decision-adapter/main.go` (`newProvider`, `runDecide`). Consistent with
  §3.4. **Verdict: VERIFIED.**

## 5. Findings register

| ID | Severity | Area | Description | Status | Recommendation |
|----|----------|------|-------------|--------|----------------|
| — | — | — | No Critical, High, Medium, or Low findings identified during this review. | — | None. |

- **No unresolved Critical/High finding remains.**
- **No required gate failed.**
- Informational (non-finding): the Ollama SystemOne decision endpoint is UNKNOWN
  (see §3.5); this does not gate SP-009 and is not a finding, because no gate or
  criterion depends on it.

## 6. Completion gating statement

- All six required gates were run with exact commands and outcomes (§2); **no
  required gate failed**.
- The selection surface was reviewed for command/argument handling,
  secret/environment handling, malformed/error/timeout handling, and fail-closed
  behavior, each with concrete evidence (§3).
- Each architecture guard — neutral seam, no policy, no routing coupling,
  fail-closed — has an explicit VERIFIED verdict with concrete evidence (§4).
- **No unresolved Critical/High finding remains** (§5).

## 7. No-production-change confirmation

The only repository change introduced by SP-009 is this report file,
`docs/reports/clef-selectable/SP-009-verification-security-review.md`. No Go
source, test file, configuration, policy, lifecycle, approval, authorization, or
routing file is modified by this stage. Pre-existing user-owned working-tree
changes were reviewed as-is and not altered.

## 8. Evidence index

| Artifact | Path |
|----------|------|
| Module definition | `go.mod` |
| CLI / selection seam | `cmd/sop-decision-adapter/main.go` (`main`, `runDecide`, `newProvider`, `buildRequest`, `printResult`) |
| Selection tests | `cmd/sop-decision-adapter/clef_selection_test.go` |
| Public decision contract | `decision/` package |
| Clef adapter / config / transport | `internal/providers/clef/clef.go`, `internal/providers/clef/config.go`, `internal/providers/clef/transport.go` |
| Clef wire types (neutral-seam evidence) | `internal/providers/clef/systemone.go`, `internal/providers/clef/translate.go` |
| Opt-in live check (not run) | `internal/providers/clef/live_systemone_test.go` |
| Prior governance/neutrality verification | `docs/reports/clef-selectable/SP-007-governance-neutrality-verification.md` |
| Prior runtime verification | `docs/reports/clef-selectable/SP-008-runtime-verification.md` |
| CLEF-002 transport selection | `docs/reports/clef-provider/CLEF-002-capability-transport.md` |
