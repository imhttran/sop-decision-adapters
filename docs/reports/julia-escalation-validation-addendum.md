# Julia Escalation Signal — Phase 4 Completion Addendum

**Status:** evaluation evidence only. No production wiring, no provider/contract
change, no threshold in `http_bridge.py`. Julia remains non-authoritative.

> Note: the addendum prompt was **truncated inside Part 15** (the gate list).
> Parts 1–15 are reported here; if a final-report template followed, it was not
> received. The decision below is derived from the gates that *were* specified.

**This addendum supersedes the provisional `GO_GOVERNANCE_SHADOW` from the
100-case study. The final decision is `HOLD_GOVERNANCE`.**

## Part 1 — Preserved evidence

| Item | Value |
| ---- | ----- |
| Revision | branch `main` @ `6a406aad40eb2c77a7eca8b99754b1e6e62c6049` |
| Existing 100-case dataset sha256 (before expansion) | `b1bea23f6accaf63e92ce5fc649617cc448e0543d61e8c27b2abe92c6dcfe844` |
| Prior 100-case results | reused from cache (not re-run): 100 rows, 50/50 |
| Frozen question | "Does this task require escalation?" (unchanged) |
| Operating points | T=0.50 (primary readiness), T=0.60 (secondary) |

## Part 2 — Expansion

To satisfy the coverage requirements (Parts 3–5) the original 100 cases were
extended by **40 new cases** (not 20), because the original set carried none of
the hard-negative / hard-positive / minimal-pair metadata that Parts 3–5 require.

| Metric | Value |
| ------ | ----- |
| Total | **140** |
| ESCALATE | 70 |
| DO_NOT_ESCALATE | 70 |
| New cases | 40 (20 minimal pairs) |
| Original 100 | **unchanged** (no relabels, no rewrites) |
| New dataset sha256 | `cd80b825047d6f7ee38a44bb5658905912f3c9c2f8feed6a4b8c93ba0fb50d4f` |

## Part 3 — Hard negatives

20 hard negatives (28.6% of the 70 non-escalation cases) — bounded tasks with
alarming vocabulary (authorization, production migration, credential, security,
destructive path, lifecycle, cache invalidation, precedence, backwards-compat, …).

| Metric | Value | Gate | Result |
| ------ | ----- | ---- | ------ |
| Hard-negative count | 20 | ≥ 15 | ✓ |
| Correctly rejected | 7 | | |
| False escalations | 13 | | |
| **Hard-negative specificity** | **0.350** | **≥ 0.65** | **FAIL** |

The model escalates most bounded-but-keyworded tasks (e.g. "Add tests covering
the existing authorization decision" → P=0.954; "Document the production schema
migration procedure" → P=1.0).

## Part 4 — Hard positives

20 hard positives (28.6% of the 70 escalation cases) — consequential work phrased
without obvious severity vocabulary (ordering guarantees, retry/idempotency,
state transitions, invalidation, ownership, precedence, recovery semantics, …).

| Metric | Value | Gate | Result |
| ------ | ----- | ---- | ------ |
| Hard-positive count | 20 | ≥ 15 | ✓ |
| Correctly escalated | 11 | | |
| Missed escalations | 9 | | |
| **Hard-positive recall** | **0.550** | **≥ 0.70** | **FAIL** |

The model misses consequential-but-plainly-worded work (e.g. "Change retry
behavior for non-idempotent requests" → P=0.409; "Change the precedence of
configuration sources" → P=0.056; "Change what the security token asserts to
downstream services" → P=0.0).

## Part 5 — Minimal pairs

20 minimal pairs (A = bounded / DO_NOT_ESCALATE, B = consequential / ESCALATE,
differing in one escalation property).

| Metric | Value | Gate | Result |
| ------ | ----- | ---- | ------ |
| Pair count | 20 | ≥ 20 | ✓ |
| Both correct | 3 | | |
| One correct | 12 | | |
| Both wrong | 5 | | |
| **Both-correct pair rate** | **0.150** | **≥ 0.60** | **FAIL** |
| Escalation-side accuracy | 0.550 | | |
| Non-escalation-side accuracy | 0.350 | | |

## Parts 6–7 — Combined 140-case metrics

New cases were run with the frozen contract; the 100 cached outputs were reused.

| Threshold | Acc | BalAcc | Prec | Recall | Spec | F1 | FPR | FNR | TP | TN | FP | FN |
| --------- | --- | ------ | ---- | ------ | ---- | -- | --- | --- | -- | -- | -- | -- |
| **T=0.50** | 0.686 | **0.686** | 0.659 | 0.771 | 0.600 | 0.711 | 0.400 | 0.229 | 54 | 42 | 28 | 16 |
| T=0.60 | 0.686 | 0.686 | 0.667 | 0.743 | 0.629 | 0.703 | 0.371 | 0.257 | 52 | 44 | 26 | 18 |

## Part 8 — Probability separation

| Class | mean | median | sd | q1 | q3 | min | max |
| ----- | ---- | ------ | -- | -- | -- | --- | --- |
| ESCALATE | 0.738 | 0.916 | 0.325 | 0.559 | 1.000 | 0.000 | 1.000 |
| DO_NOT_ESCALATE | 0.449 | 0.427 | 0.347 | 0.125 | 0.788 | 0.000 | 1.000 |

Mean separation 0.289; median separation 0.489. Separation **survives but is
materially smaller** than the original 100-case set (means 0.814 vs 0.374), and
each class still spans the full 0.0–1.0 range.

## Part 9 — False-negative analysis (16 FN)

Categories (overlapping): security/audit (esc-004, mp05-b), architecture/consistency
(esc-020, esc-025), concurrency/ordering (esc-044, mp09-b), retry/idempotency
(mp03-b, mp15-b → not both FN), destructive/cleanup ordering (mp06-b),
configuration precedence (mp11-b, mp20-b), state/status semantics (mp18-b),
batch/recovery roll-forward (mp17-b), credential handling (mp04-b),
lifecycle/financial (esc-040, esc-041).

**Systematic miss pattern:** consequential work phrased *without* severity
vocabulary is under-escalated (9/20 hard positives missed; P often < 0.30).
This is a **repeated miss in high-consequence categories** (concurrency,
idempotency, recovery, precedence) — the addendum states such a pattern must
block `GO_GOVERNANCE_SHADOW`, and it does.

## Part 10 — False-positive analysis (28 FP)

- **13 of 20 hard negatives** were escalated — driven by vocabulary
  (authorization, production/migration, credential, security, destructive,
  backwards-compatibility, batch/precedence). Classic **keyword/severity bias**.
- **15 original "easy" negatives** were also escalated at high confidence,
  including trivially safe edits ("Add a blank line between two functions" →
  0.921; "Move a constant to the file that uses it" → 1.0; "Fix an incorrect
  example in the README" → 0.945).

**Conclusion:** Julia responds substantially to **surface vocabulary and generic
severity**, not reliably to task semantics — the same bias seen in Phases 1–3.

## Part 11 — Threshold sensitivity

| T | BalAcc | Prec | Recall | Spec | F1 |
| - | ------ | ---- | ------ | ---- | -- |
| 0.40 | 0.650 | 0.613 | 0.814 | 0.486 | 0.699 |
| 0.45 | 0.664 | 0.629 | 0.800 | 0.529 | 0.704 |
| **0.50** | **0.686** | 0.659 | 0.771 | 0.600 | 0.711 |
| 0.55 | 0.686 | 0.662 | 0.757 | 0.614 | 0.707 |
| 0.60 | 0.686 | 0.667 | 0.743 | 0.629 | 0.703 |
| 0.65 | 0.671 | 0.667 | 0.686 | 0.657 | 0.676 |
| 0.70 | 0.679 | 0.687 | 0.657 | 0.700 | 0.672 |

Broad and stable (BalAcc 0.650–0.686, plateau at 0.5–0.6) — **not knife-edge**. No
threshold is selected from this; readiness uses T=0.50.

## Part 12 — Operational cost

`cost = FN·c_FN + FP·c_FP` (evaluation-only):

| T | Policy A (FN=5, FP=1) | Policy B (FN=2, FP=1) |
| - | --------------------- | --------------------- |
| 0.40 | 101 | 62 |
| **0.50** | **108** | **60** |
| 0.60 | 116 | 62 |
| 0.70 | 141 | 69 |

Under recall-oriented Policy A, lower thresholds are cheaper (fewer misses);
under balanced Policy B, T=0.50 is near-optimal. T=0.50 is operationally
reasonable under both — but cost optima do not repair the hard-case gates.

## Part 13 — Baselines (140-case balanced set)

| Baseline | Acc | BalAcc | Recall | Spec | F1 |
| -------- | --- | ------ | ------ | ---- | -- |
| Always ESCALATE | 0.500 | 0.500 | 1.000 | 0.000 | 0.667 |
| Never ESCALATE | 0.500 | 0.500 | 0.000 | 1.000 | — |
| Balanced random (seed 12345) | 0.493 | 0.493 | 0.471 | 0.514 | 0.482 |
| Keyword (security/authorization/production/migration/architecture/destructive/credential) | 0.521 | 0.521 | 0.129 | 0.914 | 0.212 |
| **Julia Q1 @0.50** | **0.686** | **0.686** | 0.771 | 0.600 | 0.711 |

**Does Julia add value beyond keyword matching?** On **overall** balance, yes
(BalAcc 0.686 vs 0.521). But the keyword baseline is *more specific on hard
negatives* (0.70 vs Julia's 0.35) — Julia is **worse** at rejecting bounded
keyword-bearing tasks. So Julia's advantage is largely recall, bought with a
keyword/severity over-fire that the hard negatives expose.

## Part 14 — Reproducibility

20 representative cases (escalation, non-escalation, hard positives, hard
negatives, pairs) re-run: **20/20 exact repeats, 0 changed results, max |ΔP| = 0.0**.
Confirmed deterministic; no probability drift.

## Part 15 — Final readiness gates (T=0.50)

| Addendum gate | Requirement | Observed | Result |
| ------------- | ----------- | -------- | ------ |
| Balanced accuracy | ≥ 0.70 | 0.686 | **FAIL** |
| F1 | ≥ 0.70 | 0.711 | pass |
| Recall | ≥ 0.75 | 0.771 | pass |
| Specificity | ≥ 0.60 | 0.600 | pass (borderline) |
| Hard-negative specificity (§3) | ≥ 0.65 | 0.350 | **FAIL** |
| Hard-positive recall (§4) | ≥ 0.70 | 0.550 | **FAIL** |
| Minimal-pair both-correct rate (§5) | ≥ 0.60 | 0.150 | **FAIL** |
| No catastrophic FN pattern (§9) | required | systematic plain-phrasing misses | **FAIL** |

## Decision

**`HOLD_GOVERNANCE`.**

The provisional `GO_GOVERNANCE_SHADOW` (from the 100-case mostly-easy set,
BalAcc 0.780) **does not survive** the addendum's hard cases and minimal pairs.
Expanding to 140 cases with 20 hard negatives, 20 hard positives and 20 minimal
pairs drops balanced accuracy to 0.686 and fails four gates: hard-negative
specificity (0.35), hard-positive recall (0.55), minimal-pair both-correct rate
(0.15), and the no-catastrophic-FN requirement. Julia's escalation judgement is
dominated by **severity/keyword surface signals** rather than task semantics.

The signal is deterministic and threshold-stable, and it does carry *some*
semantic value (BalAcc 0.686 > random 0.493 and > keyword 0.521), but it is **not
robust enough** to justify even a non-authoritative shadow governance flag.

## Artifacts (uncommitted)

- `tests/fixtures/julia/escalation-validation.json` — expanded to 140 (100 unchanged)
- `tools/julia/escalation_validation2.py` — combined/expanded evaluation
- `docs/reports/julia-escalation-validation-addendum.md` — this report

No change to `decision/`, `internal/providers/julia/`, `http_bridge.py`, the
provider-neutral contract, the frozen routing datasets, or `agentic-sop`. No label
leakage (model receives only `state` + the frozen question).
