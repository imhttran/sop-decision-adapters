# Julia Routing Calibration — Phase 2 (Bias Reduction & Decision Framing)

**Status:** evaluation evidence only. Changes no production policy, no provider
contract, and no Phase 1 dataset. Probabilities and thresholds are decision
evidence / telemetry, never policy inputs. Julia remains non-authoritative.

> Note: the task prompt was truncated mid-Phase 14 ("Preserve Provider Contract").
> Phases 1–13 above are the executed, measured work. Phase 14 (remainder), Phase 15,
> and the Required Final Report were **reconstructed by analogy with the Phase 1
> report** and are clearly labeled below; they introduce no new requirements beyond
> the constraints already given.

## Environment

| Item                      | Value                                                                                                                                          |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Repository / revision     | `sop-decision-adapters` @ `79569b884c5f38a1a7088ddb5e992ba5c0b90e6a` (branch `main`)                                                           |
| Model                     | `SupersonicLabs/Julia-1`                                                                                                                       |
| Runtime path              | `sop-decision-adapter` → Julia provider → `CommandRunner` → `JULIA_INFERENCE_CMD` → `tools/julia/http_bridge.py` → Julia HTTP service          |
| Service                   | `http://127.0.0.1:8011` (`/health` = 200)                                                                                                      |
| Evaluator                 | `tools/julia/calibrate_routing2.py` (drives `http_bridge.build_output`)                                                                        |
| Calibration set           | `tests/fixtures/julia/routing-calibration.json` — **45** (15/15/15), sha256 `b96a970c9b2a42b05f48bd7ff5af0c147a3728c770ff6ea1fe55b2314266f74b` |
| Validation set            | `tests/fixtures/julia/routing-validation.json` — **30** (10/10/10), newly authored                                                             |
| High-confidence threshold | selected P ≥ 0.80 (evaluation-only)                                                                                                            |

The Phase 1 calibration set was **frozen and unmodified** (hash above).

## Phase 3 — Baseline reproduction

Rerunning the best Phase 1 formulation (bare choices, current question,
canonical order) reproduced Phase 1 exactly:

```text
accuracy 0.356 | LOW 0.000 | MEDIUM 0.333 | HIGH 0.733
confusion (rows=expected, cols=LOW/MED/HIGH):
  LOW  [0  7  8]
  MED   [1  5  9]
  HIGH  [0  4 11]
predicted distribution: LOW 1, MEDIUM 16, HIGH 28
high-confidence misclassifications: 16
```

No drift from Phase 1. **Repetition stability was 1.000 (45/45)** — across 5
identical runs every case produced the same class. The model is effectively
**deterministic** under a fixed prompt/order, which is a key result below.

## Experiment A — Option-order randomization (6 permutations × 1 run)

Order changes the prediction for **45/45 cases (stability 0.000)** — a strong
position effect.

| Aggregate rule          | Accuracy  | MacroR | LOW-R | MED-R | HIGH-R | HC errs | Predicted L/M/H |
| ----------------------- | --------- | ------ | ----- | ----- | ------ | ------- | --------------- |
| per-case majority       | **0.444** | 0.444  | 0.200 | 0.200 | 0.933  | 0       | 4 / 3 / 38      |
| mean-probability argmax | 0.378     | 0.378  | 0.200 | 0.000 | 0.933  | 0       | 4 / 0 / 41      |

Averaging over permutations lifts accuracy above baseline (0.444) and recovers a
little LOW, and removes high-confidence errors (averaged probabilities no longer
clear 0.80), but it still collapses toward HIGH and barely represents MEDIUM.

## Experiment B — Repeated runs (canonical order, 5 reps)

Identical to baseline: **accuracy 0.356; LOW 0.000 / MEDIUM 0.333 / HIGH 0.733;
16 high-confidence errors; repeat stability 1.000**. Repetition adds nothing —
the outputs are deterministic, so ensembling reproduces the same bias.

## Experiment C — Order-randomized ensemble (6 perms × R reps)

Because single-run outputs are deterministic, R=3 and R=5 reproduce the R=1
result exactly:

| Configuration   | Accuracy | MacroR | LOW-R | MED-R | HIGH-R | HC errs |
| --------------- | -------- | ------ | ----- | ----- | ------ | ------- |
| 6 × 1 majority  | 0.444    | 0.444  | 0.200 | 0.200 | 0.933  | 0       |
| 6 × 3 majority  | 0.444    | 0.444  | 0.200 | 0.200 | 0.933  | 0       |
| 6 × 5 majority  | 0.444    | 0.444  | 0.200 | 0.200 | 0.933  | 0       |
| 6 × 3 mean-prob | 0.378    | 0.378  | 0.200 | 0.000 | 0.933  | 0       |

## Experiment D — Few-shot exemplars

Fixed exemplars (3 per class, none from the calibration set):

| Framing                        | Accuracy | MacroR | LOW-R     | MED-R | HIGH-R | HC errs | Predicted L/M/H | Prompt growth                   |
| ------------------------------ | -------- | ------ | --------- | ----- | ------ | ------- | --------------- | ------------------------------- |
| few-shot + bare choices        | 0.422    | 0.422  | **0.800** | 0.000 | 0.467  | 8       | 32 / 0 / 13     | +~460 chars (~115 tok) per call |
| few-shot + descriptive choices | 0.333    | 0.333  | 0.000     | 0.000 | 1.000  | 29      | 0 / 0 / 45      | +~600 chars per call            |

Few-shot **recovers LOW strongly (0.80)** but over-corrects — MEDIUM collapses to
0.000 (12/15 MEDIUM → LOW) and 8/15 HIGH → LOW. The few-shot framing shifts the
decision boundary; it does not create MEDIUM discriminability. The descriptive
variant is the worst of all (everything → HIGH).

## Experiment E — SCORE complexity distributions

SCORE = expected zero-based index (Σ i·pᵢ) over `[LOW, MEDIUM, HIGH]` (range 0–2):

| Expected class | mean  | median | min   | q1    | q3    | max   |
| -------------- | ----- | ------ | ----- | ----- | ----- | ----- |
| LOW            | 1.819 | 1.892  | 1.102 | 1.617 | 2.000 | 2.000 |
| MEDIUM         | 1.926 | 2.000  | 1.516 | 1.912 | 2.000 | 2.000 |
| HIGH           | 1.910 | 2.000  | 1.384 | 1.888 | 2.000 | 2.000 |

**Overlap is severe.** All classes sit near the maximum (2.0); LOW is only
slightly lower, and MEDIUM vs HIGH are indistinguishable. There is a weak ordinal
signal separating LOW from {MEDIUM, HIGH} but essentially none separating MEDIUM
from HIGH.

## Phase 9 — Offline threshold search

Transparent grid search maximizing balanced accuracy subject to preserving all
three classes:

- Best thresholds: **T1 = 1.90, T2 = 1.96** → recall LOW 0.533, MEDIUM 0.267,
  HIGH 0.733; balanced accuracy 0.511; overall accuracy 0.511.
- The MEDIUM band (1.90–1.96) is a sliver — effectively a binary LOW-vs-HIGH
  split. This is exploratory only; **there is no production recommendation**.

## Phase 12 — Method comparison

| Method                                       | Acc       | MacroR    | LOW-R     | MED-R | HIGH-R | HC errs | Order stability       | Pred L/M/H |
| -------------------------------------------- | --------- | --------- | --------- | ----- | ------ | ------- | --------------------- | ---------- |
| Baseline A (Phase 1)                         | 0.356     | 0.356     | 0.000     | 0.333 | 0.733  | 16      | det. (100% over reps) | 1/16/28    |
| Random-order majority                        | 0.444     | 0.444     | 0.200     | 0.200 | 0.933  | 0       | 0%                    | 4/3/38     |
| Random-order mean-prob                       | 0.378     | 0.378     | 0.200     | 0.000 | 0.933  | 0       | 0%                    | 4/0/41     |
| Repeated-run majority (5)                    | 0.356     | 0.356     | 0.000     | 0.333 | 0.733  | 16      | 100% (det.)           | 1/16/28    |
| Repeated-run mean-prob (5)                   | 0.356     | 0.356     | 0.000     | 0.333 | 0.733  | 16      | 100% (det.)           | 1/16/28    |
| Order ensemble 3×6 majority                  | 0.444     | 0.444     | 0.200     | 0.200 | 0.933  | 0       | 0%                    | 4/3/38     |
| Order ensemble 5×6 majority                  | 0.444     | 0.444     | 0.200     | 0.200 | 0.933  | 0       | 0%                    | 4/3/38     |
| Few-shot bare                                | 0.422     | 0.422     | 0.800     | 0.000 | 0.467  | 8       | n/a                   | 32/0/13    |
| Few-shot descriptive                         | 0.333     | 0.333     | 0.000     | 0.000 | 1.000  | 29      | n/a                   | 0/0/45     |
| SCORE threshold                              | 0.511     | 0.511     | 0.533     | 0.267 | 0.733  | n/a     | n/a                   | 8/4/33     |
| **Best combined: few-shot + order majority** | **0.444** | **0.444** | **0.600** | 0.000 | 0.733  | **3**   | n/a                   | 21/0/24    |

The best **combined** candidate (few-shot exemplars + order-randomized majority)
recovers LOW (0.60) and cuts high-confidence errors to 3, but MEDIUM remains 0.000.

## Phase 11 — Success-criteria assessment

| Criterion                | Target                 | Best result                               | Met                |
| ------------------------ | ---------------------- | ----------------------------------------- | ------------------ |
| Overall accuracy         | ≥ 0.65                 | 0.444 (order/combined); 0.511 (SCORE)     | ✗                  |
| Macro recall             | ≥ 0.60                 | 0.444                                     | ✗                  |
| LOW recall               | ≥ 0.60                 | 0.800 (few-shot) but MEDIUM 0             | ✗ (not balanced)   |
| HIGH-collapse control    | non-dominant, reported | HIGH still 38–45 of 45 in most methods    | ✗                  |
| Order stability          | ≥ 0.80                 | 0.000 (raw order sensitivity)             | ✗                  |
| High-confidence errors   | ≤ 8                    | 0 (order-avg), 3 (combined), 8 (few-shot) | ✓ (order/combined) |
| Validation-set behaviour | similar                | matched (see below)                       | —                  |

**No candidate satisfies the gates.** The only criteria met are the
high-confidence-error reduction (order averaging) — and that is a side effect of
averaging probabilities, not of improved discrimination.

## Phase 10 — Validation set

Frozen candidates (chosen on the calibration set only) evaluated on the held-out
30-case set:

| Method                    | Acc   | MacroR | LOW-R | MED-R | HIGH-R | HC errs | Pred L/M/H |
| ------------------------- | ----- | ------ | ----- | ----- | ------ | ------- | ---------- |
| Order-random majority     | 0.400 | 0.400  | 0.000 | 0.200 | 1.000  | 0       | 1/3/26     |
| Few-shot + order majority | 0.500 | 0.500  | 0.700 | 0.000 | 0.800  | 1       | 13/0/17    |

Validation **confirms** the calibration findings: LOW can be recovered, MEDIUM
cannot, and HIGH dominates. No method generalizes to a usable signal.

## Phase 13 — Error patterns

- **Severity bias (dominant).** In the baseline every LOW case (15/15) is mapped
  to MEDIUM or HIGH, most at high confidence (`low-04`, `low-05`, `low-11`,
  `low-14` at P = 1.00). SCORE expected-index means are 1.82–1.93 out of 2.0.
- **LOW suppression.** Predicted LOW count never exceeds 4/45 except under
  few-shot, which then also destroys MEDIUM.
- **HIGH over-selection.** Predicted HIGH is 28/45 (baseline) to 38–45/45
  (order-averaged), versus 15 expected.
- **MEDIUM compression.** MEDIUM is absorbed into HIGH in the baseline (9/15 →
  HIGH) and into LOW under few-shot (12/15 → LOW); no formulation leaves a
  distinct MEDIUM.
- **Position bias.** All 6 permutations change the selected class in 100% of
  cases; reversing the bare order (`HIGH,MEDIUM,LOW`) moved the distribution from
  {1/16/28} to {12/26/7}. The model responds strongly to option position as well
  as to label severity.
- **Framing sensitivity.** The same cases yield wildly different distributions
  across prompts (bare 0.356 → descriptive 0.333 → few-shot-bare 0.422 →
  few-shot-descriptive 0.333, with LOW counts 1 → 1 → 32 → 0), so the signal is
  prompt-fragile.
- **Apparent keyword effects.** The only classes reliably recovered are the
  HIGH cases whose text contains architecture/security/fallback/policy keywords,
  suggesting the model keys on surface vocabulary rather than magnitude of work.

## Recommendation

**`HOLD`.** No formulation meets the Phase 11 gates:

- The model is **deterministic** (repeat stability 1.000), so repeated-run
  ensembling cannot help.
- **Option-order randomization** is the only lever that improves accuracy
  (0.356 → 0.444) and removes high-confidence errors, but it cannot create MEDIUM
  discriminability and leaves HIGH dominant.
- **Few-shot** recovers LOW but at the cost of MEDIUM (0.000) — it moves the
  decision boundary, it does not add information.
- **SCORE** carries only weak ordinal signal (LOW slightly separable; MEDIUM/HIGH
  overlap), and its best threshold is a degenerate near-binary split.
- **Validation** confirms all of the above.

Julia-1 is therefore **not yet a useful LOW/MEDIUM/HIGH routing signal**. The
subsystem (transport, bridge, contract, tests) remains technically sound; the
limitation is the model's routing behaviour. Recommended next steps (out of scope
here): try a different model or a classifier fine-tune; test a binary
LOW-vs-rest signal rather than 3-way; or use Julia only for the sub-decisions it
does separate. Julia must **not** become authoritative.

## New/changed artifacts (uncommitted)

- `tools/julia/calibrate_routing2.py` (new) — Phase 2 experiments
- `tests/fixtures/julia/routing-validation.json` (new) — held-out validation set
- `docs/reports/julia-routing-calibration-phase2.md` (this report)
- `tools/julia/http_bridge.py`, `tools/julia/test_http_bridge.py` — Phase 1 hardening from the prior task (still uncommitted)

---

## Phase 14 — Preserve Provider Contract _(reconstructed)_

Confirmed: Phase 2 introduced **no provider-specific behavior** into the
provider-neutral contract.

- `git diff --name-only -- decision/` → **0 files**; `internal/providers/julia/
untouched`.
- Unchanged: `DecisionRequest`, `DecisionResult`, CHOICE/BOOLEAN/SCORE semantics,
  the Julia provider public API, the `Runner` interface, and the bridge's
  stdin/stdout contract.
- All Phase 2 changes are confined to evaluation artifacts: a new experiment
  script (`tools/julia/calibrate_routing2.py`), a new validation dataset
  (`tests/fixtures/julia/routing-validation.json`), and this report.
- `noul`, qtype, `/predict`, HTTP transport, and model-specific translation remain
  behind the Julia provider/runtime boundary.
- Calibration used the existing bridge and CLI only; no routing, approval,
  authorization, fallback, lifecycle, or execution-model default was touched.

## Phase 15 — Readiness Decision _(reconstructed)_

**`HOLD`.** (Definitions: `GO_SHADOW` = ready for non-authoritative shadow
integration; `HOLD` = works but remaining issues must be resolved first;
`NO_GO` = correctness/architecture/reliability defect blocks integration.)

Justification against the Phase 11 gates:

| Gate                   | Target       | Best observed                  | Result |
| ---------------------- | ------------ | ------------------------------ | ------ |
| Overall accuracy       | ≥ 0.65       | 0.511 (SCORE) / 0.444 (choice) | fail   |
| Macro recall           | ≥ 0.60       | 0.444                          | fail   |
| LOW recall (balanced)  | ≥ 0.60       | 0.80 few-shot, but MEDIUM → 0  | fail   |
| HIGH-collapse control  | non-dominant | HIGH 28–45 of 45               | fail   |
| Order stability        | ≥ 0.80       | 0.000                          | fail   |
| High-confidence errors | ≤ 8          | 0–3 (order/combined)           | pass   |
| Validation behaviour   | similar      | confirmed poor                 | —      |

This is `HOLD`, not `NO_GO`: the subsystem (transport, bridge, contract, tests)
has no correctness or reliability defect. The blocker is purely model routing
quality. Julia must **not** become authoritative.

## Required Final Report _(reconstructed, Phase-1 template)_

1. **Branch and HEAD:** `main` @ `79569b884c5f38a1a7088ddb5e992ba5c0b90e6a` (nothing committed).
2. **Working-tree baseline:** only prior-task Julia artifacts uncommitted; no unrelated changes; Julia service healthy.
3. **Calibration dataset frozen:** `routing-calibration.json` sha256 `b96a970c9b2a42b05f48bd7ff5af0c147a3728c770ff6ea1fe55b2314266f74b`, 45 (15/15/15), unmodified.
4. **Baseline reproduction:** 0.356 / LOW 0.000 / MEDIUM 0.333 / HIGH 0.733; 16 high-confidence errors; no drift from Phase 1.
5. **Determinism:** repeat stability 1.000 (45/45) — repeated runs reproduce identical output.
6. **Experiment A (order randomization):** majority 0.444, mean-prob 0.378; order stability 0.000.
7. **Experiment B (repeated runs):** identical to baseline; ensembling adds nothing.
8. **Experiment C (order ensemble 1×6 / 3×6 / 5×6):** 0.444 for all R (determinism).
9. **Experiment D (few-shot):** bare 0.422 (LOW 0.800, MEDIUM 0.000, HC 8); descriptive 0.333 (all HIGH, HC 29).
10. **Experiment E (SCORE):** means LOW 1.819, MEDIUM 1.926, HIGH 1.910 (range 0–2) — severe overlap.
11. **Threshold search:** T1=1.90, T2=1.96 → balanced acc 0.511, MEDIUM band degenerate; exploratory only.
12. **Best combined candidate:** few-shot + order majority → 0.444 (LOW 0.600, MEDIUM 0.000, HIGH 0.733, HC 3).
13. **Overall accuracy:** best 0.511 (SCORE) / 0.444 (choice-based methods).
14. **LOW accuracy/recall:** 0.000 baseline; 0.200 (order-avg); 0.800 (few-shot) but at MEDIUM 0.
15. **MEDIUM accuracy/recall:** best 0.333 (baseline); 0.000 under few-shot and the combined candidate.
16. **HIGH accuracy/recall:** 0.467–1.000 depending on framing.
17. **Confusion matrices:** reported per variant in Phases A–E above.
18. **Predicted distributions:** tabulated in Phase 12.
19. **High-confidence misclassifications:** baseline 16 → order-avg 0, combined 3, few-shot 8, few-shot-descriptive 29.
20. **Stability:** repeat stability 1.000; option-order stability 0.000.
21. **Validation set:** 30 cases (10/10/10); order-rand 0.400, few-shot+order 0.500; MEDIUM 0 in both.
22. **Success criteria:** only the high-confidence-error bound is met; all quality/stability gates fail.
23. **Unresolved risks:** the model cannot represent MEDIUM; systematic position + severity bias; prompt-fragile; HIGH collapse; 45/30-case authored sets with a deterministic (biased) model.
24. **Clef work preserved:** yes — committed in `79569b8`; untouched.
25. **`agentic-sop` untouched:** confirmed (`git -C …/agentic-sop status --short` clean).
26. **Final decision:** **`HOLD`**.
