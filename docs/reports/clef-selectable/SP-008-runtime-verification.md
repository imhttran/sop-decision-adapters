# SP-008 — Live/Runtime Verification

**Status:** verification report (documentation only; no production code changed)
**Scope:** `sop-decision-adapters`, stage SP-008 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline; SP-002 selectability contract; SP-003 hardening decision; SP-004 runtime-readiness decision; SP-005 selection verification; SP-006 selection tests; SP-007 governance/neutrality verification)
**Audience:** SP-010 (judging)

## 0. Purpose

This stage decides whether **live Clef runtime evidence** is required before the
Clef provider can be called **selectable**, and — only if required — performs an
opt-in live verification of the selectable path against a Clef runtime over the
CLEF-002-selected transport (`POST /v1/systemone`, model `clef-flash`). The live
check, if added, must be disabled by default, explicitly opted in, skip cleanly
when the runtime is absent, validate **contract shape only** (choice membership,
confidence bounds, classified failures), and never assert a fixed model verdict.

No live result is fabricated in this document.

## 1. Determination: live evidence is NOT required before selectability

**Determination: live evidence is NOT REQUIRED before selectability.**

**Basis — SP-004 (`docs/reports/clef-selectable/SP-004-runtime-readiness-decision.md`).**
SP-004 §3 records the decision verbatim:

> ### **No.**
>
> Live Clef `/v1/systemone` evidence is **not required before selectability**.
>
> **Rationale, grounded in the three-dimension distinction.**
>
> - **Selection readiness is independent of runtime quality readiness.** …
>   A live backend call contributes nothing to selecting or not selecting Clef.
> - **Implementation readiness is established without live calls.** … the
>   live test file is deliberately decoupled from production behavior
>   (`CLEF_LIVE_TEST` is separate from `CLEF_ENABLED`).
> - **Runtime quality readiness is fail-closed by construction.** A selected,
>   enabled Clef whose backend is unreachable, whose model is absent, or whose
>   response is malformed never yields a successful decision …
> - **SP-003 found no precondition to selection** among M-1/L-1/L-2/I-1 …

SP-004 §4 records the residual obligation that belongs to SP-008 but explicitly
**does not gate selectability**:

> Because the decision is **No** for *selectability*, live evidence is not a
> selectability gate; it remains an explicit obligation for the stage that owns
> runtime verification … SP-008 records `NOT_EXERCISED` / `UNAVAILABLE` with the
> reason. SP-008 does **not** gate selectability; it closes the runtime quality
> readiness dimension.

Because SP-004 determined live evidence is **not required**, **no live test file
is added** by this stage. In particular, the plan's conditional deliverable
`internal/providers/clef/live_selectable_test.go` is **not created**: the stage
instruction is explicit that this file exists "only if required", and SP-004
determined it is not required. Adding it would be an unauthorized change.

## 2. Transport confirmed against CLEF-002, not assumed

The plan describes the CLEF-002-selected transport as `POST /v1/systemone` with
model `clef-flash`. This is confirmed against the CLEF-002 artifact rather than
assumed:

- `docs/reports/clef-provider/CLEF-002-capability-transport.md` §"Primary
  transport" states:
  > **Primary transport: A — Ollama's native `POST /v1/systemone` serving
  > `clef-flash`.**
  and classifies oMLX (`clef-4bit`) `POST /v1/systemone` as HTTP 404
  (**UNSUPPORTED**).
- `docs/specs/clef-provider.md` §: "Uses the CLEF-002-selected transport:
  `POST /v1/systemone` to an Ollama-compatible backend."
- `internal/providers/clef/transport.go` line 63 implements exactly this:
  `http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/v1/systemone", …)`.
- `internal/providers/clef/config.go` `DefaultModel = "clef-flash"` and
  `DefaultBaseURL = "http://localhost:11434"`.

**Confirmed:** selected transport = `POST /v1/systemone`; default model =
`clef-flash`; backend = Ollama-compatible root (default `http://localhost:11434`).
No MLX/oMLX path is wired (`internal/providers/clef/clef.go` package comment), so
it is not a candidate transport for any live check.

## 3. Existing opt-in live check (unchanged)

Although this stage adds no new live test, the repository **already contains** an
opt-in live check on the CLEF-002-selected transport. It is recorded here so the
runtime-verification state is complete and consistent with the tree:

- **Path:** `internal/providers/clef/live_systemone_test.go` (CLEF-006).
- **Opt-in mechanism:** the explicit environment variable `CLEF_LIVE_TEST`
  (accepted values `1`/`true`/`yes`, case-insensitive), via `liveOptedIn()`.
  This opt-in is **separate** from the production `CLEF_ENABLED` gate: setting
  only `CLEF_ENABLED` does not run the live test.
- **Default-disabled:** when `CLEF_LIVE_TEST` is unset the test calls `t.Skipf`
  and never runs a live call.
- **Clean skip without a runtime:** `skipUnlessLive` skips when
  `Provider.Available(ctx)` is false (backend unreachable or configured model
  not served), naming the missing prerequisite; it never fails or hangs for a
  missing runtime.
- **Shape-only assertions, verdict-independent:** over each live outcome it
  asserts only that the returned choice is a member of the request's allowed
  choices (`req.Questions[0].Allows(ans.Choice)`) and that confidence, when
  present, lies within `[0, 1]`. Confidence absence is recorded as unavailable,
  not a failure. It never asserts a specific model verdict.
- **Classified failures / fail-closed:** outcomes are classified via
  `classifyOutcome` into `unavailable`, `malformed_response`, `invalid_request`,
  or `provider_failure`; only `Kind == "success"` outcomes are subject to the
  shape assertions, and malformed/indeterminate responses are normalized to
  `decision.KindMalformedResponse` — never counted as success. The short-timeout
  subtest asserts `decision.ErrUnavailable`, and the invalid-request subtest
  asserts `decision.ErrInvalidRequest`.

This file predates SP-008 and is **not modified** by this stage; it is inspected
and reported here. The plan's `live_selectable_test.go` would have duplicated
this existing coverage under a different name; SP-004's not-required
determination removes the need.

## 4. Live execution state

Recorded precisely, with no fabrication:

- **live Clef `POST /v1/systemone` (`clef-flash`) — NOT_EXERCISED / UNAVAILABLE.**
- **oMLX / `clef-4bit` — NOT_EXERCISED.**

**Reason.** No live Clef-serving Ollama backend was reachable in this
environment. The opt-in live test (`CLEF_LIVE_TEST`) was not run, and the
oMLX (`clef-4bit`) path is not wired into the adapter at all (only the
CLEF-002-selected `/v1/systemone` transport has an implementation). This matches
the state recorded in SP-004 §2 and CLEF-006/CLEF-008
(`docs/reports/clef-provider/CLEF-006-local-runtime-verification.md`,
`docs/reports/clef-provider/CLEF-008-benchmark.md`).

**No-fabrication statement.** The absence of live evidence is recorded as
`NOT_EXERCISED` / `UNAVAILABLE`. It is never described as a measured zero, a
pass, a success, or a backend failure. No latency, success-rate, choice-validity,
or confidence number is reported, because none was observed in this stage.
Because live evidence is not required (SP-004), this state does not block
selectability; it closes the runtime-quality-readiness dimension only.

## 5. Failure remains fail-closed

Fail-closed behavior is unchanged and independent of the live check:

- `internal/providers/clef/clef.go` — `New` never enables Clef; `Decide`
  validates, translates, calls `p.transport.SystemOne`, and normalizes.
  `classifyTransportError` maps connectivity → `KindUnavailable`, undecodable
  response → `KindMalformedResponse`, otherwise → `KindProviderFailure`. A
  failure never becomes a successful result.
- `internal/providers/clef/config.go` — `Enable` is the single gate;
  `WithDefaults` never changes `Enable`; `ConfigFromEnv` reads `CLEF_ENABLED`
  with default false. Transport defaults never enable Clef.
- `internal/providers/clef/transport.go` — non-200 → `httpStatusError`;
  undecodable body → `malformedResponseError`; connectivity errors flow through
  `isUnavailable`.
- `cmd/sop-decision-adapter/main.go` — unknown provider tokens fail closed
  (`unsupported provider %q`); selection never enables Clef.

A malformed or unexpected response shape is a failure, never a silent pass.

## 6. Verification performed in this stage

| Command | Result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go vet ./...`   | exit 0 |
| `go test ./...`  | exit 0 (all packages `ok`) |

With `CLEF_LIVE_TEST` unset (the default), the live check in
`internal/providers/clef/live_systemone_test.go` is skipped cleanly — the
affected package reports `ok` rather than failing or hanging.

No live `POST /v1/systemone` call was issued. No live result is asserted.

## 7. Acceptance-criteria disposition

| Criterion | Disposition | Evidence |
|-----------|-------------|----------|
| Live check (if added) is opt-in and skips cleanly without a runtime | SATISFIED (existing check) | §3; `internal/providers/clef/live_systemone_test.go` (`liveOptedIn`, `skipUnlessLive`) |
| If live evidence is required and unavailable, recorded as NOT_EXERCISED / UNAVAILABLE with reason | SATISFIED (not required; state recorded anyway) | §1, §4 |
| No fabricated results | SATISFIED | §4 no-fabrication statement |
| Failure remains fail-closed | SATISFIED | §5; `clef.go`, `config.go`, `transport.go`, `main.go` |
| Deliverable report exists | SATISFIED | this file |

## 8. No-production-change confirmation

The only repository change introduced by SP-008 is this report file,
`docs/reports/clef-selectable/SP-008-runtime-verification.md`. No Go source,
test file, configuration, policy, lifecycle, approval, authorization, or routing
file is modified. In particular, `internal/providers/clef/live_selectable_test.go`
is **not** created, because SP-004 determined live evidence is not required
before selectability.

## 9. Evidence index

| Artifact | Path |
|----------|------|
| Runtime-readiness decision (basis) | `docs/reports/clef-selectable/SP-004-runtime-readiness-decision.md` |
| CLEF-002 transport selection | `docs/reports/clef-provider/CLEF-002-capability-transport.md` |
| Clef provider spec | `docs/specs/clef-provider.md` |
| Opt-in live check (existing) | `internal/providers/clef/live_systemone_test.go` |
| Clef adapter / transport | `internal/providers/clef/clef.go`, `internal/providers/clef/transport.go`, `internal/providers/clef/config.go` |
| Selection seam | `cmd/sop-decision-adapter/main.go` (`newProvider`) |
| Prior live state | `docs/reports/clef-provider/CLEF-006-local-runtime-verification.md`, `docs/reports/clef-provider/CLEF-008-benchmark.md` |
