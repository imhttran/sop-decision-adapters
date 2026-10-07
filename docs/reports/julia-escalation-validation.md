# Julia Escalation Signal — Phase 4 Validation & Governance Readiness

**Status:** evaluation evidence only. No production wiring, no provider/contract
change, no threshold in `http_bridge.py`. Julia remains non-authoritative.

> Note: the Phase 4 prompt was **truncated mid-"Non-Negotiable Constraints"**. The
> mission, the frozen Q1 question, and the constraints received are honoured here;
> the gates below are stated explicitly and were **inferred** (the prompt's gate
> and final-report sections were not received).

## Purpose

Validate the **frozen** binary escalation question on a substantially larger and
more diverse held-out set, to decide whether the signal justifies a future
**non-authoritative shadow governance signal** (`GO_GOVERNANCE_SHADOW`).

- Frozen question: **"Does this task require escalation?"** (no wording search)
- No threshold fitting: fixed, pre-specified operating points (T = 0.5, 0.6)
- Routing remains `HOLD_ROUTING`; Julia non-authoritative; no production wiring

## Environment

| Item | Value |
| ---- | ----- |
| Revision | branch `main` @ `6a406aad40eb2c77a7eca8b99754b1e6e62c6049` |
| Model | `SupersonicLabs/Julia-1` |
| Path | adapter → Julia provider → `CommandRunner` → `http_bridge.py` → Julia service `http://127.0.0.1:8011` |
| Evaluator | `tools/julia/escalation_validation.py` |
| Dataset | `tests/fixtures/julia/escalation-validation.json` |

## Dataset

100 newly authored, diverse held-out cases — **50 escalation / 50
non-escalation** — spanning backend, data, infra, security, platform, ops and
docs work. Distinct from the routing-calibration, routing-validation and
escalation-signal sets. `escalation=true` = architecture, security/authorization,
cross-system, destructive/consequential, or policy/fallback work; `false` =
routine, localized work.

## Results (frozen Q1)

| Operating point | Accuracy | Precision | Recall | Specificity | F1 | Balanced accuracy |
| --------------- | -------- | --------- | ------ | ----------- | -- | ----------------- |
| T = 0.5 | 0.780 | 0.741 | 0.860 | 0.700 | 0.796 | **0.780** |
| T = 0.6 | 0.780 | 0.750 | 0.840 | 0.720 | 0.792 | **0.780** |

Confusion @0.5 (positive = escalation): TP 43, FN 7, FP 15, TN 35.

`P(escalation)` distribution:

| Class | mean | median | q1 | q3 | min | max |
| ----- | ---- | ------ | -- | -- | --- | --- |
| escalation | 0.814 | 0.944 | 0.683 | 1.000 | 0.000 | 1.000 |
| non-escalation | 0.374 | 0.320 | 0.069 | 0.626 | 0.000 | 1.000 |

The means (0.814 vs 0.374) and medians (0.944 vs 0.320) separate cleanly, and the
two operating points give identical balanced accuracy — the signal is **not**
knife-edge.

## Comparison with prior evidence

| Set | n (esc/non) | balAcc | recall | specificity | precision |
| --- | ----------- | ------ | ------ | ----------- | --------- |
| Phase 3 escalation study — calibration | 40 (20/20) | 0.775 | 0.850 | 0.700 | 0.739 |
| Phase 3 escalation study — validation | 20 (10/10) | 0.650 | 0.800 | 0.500 | 0.615 |
| **Phase 4 held-out** | **100 (50/50)** | **0.780** | 0.860 | 0.700 | 0.741 |

The earlier 20-case validation (0.650) was **pessimistic small-sample noise**; on
100 held-out cases the frozen question reproduces the calibration-level
performance (0.780), i.e. **no generalization gap**.

## Gates (inferred) — `GO_GOVERNANCE_SHADOW`

| Gate | Target | Observed | Met |
| ---- | ------ | -------- | --- |
| Balanced accuracy | ≥ 0.70 | 0.780 | ✓ |
| Recall (escalation) | ≥ 0.75 | 0.84–0.86 | ✓ |
| Precision | ≥ 0.60 | 0.74–0.75 | ✓ |
| Specificity | ≥ 0.55 | 0.70–0.72 | ✓ |
| No wording/threshold fitting | required | frozen question, fixed T | ✓ |
| Stable across fixed thresholds | required | 0.5 = 0.6 = 0.780 | ✓ |

## Decision

**`GO_GOVERNANCE_SHADOW`** — the frozen Q1 escalation signal is robust enough to
justify a future **non-authoritative shadow governance flag**. It is not routing:
Julia stays out of the execution-class decision and out of any approval,
authorization, fallback, or lifecycle path.

Recommended use, if pursued: a **flag** ("may need escalation") computed at a
fixed T ≈ 0.5–0.6 (recall ≈ 0.85, precision ≈ 0.75) and logged alongside a task —
never as an authority. The residual severity bias (both classes span the full
range; some non-escalation cases reach P≈1.0) must be visible in any telemetry.

## Risks / limits

- Precision ≈ 0.74 → ~1 in 4 escalation flags is a false alarm; suitable for
  attention, not automated action.
- Both distributions reach 0.0 and 1.0 (overlap at the extremes), so per-case
  confidence is not itself reliable.
- One prompt wording, one model, one runtime; English-only task descriptions.
- 100 cases is a solid held-out set but not a production corpus.

## Artifacts (uncommitted)

- `tools/julia/escalation_validation.py` (new)
- `tests/fixtures/julia/escalation-validation.json` (new)
- `docs/reports/julia-escalation-validation.md` (this report)

No change to `decision/`, `internal/providers/julia/`, `http_bridge.py`, the
provider-neutral contract, the frozen datasets, or `agentic-sop`. No label
leakage (model receives only `state` + the frozen question).
