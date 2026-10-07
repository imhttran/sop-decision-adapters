# CLEF-008 — Benchmark and Compare Decision Quality

**Type:** Focused decision-provider benchmark report (non-normative).
**Status:** IMPLEMENTED — the adapter-side benchmark harness and corpus are
implemented and committed in `internal/benchmark/`. The primary **Nimble** cell
was **MEASURED** against the local runtime; the primary **Clef** cell is
**UNAVAILABLE** in this environment; the conditional cells are **NOT_EXERCISED**.
No value is fabricated and no unavailable observation is reported as a measured
zero.
**Scope:** `sop-decision-adapters` decision-provider benchmark only. Not a general
LLM benchmark. No SOP policy change. Benchmark code is off the production
decision path.
**Dependencies:** CLEF-001, CLEF-002, CLEF-003, CLEF-006, CLEF-007.

This report uses four explicit measurement states throughout:

| State             | Meaning                                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------ |
| **IMPLEMENTED**   | Code/artifact exists in the repository.                                                    |
| **MEASURED**      | A value was observed from an executed live run.                                            |
| **DERIVED**       | A value computed adapter-side from measured signals (for example latency percentiles).     |
| **UNAVAILABLE**   | The cell's runtime/prerequisite was absent; no value was measured. Never reported as zero. |
| **NOT_EXERCISED** | The cell's entry condition was not met; the cell did not run.                              |

---

## 0. Purpose, method, and non-goals

CLEF-008 runs a **focused decision-provider benchmark** measuring decision quality
of the Clef provider against ground-truth expected outcomes and, where
semantically comparable, Nimble. It reuses the existing provider-neutral decision
seam (`decision.Provider.Decide`) and the observation vocabulary of the CLEF-007
shadow harness (`internal/shadow`).

Non-goals (explicitly out of scope):

- **Not a general LLM benchmark.** The corpus contains only choice-decision cases
  (allowed-choice sets), not general capability, prompt-suite, or leaderboard
  evaluation.
- **No SOP policy modification.** The benchmark only reads and records; no code
  path writes SOP state or policy, and no result can influence routing/governance.

---

## 1. Deliverables (IMPLEMENTED)

| Deliverable       | Path                                                                       | State                   |
| ----------------- | -------------------------------------------------------------------------- | ----------------------- |
| Benchmark harness | `internal/benchmark/benchmark.go`, `internal/benchmark/cells.go`           | IMPLEMENTED             |
| Benchmark corpus  | `internal/benchmark/corpus.json` (+ loader `internal/benchmark/corpus.go`) | IMPLEMENTED             |
| Harness tests     | `internal/benchmark/benchmark_test.go`                                     | IMPLEMENTED             |
| Benchmark report  | `docs/reports/clef-provider/CLEF-008-benchmark.md`                         | IMPLEMENTED (this file) |

The harness is a library (`benchmark.Run(ctx, corpus, cells)`) rather than a new
CLI subcommand: no benchmark command was added to `cmd/`, to keep the change
minimal and off the production surface. Invocation and reproduction are described
in §10.

---

## 2. Benchmark corpus (IMPLEMENTED)

- **Version:** `clef-bench-v1`
- **SHA-256 (canonical):** `9825f18a871e57c05725d467fee5a2642330b255c00ddfba595ead2f939a81dc`
- **Cases:** 6 (4 labeled with ground truth, 2 unlabeled)
- **Case IDs (stable):** `bench-risk-auth-001`, `bench-approval-schema-001`,
  `bench-risk-docs-002`, `bench-priority-001`, `bench-rollback-001`,
  `bench-escalate-001`

Each case carries a stable `id`, a `state`, a provider-neutral `choice` question
with an allowed-choice set, and an optional labeled ground truth (`expected` +
`ground_truth`). The corpus is embedded in the binary (`go:embed`) and is
therefore versioned with the repository and loadable deterministically.

The loader rejects malformed corpora explicitly: an empty version, an empty case
list, a missing or duplicate case ID, an empty state, a non-choice question, a
labeled case with an empty or out-of-set `expected`, an unlabeled case carrying an
`expected`, and unknown JSON fields are all errors (see §9).

The corpus hash is computed over the canonical JSON encoding, independent of
source whitespace, and is stable across processes (asserted by test).

---

## 3. Harness architecture (IMPLEMENTED)

- `Run(ctx, corpus, cells)` evaluates every cell against the **identical corpus
  and case order** and returns a deterministic `Report`.
- A `Source` is provider-neutral: `Name`, `Available` (non-inference),
  `Comparable`, and `Observe` returning a `shadow.Observation`. Comparability uses
  the CLEF-007 predicate (choice question with a non-empty allowed set).
- `ProviderSource` adapts any `decision.Provider` to `Source`, reaching the
  provider **only** through the `decision.Provider` seam. Provider-specific
  transport stays behind the provider boundary; no provider-specific policy logic
  exists in the harness. `cells.go` is the only provider-aware file and is the
  cell registry, mirroring the existing `newProvider` switch.
- Observation failures are contained: `Observe` is invoked with panic recovery, so
  a misbehaving subject can never crash the harness or reach a primary decision.
- Every metric is an explicit `Stat{Numerator, Denominator}`; a zero denominator
  is rendered "not measurable", never a measured zero. `Sample` and `Excluded`
  counts are preserved, and `Sample + Excluded == Cases` for a measured cell.
- Outcome classification preserves invalid, timeout, malformed, failure, and
  indeterminate outcomes distinctly (labels reuse `shadow.Label`).
- `Report.JSON()` emits deterministic, machine-readable JSON (map keys sorted);
  only measured latency values are non-deterministic.

---

## 4. Execution results

Executed against `DefaultCells()` in this environment on the corpus above.

### 4.1 Primary matrix — Clef (Ollama `/v1/systemone`) — UNAVAILABLE

**State: UNAVAILABLE.** The Clef provider is OFF by default and `CLEF_ENABLED` was
unset in this environment, so `Available` reported false and the cell did not
reach an inference. Missing prerequisite: the local Ollama backend serving
`clef-flash` over `POST /v1/systemone`, with Clef explicitly enabled. **All
metrics are reported as not measurable (denominator 0), not as zero.** Status
reason recorded in the artifact: `provider/runtime not available; live
measurement not exercised`.

| Metric                           | Value          | Sample | State       |
| -------------------------------- | -------------- | ------ | ----------- |
| Choice validity                  | not measurable | 0      | UNAVAILABLE |
| Agreement with expected outcomes | not measurable | 0      | UNAVAILABLE |
| Confidence availability          | not measurable | 0      | UNAVAILABLE |
| Confidence calibration           | not measurable | 0      | UNAVAILABLE |
| Latency p50 / p95                | not measurable | 0      | UNAVAILABLE |
| Timeout rate                     | not measurable | 0      | UNAVAILABLE |
| Failure rate                     | not measurable | 0      | UNAVAILABLE |
| Indeterminate rate               | not measurable | 0      | UNAVAILABLE |

### 4.2 Primary matrix — Nimble — MEASURED

**State: MEASURED.** A local Ollama-compatible backend served the configured
Nimble model, so this cell executed live across all 6 corpus cases. The figures
below are a **single observed local run** and are not an SLA or guarantee.

Numerators and denominators are the harness's explicit `Stat` values.

| Metric                           | Numerator | Denominator | Value | State    |
| -------------------------------- | --------- | ----------- | ----- | -------- |
| Choice validity                  | 6         | 6           | 1.000 | MEASURED |
| Agreement with expected outcomes | 4         | 4           | 1.000 | MEASURED |
| Confidence availability          | 6         | 6           | 1.000 | MEASURED |
| Timeout rate                     | 0         | 6           | 0.000 | MEASURED |
| Failure rate (incl. malformed)   | 0         | 6           | 0.000 | MEASURED |
| Indeterminate rate               | 0         | 6           | 0.000 | MEASURED |

- **Sample:** 6 comparable cases; **Excluded / non-comparable:** 0.
- **Latency (DERIVED adapter-side from measured wall-clock):** p50 = 72.5 ms,
  p95 = 99.2 ms (nearest-rank).
- **Confidence distribution (MEASURED):** `0.0–0.2`: 1, `0.2–0.4`: 0,
  `0.4–0.6`: 0, `0.6–0.8`: 1, `0.8–1.0`: 4.
- **Confidence calibration over labeled cases (MEASURED, bucket → predictions /
  correct):** `0.0–0.2`: 1 / 1; `0.6–0.8`: 1 / 1; `0.8–1.0`: 2 / 2. (Unlabeled
  cases are excluded from calibration.)

Observed per-case choices: auth→`HIGH`, schema→`REJECT`, docs→`LOW`,
priority→`P2`, rollback→`ROLLBACK`, escalate→`ESCALATE`.

### 4.3 Conditional matrix — `mlx-community/clef-4bit` via oMLX — NOT_EXERCISED

**NOT_EXERCISED.** Per CLEF-002, oMLX does not expose a fidelity-preserving Clef
decision transport (`/v1/systemone` returns HTTP 404), so the entry condition
("CLEF-002 proved faithful decision semantics") was not met. No metrics.

### 4.4 Conditional matrix — Clef 8-bit — NOT_EXERCISED

**NOT_EXERCISED.** The conditional quantization/fidelity cell is exercised only
against a measurable primary baseline; no measurable Clef baseline exists in this
environment. No metrics.

---

## 5. Comparison integrity

The primary matrix is **not apples-to-apples in this environment**: only the Nimble
cell was measurable, while the Clef cell is UNAVAILABLE. Therefore **no Clef-vs-
Nimble comparison is claimed.** Both cells were evaluated against the identical
corpus and case order (§3) and with identical acceptance criteria; the limitation
is availability, not methodology. A one-cell result must not be read as a
provider-quality ranking.

---

## 6. Quantization / fidelity impact

**Not measurable.** Neither conditional cell produced measurable results, so no
quantization/fidelity delta is reported. Recorded explicitly rather than inferred.

| Comparison               | State         | Reason                         |
| ------------------------ | ------------- | ------------------------------ |
| oMLX `clef-4bit` vs Clef | NOT_EXERCISED | CLEF-002 prerequisite unmet    |
| Clef 8-bit vs Clef       | NOT_EXERCISED | no measurable primary baseline |

---

## 7. Measurement boundaries and limitations

- **Latency (DERIVED):** wall-clock milliseconds measured around the
  `decision.Provider.Decide` call, capturing provider round-trip; the seam does
  not report latency. Reported as nearest-rank p50/p95.
- **Confidence calibration (MEASURED over labeled cases only):** fixed `[0.0,1.0]`
  bins in 0.2 steps; cases without a confidence value or without ground truth are
  excluded from calibration.
- **Agreement denominator:** ground-truth-labeled cases with a definite, allowed
  choice (agreement + disagreement). Invalid or indeterminate outputs are not
  counted in the agreement denominator; they are reported in their own metrics.
- **Unavailable ≠ zero:** an unavailable cell carries zero-valued stats with a zero
  denominator and is reported as not measurable.
- **Single observed run:** the Nimble values are one local sample and may vary
  run-to-run; the harness measures the observed distribution rather than asserting
  bit-identical output.

---

## 8. Policy-neutrality and containment

- **Off the production path:** no `.go` file outside `internal/benchmark/` imports
  the benchmark harness; this is asserted by
  `TestBenchmarkNotImportedByProductionPath`.
- **No authority:** the harness holds no reference to any SOP policy writer, and no
  result can influence routing, governance, or approval. Failures are contained
  (`TestPanicContained`).
- **Provider-neutral:** the harness core imports only `decision` and `internal/shadow`.
  Provider transport stays behind the `decision.Provider` boundary; `cells.go`
  references providers only through that seam.
- **No production wiring:** no benchmark output is written to SOP state or policy.

---

## 9. Verification checklist (acceptance criteria → evidence)

| Acceptance criterion                                   | State     | Evidence                                                |
| ------------------------------------------------------ | --------- | ------------------------------------------------------- |
| Choice validity measured                               | SATISFIED | `Stat` `choice_valid` numerator/denominator; Nimble 6/6 |
| Agreement with expected outcomes measured              | SATISFIED | `Stat` `agreement`; Nimble 4/4                          |
| Confidence availability measured                       | SATISFIED | `Stat` `confidence_available`; Nimble 6/6               |
| Confidence calibration where ground truth exists       | SATISFIED | `Calibration` bins over labeled cases; Nimble §4.2      |
| Latency measured                                       | SATISFIED | DERIVED p50/p95; Nimble 72.5/99.2 ms                    |
| Timeout rate measured                                  | SATISFIED | `Stat` `timeout`; Nimble 0/6                            |
| Failure rate measured                                  | SATISFIED | `Stat` `failure`; Nimble 0/6                            |
| Indeterminate rate measured                            | SATISFIED | `Stat` `indeterminate`; Nimble 0/6                      |
| Quantization/fidelity impact recorded where measurable | SATISFIED | §6 — "not measurable" marked explicitly                 |
| Not a general LLM benchmark                            | SATISFIED | §0, §2 — choice-decision corpus only                    |
| Does not modify SOP policy                             | SATISFIED | §8 — off-path; no importers; no policy writer           |
| Benchmark harness deliverable present                  | SATISFIED | `internal/benchmark/benchmark.go`, `cells.go`           |
| Benchmark corpus deliverable present                   | SATISFIED | `internal/benchmark/corpus.json`                        |

Harness tests (`internal/benchmark/benchmark_test.go`) additionally prove: corpus
loads deterministically with stable IDs; malformed corpus entries fail explicitly;
every cell uses the same corpus; unavailable providers are reported as unavailable,
not zero; numerator/denominator/sample/excluded counts are preserved; invalid and
indeterminate outcomes are counted consistently; incomparable samples are excluded
explicitly; agreement uses the documented denominator; panics are contained; and
the harness is not imported by the production path.

---

## 10. Reproduction

Build-time dependencies only (Go toolchain). Run the harness against the embedded
corpus and the default matrix:

```go
corpus, _ := benchmark.DefaultCorpus()
report, _ := benchmark.Run(ctx, corpus, benchmark.DefaultCells())
out, _ := report.JSON() // deterministic machine-readable aggregate (latency aside)
```

The Clef cell becomes measurable only with the local Ollama backend serving
`clef-flash` over `POST /v1/systemone` and `CLEF_ENABLED` set; the conditional cells
become measurable only when their entry conditions (§4.3, §4.4) are met. The corpus
hash `9825f18a…` identifies the exact case set evaluated.
