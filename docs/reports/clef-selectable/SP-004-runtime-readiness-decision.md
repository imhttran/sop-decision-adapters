# SP-004 — Runtime-Readiness Decision

**Status:** decision (documentation only; no production code changed)
**Scope:** `sop-decision-adapters`, stage SP-004 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline; SP-002 selectability contract; SP-003 hardening decision)
**Audience:** SP-005 (implementation), SP-006 (tests), SP-008 (live-evidence collection), SP-010 (judging)

## 0. Purpose

Decide whether **live Clef evidence** is required before Clef can be called
**selectable**, and record the distinction between the three readiness
dimensions the PRD names:

- **implementation readiness** — the adapter/seam code is correct and tested;
- **selection readiness** — the selection surface is explicit, opt-in, and
  default-preserving;
- **runtime quality readiness** — the live Clef backend serves valid,
  fail-closed decisions.

The report also records the **current live-evidence state** and states whether
that absence blocks selectability, with rationale. No live result is fabricated
in this document.

## 1. The three readiness dimensions, distinguished

These three dimensions are independent: meeting one does not imply the others.
They are stated separately below, each mapped to concrete repository artifacts
inspected at baseline `2b0909b`.

### 1.1 Implementation readiness — the adapter/seam code is correct and tested

**Definition.** The Clef adapter normalizes transport failures into the
provider-neutral `decision` contract, never turns a failure into successful
evidence, and is exercised by unit tests that run offline.

**Observed state: MET (by static inspection and offline tests).**

| Artifact | Evidence |
|----------|----------|
| `internal/providers/clef/clef.go` | `Decide` validates, translates, calls `p.transport.SystemOne`, and normalizes. `classifyTransportError` maps connectivity → `KindUnavailable`, undecodable response → `KindMalformedResponse`, otherwise → `KindProviderFailure`. Failures never become successful results. |
| `internal/providers/clef/transport.go` | `SystemOne` performs `POST {baseURL}/v1/systemone`; non-200 → `httpStatusError`; undecodable body → `malformedResponseError`; connectivity errors flow through `isUnavailable`. `ListModels` performs `GET {baseURL}/api/tags` only (no inference). |
| `internal/providers/clef/config.go` | `Config.Enable` is the single gate; `WithDefaults` never changes `Enable`; `ConfigFromEnv` reads `CLEF_ENABLED`. |
| `internal/providers/clef/systemone.go`, `translate.go` | Request building and response normalization. |
| Tests | `clef_test.go`, `capability_test.go`, `config_test.go`, `governance_test.go`, `translate_test.go`, `transport_test.go` exercise the adapter offline. `cmd/sop-decision-adapter/clef_selection_test.go` covers selection. |

**Limit (recorded, not assumed away):** implementation readiness is asserted
against *static code and offline tests*, not against a live backend. The live
path is exercised only by `internal/providers/clef/live_systemone_test.go`,
which is **disabled by default** (opt-in `CLEF_LIVE_TEST=1`, separate from
`CLEF_ENABLED`) and **SKIPS cleanly** when the local runtime is absent.

### 1.2 Selection readiness — the selection surface is explicit, opt-in, and default-preserving

**Definition (SP-002 contract).** Selection is a single CLI flag, CLI-driven
only, explicit opt-in, default-preserving, and fail-closed on unknown tokens.

**Observed state: MET (by static inspection and selection tests).**

| Clause | Artifact | Observation |
|--------|----------|-------------|
| S1/S2 surface + tokens | `cmd/sop-decision-adapter/main.go` | `-provider` flag default `"nimble"`; tokens `nimble`/`julia`/`clef`/`""`; `availableProviders = "nimble, julia, clef"`. |
| S3 CLI-driven only | same | Environment variables feed *construction* (URL/model/timeout/enable), never selection. |
| S4/S5 opt-in / absence | `newProvider` `clef` branch | Only `-provider clef` selects Clef; `""`/absent → `nimble`. |
| S6 fail closed | `newProvider` default branch | Unknown token → `unsupported provider %q (available: …)`, exit 2. |
| S7 selection ≠ enablement | `internal/providers/clef/clef.go` | `New` never enables; `Available` returns false when `!cfg.Enabled()`; `Decide` returns `KindUnavailable`/`ErrDisabled`. |
| S9 default preservation | `newProvider` | `nimble` remains default; `julia`/`clef` unchanged as non-defaults. |
| S10/S11 rollback | SP-002 §6 | Deselect + unset `CLEF_ENABLED`; no policy change. |
| S16 seam | `decision/provider.go` | `newProvider` returns `decision.Provider` for every token. |

The selection surface therefore matches the contract characterization:
**explicit, opt-in, and default-preserving**. No deviation was located.

### 1.3 Runtime quality readiness — the live Clef backend serves valid, fail-closed decisions

**Definition.** The Clef backend reachable over the CLEF-002-selected
transport (`POST /v1/systemone`) returns responses the adapter can normalize
into valid, fail-closed `decision.DecisionResult` values.

**Observed state: NOT ESTABLISHED. Live evidence absent.**

This dimension cannot be confirmed from source inspection or offline tests.
It requires a running Clef-serving backend. The live test that would exercise
it (`internal/providers/clef/live_systemone_test.go`) is opt-in
(`CLEF_LIVE_TEST=1`) and skips cleanly when the local runtime is absent.

## 2. Current live-evidence state

Recorded verbatim, as required:

- **live Clef `/v1/systemone` — NOT_EXERCISED_LIVE / UNAVAILABLE.**
- **oMLX — NOT_EXERCISED.**

**Reason.** No live Clef-serving Ollama backend was reachable in this
environment; the opt-in live test was not run, and the oMLX (`clef-4bit`) path
is not wired into the adapter at all — only the CLEF-002-selected
`/v1/systemone` transport has an implementation (`internal/providers/clef/clef.go`
package comment states the unverified MLX/oMLX runtime is NOT wired here and
cannot be enabled).

**No-fabrication statement.** The absence of live evidence is recorded as
`NOT_EXERCISED_LIVE` / `UNAVAILABLE` / `NOT_EXERCISED` throughout this document.
It is **never** described as a measured zero, a pass, a success, or a failure
of the backend. No latency, success-rate, choice-validity, or confidence number
is reported, because none was observed.

## 3. Decision: is live evidence required before selectability?

### **No.**

Live Clef `/v1/systemone` evidence is **not required before selectability**.

**Rationale, grounded in the three-dimension distinction.**

- **Selection readiness is independent of runtime quality readiness.** The
  selection surface pinned by SP-002 is a pure CLI resolution: whether
  `-provider clef` resolves to the Clef adapter and whether that adapter stays
  OFF unless `CLEF_ENABLED` is set is fully determined by static code and by
  offline tests (`cmd/sop-decision-adapter/clef_selection_test.go`). A live
  backend call contributes nothing to selecting or not selecting Clef.
- **Implementation readiness is established without live calls.** The
  adapter's correctness — validation, translation, normalization, and
  fail-closed classification — is verified against stub transports offline,
  and the live test file is deliberately decoupled from production behavior
  (`CLEF_LIVE_TEST` is separate from `CLEF_ENABLED`).
- **Runtime quality readiness is fail-closed by construction.** A selected,
  enabled Clef whose backend is unreachable, whose model is absent, or whose
  response is malformed never yields a successful decision:
  `Available` returns `false` (`clef.go`), and `Decide` maps connectivity to
  `KindUnavailable` and undecodable responses to `KindMalformedResponse`
  (`classifyTransportError`). The absence of live evidence degrades to an
  unavailable provider, not to a permissive one. This is the "placeholder"
  contract: the adapter is selectable while its runtime path remains
  fail-closed and clearly non-serving until a real backend answers.
- **SP-003 found no precondition to selection** among M-1/L-1/L-2/I-1; the
  carried-forward items are hardening or not applicable, not live-evidence
  blockers.

**Residual risk accepted.** Proceeding without live evidence accepts that the
wire-level behavior of a real Clef-serving backend (SystemOne request/response
shapes, latency, choice/confidence population) is unverified in this
environment. That risk is bounded because (a) Clef is OFF by default and only
selected by an explicit `-provider clef` **and** `CLEF_ENABLED`, and (b) every
unverified runtime path fails closed to `KindUnavailable` /
`KindMalformedResponse` rather than serving a decision. Selectability concerns
whether Clef can be *chosen and safely off*; runtime quality concerns whether a
live backend serves correct decisions, which is a separate, later concern.

## 4. SP-008 obligation (live evidence)

Because the decision is **No** for *selectability*, live evidence is not a
selectability gate; it remains an explicit obligation for the stage that owns
runtime verification:

**SP-008 must collect the following live evidence before any claim of runtime
quality readiness:**

1. A live `POST /v1/systemone` call against a running Ollama serving
   `clef-flash` (or the configured `CLEF_MODEL`), via the opt-in test path
   `CLEF_LIVE_TEST=1` (`internal/providers/clef/live_systemone_test.go`).
2. The resulting machine-readable artifact (`clef-006-live-results.json`
   under the OS temp dir) transcribed into a report, recording: success count,
   success rate, latency p50/p95, choice-validity count, confidence-availability
   count, timeout behavior, and malformed-response behavior.
3. `TestLiveSystemOneTimeoutBehavior` and `TestLiveSystemOneInvalidRequest`
   confirmed on the live host; and the oMLX path confirmed either still
   **NOT_EXERCISED** or separately verified — never inferred from the
   `/v1/systemone` result.

**Condition/owner.** Owner: the local runtime operator (per the plan's
"Ollama SystemOne decision endpoint serving Clef" capability). Condition: a
reachable Ollama backend serving the configured model. Until then, SP-008
records `NOT_EXERCISED` / `UNAVAILABLE` with the reason. SP-008 does **not**
gate selectability; it closes the runtime quality readiness dimension.

## 5. No-production-change confirmation

The only repository change introduced by SP-004 is this report file,
`docs/reports/clef-selectable/SP-004-runtime-readiness-decision.md`. No Go
source, configuration, policy, lifecycle, approval, authorization, or routing
file is modified. The required validations `go build ./...`, `go vet ./...`,
and `go test ./...` run against the unchanged tree.
