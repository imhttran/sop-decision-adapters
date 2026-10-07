# CLEF-006 — Local Clef Runtime Verification

**Type:** Verification report (non-normative; observed local sample evidence).
**Status:** COMPLETE — the opt-in live integration test is implemented and the
verification methodology is recorded. Live metric values are produced only when
the operator opts in against a real local runtime; when the runtime is absent the
test skips cleanly and this report records the missing prerequisite rather than
fabricated numbers.
**Scope:** `sop-decision-adapters` only. No production wiring change. The Clef
provider remains OFF by default. Artifacts: this report and
`internal/providers/clef/live_systemone_test.go`.
**Dependencies:** CLEF-001 (`CLEF-001-repository-truth.md`), CLEF-002
(`CLEF-002-capability-transport.md`), CLEF-003
(`CLEF-003-adapter-contract-mapping.md`).

---

## 0. Purpose and method

CLEF-006 verifies the already-implemented Clef provider against a **real local
Clef runtime** over the CLEF-002-selected transport:

```text
Ollama
  +
clef-flash
  +
POST /v1/systemone
```

The task name intentionally does not claim that `mlx-community/clef-4bit` /
oMLX has been verified. Clef-002 selected Ollama `/v1/systemone` as the primary
transport and classified oMLX as **UNSUPPORTED** both as a native decision
interface (live HTTP 404 on `/v1/systemone`) and as a faithful translation target
(the decision semantics live in a joint schema head oMLX does not execute). See
§5 below.

The verification is delivered as an **opt-in live integration test**
(`internal/providers/clef/live_systemone_test.go`). It validates **contract shape
rather than fixed model judgment**: it never asserts a specific verdict, only that
the returned choice is one of the request's allowed choices, that confidence (when
present) is within bounds, and that failures are classified using the provider's
existing normalized error kinds.

---

## 1. Runtime configuration exercised

| Setting | Source | Default |
| ------- | ------ | ------- |
| Backend URL | `OLLAMA_BASE_URL` | `http://localhost:11434` |
| Model | `CLEF_MODEL` | `clef-flash` |
| Per-call timeout | `CLEF_TIMEOUT` | `30s` |
| **Live-test opt-in** | `CLEF_LIVE_TEST` | unset (test skipped) |
| Production enable gate | `CLEF_ENABLED` | unset (provider disabled) |

The live-test opt-in (`CLEF_LIVE_TEST`) is **separate from** the production
`CLEF_ENABLED` gate. Setting only `CLEF_ENABLED` does **not** cause the live test
to execute; the live test constructs its own `clef.Provider` with
`Config.Enable=true` only inside the opted-in path. Production defaults are
unchanged: the Clef provider is OFF by default.

The default `go test ./...` (opt-in unset) does not contact any runtime: every
live test calls `t.Skipf` with the reason
`live Clef runtime test disabled by default; set CLEF_LIVE_TEST=1 to opt in`.

---

## 2. Commands

Default (non-opted-in) validation:

```bash
go build ./...
go vet ./...
go test ./...
```

Opted-in live run against a real local runtime:

```bash
CLEF_LIVE_TEST=1 OLLAMA_BASE_URL=http://localhost:11434 CLEF_MODEL=clef-flash \
  go test -run 'TestLiveSystemOne' -v ./internal/providers/clef/
```

---

## 3. Opt-in live integration test

`internal/providers/clef/live_systemone_test.go` provides:

- `TestLiveSystemOneContract` — the repeat-run contract-shape probe.
- `TestLiveSystemOneTimeoutBehavior` — timeout classification.
- `TestLiveSystemOneInvalidRequest` — invalid-request classification (runs
  without a backend).

The harness:

- **is disabled by default** (`CLEF_LIVE_TEST` unset → `t.Skipf`);
- **requires explicit opt-in** (`CLEF_LIVE_TEST` in `1`/`true`/`yes`);
- **skips cleanly when local runtime prerequisites are absent** (reuses the
  provider's `Available`/`ListModels` behaviour via `GET /api/tags` rather than
  issuing an inference probe; skips with a reason naming the unreachable backend
  or unserved model, and never fails);
- **issues exactly `POST /v1/systemone` through the existing clef transport** —
  it introduces no new transport, endpoint, or provider;
- **validates contract shape rather than fixed model judgment**;
- **verifies the returned choice belongs to the allowed choices**;
- **verifies confidence bounds when confidence is provided** (absence is recorded
  as unavailability, not a shape failure).

### 3.1 Contract-shape assertions

For every successful call the test asserts, over `decision.DecisionRequest` /
`decision.DecisionResult` and the existing normalize path:

1. the returned choice is a member of the request's allowed choices
   (`Question.Allows`); a choice outside the allowed set fails the test;
2. when confidence is present it lies within `[0, 1]`; when absent it is recorded
   as unavailable;
3. no fixed expected verdict is asserted — pass/fail depends only on contract
   shape, allowed-choice membership, and confidence bounds.

### 3.2 Outcome classification

The test classifies each call using the provider's existing normalized
`decision.ErrorKind` values:

| Outcome label | Normalized kind | Meaning |
| ------------- | --------------- | ------- |
| `unavailable` | `decision.KindUnavailable` | backend unreachable / timeout |
| `malformed_response` | `decision.KindMalformedResponse` | undecodable or indeterminate response |
| `invalid_request` | `decision.KindInvalidRequest` | malformed or unsupported request |
| `provider_failure` | `decision.KindProviderFailure` | any other backend failure |
| `success` | — | normalized `DecisionResult` returned |

### 3.3 Results artifact

The test writes a machine-readable JSON artifact (path logged via `t.Logf`, also
written to the OS temp dir as `clef-006-live-results.json`) containing: backend
URL, model, timeout, repeat count, sample size, success count, success rate,
latency p50/p95 (ms), choice-valid count, confidence-available count, timeout
behaviour, malformed/indeterminate behaviour, the repeatability statement, and the
per-call outcomes. The artifact's `note` field states that the figures are
observed local sample data and not an SLA or guarantee.

---

## 4. Recorded metrics

Metrics are transcribed from the results artifact produced by
`TestLiveSystemOneContract`. The repeat count is **5** (`liveRepeatCount`) and the
sample size is the number of calls attempted. **All values are observed local
runtime sample data, not an SLA or guarantee.**

| Metric | Source field | Value |
| ------ | ------------ | ----- |
| Request success rate | `success_rate` | see §4.1 |
| Latency p50 | `latency_p50_ms` | see §4.1 |
| Latency p95 | `latency_p95_ms` | see §4.1 |
| Choice validity | `choice_valid_count` / `sample_size` | see §4.1 |
| Confidence availability | `confidence_available_count` / `sample_size` | see §4.1 |
| Timeout behavior | `timeout_behavior` | see §4.2 |
| Malformed/indeterminate behavior | `malformed_behavior` | see §4.3 |
| Repeatability | `repeatability` | see §4.4 |

### 4.1 Observed-run values

> **Prerequisite status: unavailable on this host.** The required local Clef
> runtime (Ollama serving `clef-flash` over `POST /v1/systemone`) was **not
> available** to the governed implementation environment. The missing
> prerequisite is the live local Ollama backend with the `clef-flash` model
> served. The live test therefore **skipped** with the reason
> `local Clef runtime prerequisite absent`. No numeric values are asserted here;
> the operator-run command in §2 produces them. This report does not fabricate
> results.

| Metric | Value | Sample size |
| ------ | ----- | ----------- |
| Request success rate | unavailable (runtime not exercised in this environment) | 0 |
| Latency p50 | unavailable (runtime not exercised in this environment) | 0 |
| Latency p95 | unavailable (runtime not exercised in this environment) | 0 |
| Choice validity | unavailable (runtime not exercised in this environment) | 0 |
| Confidence availability | unavailable (runtime not exercised in this environment) | 0 |

Run the §2 opted-in command to populate these values from the emitted artifact.

### 4.2 Timeout behavior

Exercised by `TestLiveSystemOneTimeoutBehavior`: a deliberately short per-call
timeout (1ns) is applied via `Config.Timeout`. The resulting error is asserted to
satisfy `errors.Is(err, decision.ErrUnavailable)` — i.e. it surfaces as
`decision.KindUnavailable`, never as success and never as an unclassified error.
When the runtime is fast enough that the call nonetheless completes, the subtest
**skips** with an explicit reason rather than failing.

### 4.3 Malformed/indeterminate behavior

Undecodable responses produce a `malformedResponseError` and are normalized to
`decision.KindMalformedResponse`; indeterminate responses (missing answer,
missing boolean probability, missing score, empty choice) are likewise normalized
to `decision.KindMalformedResponse` by `NormalizeSystemOneResponse` and are never
converted into a success-shaped answer. The live test records these outcomes with
label `malformed_response` and never counts them as success.

### 4.4 Repeatability

The repeatability statement is derived from the collected outcomes: the number of
distinct choices observed across the successful calls out of the sample size.
Identical requests have been observed to produce bit-identical single-question
choices in CLEF-002, with documented small run-to-run drift on multi-question
inputs; the live test measures the observed distribution rather than asserting
bit-identical output.

---

## 5. clef-4bit / oMLX

```text
NOT EXERCISED — see CLEF-002
```

CLEF-002 did **not** prove a fidelity-preserving decision transport for
oMLX/clef-4bit: `/v1/systemone` on oMLX returns HTTP 404, oMLX serves the Qwen3.5
backbone via its generic `vlm` engine and does not execute the Clef joint schema
head. The oMLX/clef-4bit path is therefore **not exercised** by CLEF-006.

No claim is made that `clef-4bit` was verified, and no claim is made that oMLX is
implemented in this codebase. The Clef provider implements only the
CLEF-002-selected Ollama `/v1/systemone` transport.

---

## 6. Contract-shape validation statement

Validation is **contract-shape based**, not fixed model judgment. The live test
asserts only: allowed-choice membership, confidence bounds when confidence is
provided, and correct classification of timeout / unavailable /
malformed-indeterminate / invalid-request outcomes. It does not encode any
expected model verdict, and it passes or fails solely on contract shape, allowed-
choice membership, and confidence bounds.

---

## 7. Scope and change discipline

- Added `internal/providers/clef/live_systemone_test.go` (opt-in live integration
  test; disabled by default).
- Added this report.
- No production code path was altered to make the test runnable. The Clef
  provider remains OFF by default; `CLEF_ENABLED` is unchanged.
- Metrics are observed local sample data and are not presented as SLAs or
  guarantees.
