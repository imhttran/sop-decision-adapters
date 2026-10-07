# Julia Routing Calibration Report

**Status:** evaluation evidence only. This report does **not** tune the model, does
**not** define routing policy, and does **not** make Julia authoritative for
routing. Probabilities are decision evidence / telemetry, never policy inputs.

## Environment

| Item | Value |
| ---- | ----- |
| Repository | `sop-decision-adapters` |
| Revision | `79569b884c5f38a1a7088ddb5e992ba5c0b90e6a` (branch `main`) |
| Model | `SupersonicLabs/Julia-1` |
| Runtime path | `sop-decision-adapter` → Julia provider → `CommandRunner` → `JULIA_INFERENCE_CMD` → `tools/julia/http_bridge.py` → Julia HTTP service |
| Service | `http://127.0.0.1:8011` (`/health` = 200) |
| Bridge config | `JULIA_URL=http://127.0.0.1:8011`, default `JULIA_HTTP_TIMEOUT=30` |
| Evaluator | `tools/julia/calibrate_routing.py` (uses `http_bridge.build_output`, i.e. the real translation layer) |
| Dataset | `tests/fixtures/julia/routing-calibration.json` |

## Dataset

45 realistic software-engineering task descriptions, 15 per expected class:

| Class | Count | Character |
| ----- | ----- | --------- |
| LOW | 15 | typos, doc-only changes, comment cleanup, tightly bounded renames, formatting, trivial config, one-line fixes |
| MEDIUM | 15 | multi-function handler changes, localized features, moderate test changes, repo-local refactors, bounded query/schema tweaks |
| HIGH | 15 | provider-routing architecture, authorization/security, approval-boundary changes, schema-migration strategy, concurrency, destructive ops, fallback/policy changes, high ambiguity |

Unambiguous baseline cases only. Expected class is fixed by the dataset author
(ground truth), not by the model.

## Variants

| # | Question | Options |
| - | -------- | ------- |
| **A** | "Select the execution model for the described task." | bare `LOW`, `MEDIUM`, `HIGH` |
| **B** | "Select the execution model for the described task." | descriptive (LOW/MEDIUM/HIGH plus a one-line definition each) |
| **C** | "Determine the execution capability required to complete the described task." | descriptive |

## Results

High-confidence misclassification is defined as *incorrect* **and** selected
probability ≥ **0.80** (evaluation-only threshold; **not** production policy).

| Variant | Overall acc | LOW acc | MEDIUM acc | HIGH acc | Avg selected P | Avg margin | High-conf misclassifications |
| ------- | ----------- | ------- | ---------- | -------- | -------------- | ---------- | ---------------------------- |
| A (bare) | **0.356** | 0.000 | **0.333** | 0.733 | 0.789 | 0.610 | 16 / 45 |
| B (descriptive) | 0.333 | 0.000 | 0.200 | 0.800 | 0.851 | 0.721 | 19 / 45 |
| C (descriptive + capability Q) | 0.333 | 0.000 | 0.067 | 0.933 | 0.857 | 0.740 | 23 / 45 |

### Confusion matrices (rows = expected, columns = predicted)

**Variant A (bare choices)**

```text
                 Predicted
             LOW  MED  HIGH
Expected LOW   0    7    8
Expected MED   1    5    9
Expected HIGH  0    4   11
```

**Variant B (descriptive choices)**

```text
                 Predicted
             LOW  MED  HIGH
Expected LOW   0    6    9
Expected MED   0    3   12
Expected HIGH  1    2   12
```

**Variant C (descriptive + capability question)**

```text
                 Predicted
             LOW  MED  HIGH
Expected LOW   0    3   12
Expected MED   0    1   14
Expected HIGH  0    1   14
```

### Predicted class distribution

| Variant | LOW | MEDIUM | HIGH |
| ------- | --- | ------ | ---- |
| A | 1 | 16 | 28 |
| B | 1 | 11 | 33 |
| C | 0 | 5 | 40 |
| _Expected_ | _15_ | _15_ | _15_ |

## Bias investigation

**The model never predicts LOW (0/15 across every variant) and collapses toward
HIGH.**

1. **Severity bias.** Predicted HIGH dominates: 28/45 (A), 33/45 (B), 40/45 (C),
   versus 15 expected. MEDIUM is also over-predicted relative to its true share.
   LOW is essentially never selected (1, 1, 0 times).

2. **Position bias.** Holding the question and labels fixed but reversing the
   option order flips the predictions — the model prefers later positions:

   | Option order | Predicted LOW | MEDIUM | HIGH |
   | ------------ | ------------- | ------ | ---- |
   | `LOW, MEDIUM, HIGH` | 1 | 16 | 28 |
   | `HIGH, MEDIUM, LOW` | 12 | 26 | 7 |

   The `MEDIUM` label is preferred in both orders (16 → 26); the last-position
   label is selected far more often than the first. So the output is driven
   substantially by option position and label severity, not by task semantics.

3. **Severity (SCORE) framing does not help.** Asking the same cases as a SCORE
   question over `[LOW, MEDIUM, HIGH]` yields a mean expected index of **1.885 /
   2.0** (max) — i.e. "almost always maximally severe" — with bucketed accuracy
   0.311 and LOW recall still 0. This confirms the bias is a property of the
   model's severity judgement, not of the choice decoding.

## Major misclassification patterns

- **LOW → MEDIUM/HIGH is systematic (15/15).** No LOW case is ever classified
  LOW. Typical low-risk tasks ("fix a typo", "remove a stale TODO", "delete an
  unused import", "rename a local variable") are routed to MEDIUM or HIGH — often
  with high confidence.
- **MEDIUM → HIGH is the dominant MEDIUM error** (9/12/14 of 15).
- **Descriptive options make it worse**, not better: they push more cases into
  HIGH (B: 33, C: 40) and reduce MEDIUM recall (0.333 → 0.200 → 0.067). The added
  descriptive text appears to read as "this work is substantial", amplifying the
  HIGH bias.
- **High-confidence misclassifications are widespread** (16–23 of 45), so the
  errors are not merely low-confidence noise.

## Recommendation

- **Best-performing variant measured: A (bare `LOW`/`MEDIUM`/`HIGH` options with
  the current question).** It has the highest overall accuracy (0.356) and MEDIUM
  accuracy (0.333), and the least HIGH-collapse. Descriptive options (B, C) are
  measurably worse and should be avoided.
- Even the best variant is far from usable for routing: it has **zero LOW recall**
  and a strong severity/position bias. This is a model-capability limitation, not
  a translation defect (the bridge preserves option order and probabilities
  faithfully; see the Architecture review).
- Before this provider's decisions could inform routing — even in shadow mode —
  the bias must be addressed, e.g. by order-randomised aggregation, few-shot
  class exemplars, a calibrated threshold on the expected-index score, or a
  different model. None of that is in scope here.

## Limitations

- 45 cases, one run per case per variant; the model is stochastic, so absolute
  accuracies carry sampling noise. The qualitative findings (no LOW, HIGH
  collapse, position bias) are large and consistent.
- The dataset is authored, not captured from production history; ground-truth
  labels reflect the author's judgement.
- Probabilities are self-reported by the model and are not probabilities of
  correctness.
- This evaluation covers routing-class agreement only; it does not evaluate the
  provider's plumbing, which is covered by the unit and opt-in live tests.
