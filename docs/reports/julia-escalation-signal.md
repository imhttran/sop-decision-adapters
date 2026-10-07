# Julia Escalation-Signal Study (Phase 3 follow-up)

**Status:** evaluation evidence only. No production wiring, no provider/contract
change, no threshold in `http_bridge.py`. Julia remains non-authoritative.

## Purpose

Phase 3 (`HOLD_ROUTING`) found that 3-class LOW/MEDIUM/HIGH routing is unreliable,
but that one narrower binary question — **"Does this task require escalation?"** —
separated HIGH-vs-rest and *held on validation*. This follow-up characterizes that
signal on a larger, newly authored set with a calibration/validation split and
several fixed prompt variants.

## Environment

| Item | Value |
| ---- | ----- |
| Revision | branch `main` (post-`7600310`) |
| Model | `SupersonicLabs/Julia-1` |
| Path | adapter → Julia provider → `CommandRunner` → `http_bridge.py` → Julia service `http://127.0.0.1:8011` |
| Evaluator | `tools/julia/escalation_probe.py` |
| Dataset | `tests/fixtures/julia/escalation-signal.json` (binary `escalation`) |

## Dataset

60 newly authored cases, split:

| Split | escalation=true | escalation=false |
| ----- | --------------- | ---------------- |
| Calibration | 20 | 20 |
| Validation | 10 | 10 |

`escalation=true` = architecture, security/authorization, cross-system,
destructive/consequential, or policy/fallback work. `false` = routine, localized
work. Cases are distinct from the routing-calibration and routing-validation sets.

## Questions (fixed, unchanged across the dataset)

| ID | Question | true means |
| -- | -------- | ---------- |
| Q1_escalation | "Does this task require escalation?" | escalation |
| Q2_highest_tier | "Should this task be escalated to the highest reasoning and execution tier?" | escalation |
| Q3_consequence_list | "Does this task involve architecture, security, authorization, cross-system behavior, destructive operations, or significant operational consequences?" | escalation |
| Q4_routine_inverted | "Is this task routine, localized work that does not require escalation?" | routine (inverted) |

## Results

Binary metrics; `T` = threshold on P(escalation). "cal bestT" is chosen on
calibration only; "val @calT" applies that frozen threshold to validation.

| Question | cal @0.5 balAcc / F1 / rec / spec | cal bestT | val @0.5 balAcc / F1 / rec / spec | val @calT balAcc |
| -------- | --------------------------------- | ---------- | --------------------------------- | ---------------- |
| **Q1_escalation** | **0.775 / 0.791 / 0.850 / 0.700** | 0.81 (balAcc 0.800) | **0.650 / 0.696 / 0.800 / 0.500** | 0.600 (overfit) |
| Q2_highest_tier | 0.550 / 0.679 / 0.950 / 0.150 | 0.34 | 0.500 / 0.643 / 0.900 / 0.100 | 0.500 |
| Q3_consequence_list | 0.550 / 0.625 / 0.750 / 0.350 | 0.85 (balAcc 0.675) | 0.450 / 0.522 / 0.600 / 0.300 | 0.600 |
| Q4_routine_inverted | 0.475 / 0.432 / 0.400 / 0.550 | 0.68 | 0.450 / 0.353 / 0.300 / 0.600 | 0.400 |

`P(escalation)` means:

| Question | calibration esc / non | validation esc / non |
| -------- | --------------------- | -------------------- |
| Q1 | 0.826 / 0.388 | 0.664 / 0.477 |
| Q2 | 0.848 / 0.792 | 0.722 / 0.891 |
| Q3 | 0.717 / 0.646 | 0.631 / 0.683 |
| Q4 | 0.396 / 0.474 | 0.256 / 0.509 |

### Q1 threshold sensitivity (calibration and validation)

| T | cal balAcc | cal F1 | val balAcc | val F1 |
| - | ---------- | ------ | ---------- | ------ |
| 0.3 | 0.750 | 0.783 | 0.550 | 0.640 |
| 0.5 | 0.775 | 0.791 | 0.650 | 0.696 |
| 0.6 | 0.775 | 0.791 | **0.700** | **0.700** |
| 0.7 | 0.775 | 0.791 | 0.550 | 0.471 |
| 0.8 | 0.775 | 0.780 | 0.550 | 0.471 |

Calibration has a **broad plateau** (balAcc 0.775 from T=0.5 to 0.8) — not a
knife-edge. Validation peaks at **T = 0.6** (balAcc 0.700, F1 0.700); the
calibration-selected T=0.81 is too high for validation (overfit).

## Findings

1. **Q1 is the usable signal.** "Does this task require escalation?" gives
   calibration balAcc 0.775 / F1 0.791 (recall 0.85) and validation balAcc 0.650
   (0.700 at T=0.6), recall 0.80. It is a real, moderately robust HIGH-vs-rest
   signal — far better than any 3-class method (0.356–0.556) and better than the
   Phase 3 binary HIGH (balAcc 0.55–0.72).
2. **Wording matters and only Q1 works well.** The "highest tier" wording (Q2)
   over-fires (spec 0.15 — everything escalated); the consequence list (Q3) is
   moderate; the polarity-inverted routine question (Q4) fails (balAcc 0.45),
   confirming the **severity bias** (the model raises `P(true)` for consequential
   tasks regardless of question polarity).
3. **Operating point.** Practical operating range T ≈ 0.5–0.6: high recall
   (0.70–0.80 escalation recall) with precision ~0.6–0.7 — a useful *flag* for
   human attention, not an authority.
4. **Calibration↔validation gap.** balAcc drops 0.775 → 0.650–0.700 and the
   calibration-optimal threshold overfits. The signal is real but needs a larger
   validation set and a fixed 0.5–0.6 operating point (not the calibration
   optimum).

## Recommendation

Julia-1's **binary escalation question (Q1) is a usable, non-authoritative
governance signal** — supporting the Phase 3 `HOLD_ROUTING` conclusion and Phase
18 Option B. It is a candidate for a future shadow **flag** ("may need
escalation") using a fixed threshold in 0.5–0.6, subject to validation on a much
larger set. It must **not** drive routing, approval, or authorization, and Julia
must not become authoritative.

## Artifacts (uncommitted)

- `tools/julia/escalation_probe.py` (new)
- `tests/fixtures/julia/escalation-signal.json` (new)
- `docs/reports/julia-escalation-signal.md` (this report)

No change to `decision/`, `internal/providers/julia/`, `http_bridge.py`, the
provider contract, the frozen datasets, or `agentic-sop`. No label leakage: the
model receives only `state` + the fixed question text.
