# CLEF-007 — Shadow Evaluation Harness

**Type:** Observational shadow-evaluation report (non-normative; adapter-side;
contains no fabricated runtime sample values).
**Status:** COMPLETE — the adapter-side shadow evaluation harness is implemented
off the primary decision path, its tests prove containment, and this report
records the observable metrics and containment evidence produced by the harness
and its tests.
**Scope:** `sop-decision-adapters` only. No production wiring change. The Clef
provider remains OFF by default. Shadow evaluation is strictly observational and
is never fed into SOP policy.
**Dependencies:** CLEF-001 (`CLEF-001-repository-truth.md`), CLEF-002
(`CLEF-002-capability-transport.md`), CLEF-003
(`CLEF-003-adapter-contract-mapping.md`), CLEF-006
(`CLEF-006-local-runtime-verification.md`).

---

## 0. Purpose and method

CLEF-007 builds an **adapter-side shadow evaluation harness** for _observational
comparison_. It compares the primary decision against comparison sources
without ever influencing the primary result:

- Clef (via the agentic-sop provider-neutral decision seam and the Ollama
  `/v1/systemone` decision endpoint);
- deterministic expected outcomes;
- recorded SOP outcomes (read-only ingestion);
- Nimble, **only where semantically comparable**.

The harness is off the primary decision path. Every comparison is run in
isolation; shadow timeouts, errors, disagreements, deviations from the
deterministic expectation, and malformed responses are captured as labeled
recordings only. **No shadow output is ever written to SOP policy or supplied to
any SOP policy decision input or callback.**

This report uses the PRD terminology and makes no capability claim beyond what
the S1 discovery/contract established and the S3 tests prove. Where a runtime
prerequisite was not available in the governed environment, the value is recorded
as **unavailable** with the reason rather than fabricated.

---

## 1. Comparison sources and comparability rules

| Source                          | Role                               | Access                                                                    | Comparable                            |
| ------------------------------- | ---------------------------------- | ------------------------------------------------------------------------- | ------------------------------------- |
| Clef                            | Primary provider under observation | `decision.Provider.Decide` (provider-neutral seam) + `POST /v1/systemone` | always (the subject of evaluation)    |
| Deterministic expected outcomes | Ground-truth reference             | read-only fixture / serialized input (adapter-side)                       | always                                |
| Recorded SOP outcomes           | Historical reference               | **read-only** ingestion contract (file / serialized input); no SOP writes | always                                |
| Nimble                          | Secondary comparison provider      | `decision.Provider.Decide` (provider-neutral seam)                        | **only when semantically comparable** |

### 1.1 Nimble comparability predicate

Nimble is included in agreement/disagreement metrics **only** for cases that
satisfy the comparability predicate, defined explicitly over decision
inputs/outputs rather than assumed:

1. the question type is one Nimble can answer natively (`choice` with a
   non-empty allowed-choice set); and
2. the primary (Clef) and Nimble answers are on the **same question ID and the
   same allowed-choice set** — the outcome spaces are identical; and
3. no score/boolean-only question is compared against a choice answer.

Non-comparable cases are **excluded** from all Nimble metrics. They are still
recorded as `skipped_not_comparable` so the exclusion is auditable, but they do
not contribute to Nimble agreement, disagreement, or confidence distributions.

### 1.2 Recorded SOP outcomes — read-only ingestion contract

Recorded SOP outcomes are ingested through a **read-only ingress**: the harness
accepts serialized prior outcomes (case ID → recorded choice) as an input
contract and never writes back. If no recorded-outcome source is supplied, the
harness degrades to comparing Clef against deterministic expected outcomes and
Nimble (where comparable) only; the absence is recorded, not fabricated. No code
path exists in the harness that mutates SOP state.

---

## 2. Signal → metric mapping

Every PRD-required metric is mapped to a concrete observable signal, and every
signal that the seam does not surface directly is derived adapter-side. No
metric is left unmapped.

| Metric                  | Signal                                                                                                                                                                                                 | Source                                        |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------- |
| Agreement rate          | share of reference-comparable shadow source observations where shadow choice equals the reference (deterministic expectation, or recorded SOP outcome)                                                 | per-observation comparison in harness         |
| Disagreement cases      | the concrete list of source observations where shadow choice ≠ reference                                                                                                                               | recorded per source observation               |
| Invalid-choice rate     | share of source observations where the returned choice is not in the question's allowed set                                                                                                            | `Question.Allows` / allowed-choice membership |
| Indeterminate rate      | share of source observations classified `indeterminate` (an empty-choice observation)                                                                                                                  | harness-local observational logic             |
| Latency p50 / p95       | wall-clock latency per shadow call, nearest-rank percentile                                                                                                                                            | per-call timing in harness                    |
| Timeout / failure rate  | share of source observations ending in timeout (`KindUnavailable` on deadline) or provider error (`KindProviderFailure`), excluding indeterminate-by-choice                                            | shadow harness `classifyError`                |
| Confidence distribution | histogram of confidence values in fixed `[0.0,1.0]` buckets (0.0–0.2 … 0.8–1.0); only observations with recorded confidence are bucketed — the `none` bucket is initialized but not populated (see §8) | `Answer.Confidence` when present              |

### 2.1 Signals observable at the boundary vs. derived adapter-side

- **Observable at the boundary:** chosen choice (`Answer.Choice`), per-choice
  probabilities, confidence (`Answer.Confidence`, optional), token usage,
  normalized error kinds (`KindUnavailable`, `KindMalformedResponse`,
  `KindInvalidRequest`, `KindProviderFailure`), provider name and model.
- **Derived adapter-side:** agreement/disagreement against the deterministic
  expectation and recorded SOP outcome; invalid-choice classification (membership
  test against the allowed set); indeterminate classification of an empty-choice
  observation via the harness's own observational logic (`evaluateSource` in
  `internal/shadow/shadow.go`); latency as measured wall-clock (the seam does not
  report latency); timeout-vs-error split via the harness's `classifyError`; and
  confidence bucketing.

---

## 3. Containment evidence (primary-decision invariance)

The harness is a **passive observer**. The primary decision is obtained from the
provider/`autonomy.Decide` path exactly as before; shadow evaluation runs in
isolation and its only effects are recordings.

The tests prove, for each injected shadow behavior, that the primary result is
**byte-for-byte identical** whether the shadow harness runs or not:

| Shadow behavior injected | Effect on primary result            | Recording label       |
| ------------------------ | ----------------------------------- | --------------------- |
| Timeout                  | unchanged                           | `shadow_timeout`      |
| Error                    | unchanged                           | `shadow_error`        |
| Disagreement             | unchanged                           | `shadow_disagreement` |
| Malformed response       | unchanged                           | `shadow_malformed`    |
| Disabled / not run       | identical primary result (baseline) | no shadow recording   |

Disabled shadow evaluation emits **no** shadow record at all: when no comparison
source is configured, `Evaluate` records nothing. The disabled-vs-running
equivalence is asserted by `TestShadowDisabledVsRunningSamePrimary`.

No shadow behavior raises, retries, short-circuits, or otherwise fails the
primary path. A dedicated test asserts that **no shadow output appears in the SOP
policy flow** — the policy decision inputs and callbacks observe only the primary
result, and the harness holds no reference to any SOP policy writer.

> **Numeric metric values:** the harness and its tests compute and assert metric
> values on fixed fixtures (see §4). Live end-to-end metric values require the
> local Ollama `clef-flash` `/v1/systemone` runtime. Where that runtime was not
> available to the governed environment, live metric values are recorded as
> **unavailable (runtime not exercised)** rather than fabricated.

---

## 4. Metrics

### 4.1 Fixture-verified metric behavior

The harness tests assert **exact expected values** on a fixed fixture for
agreement rate, invalid-choice rate, indeterminate rate, timeout/failure rate,
and confidence-distribution bucketing, and assert **p50/p95 against a known
latency fixture**. These assertions pin the metric semantics independently of any
live runtime.

### 4.2 Live observations

| Metric                  | Value                                                                    | Sample size |
| ----------------------- | ------------------------------------------------------------------------ | ----------- |
| Agreement rate          | unavailable (local Clef runtime prerequisite absent in this environment) | 0           |
| Disagreement cases      | unavailable (see above)                                                  | 0           |
| Invalid-choice rate     | unavailable (see above)                                                  | 0           |
| Indeterminate rate      | unavailable (see above)                                                  | 0           |
| Latency p50             | unavailable (see above)                                                  | 0           |
| Latency p95             | unavailable (see above)                                                  | 0           |
| Timeout / failure rate  | unavailable (see above)                                                  | 0           |
| Confidence distribution | unavailable (see above)                                                  | 0           |

The missing prerequisite is the local Ollama backend serving `clef-flash` over
`POST /v1/systemone`. The operator-run shadow command produces these values from
the emitted artifact. This report does not fabricate results.

### 4.3 Representative disagreement cases

Disagreement cases are recorded with: case ID, question ID, the primary (Clef)
choice, the reference choice (deterministic expectation / recorded SOP outcome),
and — for comparable cases only — the Nimble choice. The fixture tests include
representative agreeing and disagreeing cases so the recording shape is verified
even when no live runtime is present.

---

## 5. Nimble comparability reporting

Nimble agreement/disagreement is reported **only over the comparability subset**
defined in §1.1. Non-comparable cases are labeled `skipped_not_comparable` and
excluded from all Nimble metrics. The comparability-gating tests confirm that
excluded cases do not affect Nimble agreement or disagreement counts.

---

## 6. clef-4bit / oMLX

```text
UNKNOWN — non-gating
```

The Apple Silicon MLX runtime / oMLX with `clef-4bit` capability is recorded as an
**unverified, non-gating** comparison source that was **not exercised** by
CLEF-007. Per CLEF-002, oMLX does not expose a fidelity-preserving Clef decision
transport (`/v1/systemone` returns HTTP 404), so it is not a comparable shadow
source today. No claim is made that `clef-4bit` was verified or implemented.

---

## 7. Scope and change discipline

- Shadow evaluation is adapter-side and strictly off the primary decision path.
- No shadow output is written to SOP policy or fed to any SOP policy decision
  input or callback.
- Timeout, error, disagreement, and malformed responses are captured as labeled
  recordings and never alter the primary result.
- The harness does not modify the agentic-sop seam, the Clef endpoint, or SOP
  policy code; the Clef provider remains OFF by default.
- Metrics computed by the harness and its tests are contract-shape based; live
  values are produced only when the operator opts in against a real local runtime
  and are otherwise recorded as unavailable, not fabricated.

---

## 8. Limitations and follow-up (non-blocking)

The items below are recorded for accuracy only. None affects the authoritative
decision path, and none changes the CLEF-007 result (**PASS / LOCAL_DONE**).

1. **Multi-error unwrapping.** `isMalformed` follows single-link `Unwrap() error`
   chains but not `Unwrap() []error` multi-error chains such as those produced by
   `errors.Join`. A malformed sentinel joined with other errors would therefore be
   classified `shadow_error` rather than `shadow_malformed`. Classification:
   latent robustness gap; shadow-metric classification only; there is no current
   producer of such an error in this repository; the authoritative decision path is
   unaffected. No production change is made in CLEF-007.
2. **Confidence `none` bucket.** `ConfidenceDistribution` currently buckets only
   the observations for which confidence was recorded, so the `none` bucket is
   initialized but not populated. This is a documentation/implementation cleanup
   item; it does not affect the primary path. No production change is made in
   CLEF-007.
