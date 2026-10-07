# Julia Routing Calibration — Phase 3 (Binary & Hierarchical Evaluation)

**Status:** evaluation evidence only. No production policy, no provider/contract
change, no threshold in `http_bridge.py`, no change to the frozen datasets.

Phase 1–8 were reported previously; this document completes Phases 9–22.

## Environment

| Item                  | Value                                                                                                                                       |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Repository / revision | `sop-decision-adapters` @ `4a0d65dbdff1e4373dfe92d9dfc8bfdf72c2970b` (branch `main`, clean)                                                 |
| Model                 | `SupersonicLabs/Julia-1`                                                                                                                    |
| Runtime path          | adapter → Julia provider → `CommandRunner` → `JULIA_INFERENCE_CMD` → `tools/julia/http_bridge.py` → Julia service (`http://127.0.0.1:8011`) |
| Evaluator             | `tools/julia/calibrate_routing3.py`                                                                                                         |
| Calibration set       | `routing-calibration.json` sha256 `b96a970c9b2a42b05f48bd7ff5af0c147a3728c770ff6ea1fe55b2314266f74b` (45; 15/15/15) — frozen                |
| Validation set        | `routing-validation.json` sha256 `d0cbd40e67dbb4ca1e3896c2659f04ee860b9471950fd245c467a115c6cbc2b6` (30; 10/10/10)                          |
| Binary channel        | Julia native boolean (NOUL); bridge maps `P(true)=probabilities[1]`                                                                         |

## Phases 2–8 (recap)

Binaries: **A_low** ("Is this task simple, localized, low-risk, and suitable for
the LOW execution class?"), **B_high** ("…require the HIGH execution class?"),
**C_escalation** (architecture/security/… escalation wording).

| Binary       | @0.5 acc | prec  | rec   | spec  | F1    | balAcc    | best T | balAcc@T  |
| ------------ | -------- | ----- | ----- | ----- | ----- | --------- | ------ | --------- |
| A_low        | 0.711    | 0.750 | 0.200 | 0.967 | 0.316 | 0.583     | 0.09   | 0.617     |
| B_high       | 0.667    | 0.500 | 0.667 | 0.667 | 0.571 | **0.667** | 0.65   | **0.717** |
| C_escalation | 0.622    | 0.469 | 1.000 | 0.433 | 0.638 | 0.717     | 0.49   | 0.717     |

`P(true)` by expected class (means): A_low LOW 0.227 / MED 0.094 / HIGH 0.125;
B_high **LOW 0.304 / MED 0.398 / HIGH 0.662** (ordinal); C_escalation 0.607 /
0.640 / 0.864. B_high is the only ordinal signal; A_low almost never fires
(4/45 true). Hierarchies: H1(default) 0.489, H1(searched) 0.511, H2(default)
0.467, **H2(searched) 0.556** (macroR 0.556; pred 10/18/17 — the first
non-collapsing method).

## Phase 9 — Independent dual-boundary

`LOW=true, HIGH=false → LOW`; `false,true → HIGH`; `false,false → MEDIUM`;
`true,true → CONFLICT` (never silently resolved).

| Thresholds           | low_only | high_only | neither | both (conflict) | conflict % |
| -------------------- | -------- | --------- | ------- | --------------- | ---------- |
| 0.5 / 0.5            | 1        | 17        | 24      | 3               | **6.7%**   |
| searched 0.09 / 0.65 | 10       | 7         | 18      | 10              | **22.2%**  |

| Resolution @0.5        | Acc   |     | @searched              | Acc                  |
| ---------------------- | ----- | --- | ---------------------- | -------------------- |
| excluding conflicts    | 0.476 |     | excluding conflicts    | 0.600 (macroR 0.604) |
| conflicts → LOW-first  | 0.489 |     | conflicts → LOW-first  | 0.511                |
| conflicts → HIGH-first | 0.467 |     | conflicts → HIGH-first | 0.556                |

At 0.5 the conflict rate is low (6.7%) but so is accuracy; the searched
thresholds raise accuracy only by creating a 22.2% conflict rate (over the ≤10%
target). Excluding conflicts (0.600) is not a fair comparison — it drops the
hardest cases.

## Phase 10 — Margin-based classification

`margin = P(LOW) − P(HIGH)` (positive ⇒ toward LOW):

| Expected | mean   | median | q1     | q3     | min    | max   |
| -------- | ------ | ------ | ------ | ------ | ------ | ----- |
| LOW      | −0.077 | 0.000  | −0.277 | 0.121  | −0.559 | 0.371 |
| MEDIUM   | −0.304 | −0.143 | −0.673 | −0.050 | −0.931 | 0.250 |
| HIGH     | −0.537 | −0.573 | −0.879 | −0.275 | −1.000 | 0.226 |

The margin is **monotone** (LOW > MEDIUM > HIGH) but shifted negative (driven by
`P(HIGH)`), and the classes overlap heavily. Best symmetric band (T=0.16,
calibration-only): acc 0.489, LOW 0.133 / MED 0.533 / HIGH 0.800 — the LOW band
is effectively lost. Margin does not beat the hierarchy.

## Phase 11 — Binary prompt variants

| Variant                                    | @0.5 acc | balAcc | prec  | rec   | spec  | F1    | best T | balAcc@T | P(true) L/M/H         |
| ------------------------------------------ | -------- | ------ | ----- | ----- | ----- | ----- | ------ | -------- | --------------------- |
| LOW_A ("appropriate for LOW?")             | 0.600    | 0.500  | 0.333 | 0.200 | 0.800 | 0.250 | 0.05   | 0.600    | 0.213 / 0.180 / 0.373 |
| LOW_B ("simple… no consequential change?") | 0.444    | 0.367  | 0.143 | 0.133 | 0.600 | 0.138 | 0.00   | 0.500    | 0.216 / 0.364 / 0.478 |
| HIGH_A ("require HIGH?")                   | 0.644    | 0.617  | 0.471 | 0.533 | 0.700 | 0.500 | 0.58   | 0.667    | 0.388 / 0.143 / 0.466 |
| HIGH_B (escalation wording)                | 0.511    | 0.550  | 0.370 | 0.667 | 0.433 | 0.476 | 0.59   | 0.583    | 0.582 / 0.569 / 0.653 |

Both LOW formulations give **HIGH the highest `P(true)`** — the boolean channel is
severity-driven, and LOW wording does not invert it. HIGH_A is the better HIGH
formulation (balAcc 0.667). No variant is a reliable LOW detector.

## Phase 12 — Label leakage

**Confirmed: no leakage.** The only model inputs are `state` (the task
description) and the fixed `question` text. `ask_boolean` builds
`{"state": case["state"], "questions":[{"id":"binary","type":"boolean","text":question}]}`
— the expected class is used solely in evaluation metadata and is never placed in
any request.

## Phase 13 — Calibration comparison table

| Method                             | Acc       | MacroR       | LOW-R | MED-R | HIGH-R | Pred L/M/H | Conflict |
| ---------------------------------- | --------- | ------------ | ----- | ----- | ------ | ---------- | -------- |
| Phase 1 direct CHOICE              | 0.356     | 0.356        | 0.000 | 0.333 | 0.733  | 1/16/28    | —        |
| Phase 2 best order-ensemble        | 0.444     | 0.444        | 0.200 | 0.200 | 0.933  | 4/3/38     | —        |
| Phase 2 few-shot                   | 0.422     | 0.422        | 0.800 | 0.000 | 0.467  | 32/0/13    | —        |
| Phase 2 SCORE threshold            | 0.511     | 0.511        | 0.533 | 0.267 | 0.733  | 8/4/33     | —        |
| Phase 3 Binary LOW (A_low)         | _binary_  | balAcc 0.583 | 0.200 | —     | —      | —          | —        |
| Phase 3 Binary HIGH (B_high)       | _binary_  | balAcc 0.667 | —     | —     | 0.667  | —          | —        |
| H1 default                         | 0.489     | 0.489        | 0.200 | 0.667 | 0.600  | 4/24/17    | —        |
| H1 searched                        | 0.511     | 0.511        | 0.600 | 0.533 | 0.400  | 20/18/7    | —        |
| H2 default                         | 0.467     | 0.467        | 0.067 | 0.667 | 0.667  | 1/24/20    | —        |
| **H2 searched**                    | **0.556** | **0.556**    | 0.467 | 0.533 | 0.667  | 10/18/17   | —        |
| Dual-boundary default              | 0.489     | 0.489        | 0.200 | 0.667 | 0.600  | 4/24/17    | 6.7%     |
| Dual-boundary searched             | 0.556     | 0.556        | 0.467 | 0.533 | 0.667  | 10/18/17   | 22.2%    |
| Margin band                        | 0.489     | 0.489        | 0.133 | 0.533 | 0.800  | 4/19/22    | —        |
| **Best calibration (H2 searched)** | **0.556** | **0.556**    | 0.467 | 0.533 | 0.667  | 10/18/17   | —        |

## Phase 14 — Frozen candidate

Frozen on calibration only: **H2 (HIGH-first)**, LOW question = **A_low**, HIGH
question = **B_high**, thresholds **LOW = 0.09, HIGH = 0.65**, conflict policy =
never silently resolved (HIGH precedence when both fire), aggregation = single
run (the model is deterministic — Phase 2).

It is **not credible enough for `GO_SHADOW`** (fails the Phase 16 gates), but it
is frozen here strictly to obtain validation evidence. Calibration: acc 0.556,
macroR 0.556, LOW 0.467, MED 0.533, HIGH 0.667.

## Phase 15 — Held-out validation

| Metric        | Calibration | Validation | Δ          |
| ------------- | ----------- | ---------- | ---------- |
| Accuracy      | 0.556       | **0.367**  | **−0.189** |
| Macro recall  | 0.556       | 0.367      | −0.189     |
| LOW recall    | 0.467       | 0.300      | −0.167     |
| MEDIUM recall | 0.533       | 0.200      | −0.333     |
| HIGH recall   | 0.667       | 0.600      | −0.067     |

Validation confusion (rows=expected): LOW [3,2,5], MED [4,2,4], HIGH [1,1,8]
(predicted 6/9/15). Binary HIGH on validation: acc 0.533, balAcc 0.550, rec 0.600,
prec 0.375. **The calibration gain does not generalize**; the searched thresholds
overfit the 45-case set.

## Phase 16 — Success criteria (frozen candidate)

| Gate                           | Target | Calibration                   | Validation | Met          |
| ------------------------------ | ------ | ----------------------------- | ---------- | ------------ |
| Accuracy                       | ≥ 0.65 | 0.556                         | 0.367      | ✗            |
| Macro recall                   | ≥ 0.60 | 0.556                         | 0.367      | ✗            |
| LOW recall                     | ≥ 0.60 | 0.467                         | 0.300      | ✗            |
| MEDIUM recall                  | ≥ 0.50 | 0.533                         | 0.200      | ✗            |
| HIGH recall                    | ≥ 0.60 | 0.667                         | 0.600      | ✓            |
| Conflict rate                  | ≤ 10%  | 22.2% (searched) / 6.7% (0.5) | —          | ✗ (searched) |
| No class disappears            | —      | MEDIUM present                | present    | ✓            |
| Validation preserves behaviour | —      | —                             | degrades   | ✗            |

**No `GO_SHADOW`.**

## Phase 17 — Versus Phase 1 and Phase 2

| Method                      | Acc            | MacroR    | LOW-R | MED-R | HIGH-R |
| --------------------------- | -------------- | --------- | ----- | ----- | ------ |
| Phase 1 direct CHOICE       | 0.356          | 0.356     | 0.000 | 0.333 | 0.733  |
| Phase 2 best order-ensemble | 0.444          | 0.444     | 0.200 | 0.200 | 0.933  |
| Phase 2 best few-shot       | 0.422          | 0.422     | 0.800 | 0.000 | 0.467  |
| Phase 2 SCORE               | 0.511          | 0.511     | 0.533 | 0.267 | 0.733  |
| Phase 3 Binary LOW          | _balAcc 0.583_ | —         | 0.200 | —     | —      |
| Phase 3 Binary HIGH         | _balAcc 0.667_ | —         | —     | —     | 0.667  |
| Phase 3 H1                  | 0.511          | 0.511     | 0.600 | 0.533 | 0.400  |
| Phase 3 H2                  | **0.556**      | **0.556** | 0.467 | 0.533 | 0.667  |
| Phase 3 dual-boundary       | 0.556          | 0.556     | 0.467 | 0.533 | 0.667  |
| Phase 3 margin              | 0.489          | 0.489     | 0.133 | 0.533 | 0.800  |
| Phase 3 best calibration    | **0.556**      | **0.556** | 0.467 | 0.533 | 0.667  |
| Phase 3 validation          | 0.367          | 0.367     | 0.300 | 0.200 | 0.600  |

**Did binary decomposition materially improve routing?** Partly on calibration —
it removed the total HIGH collapse (pred 10/18/17 vs 1/16/28) and produced the
first balanced, all-classes-present result (H2 0.556 vs direct 0.356/0.444), and
it exposed an ordinal HIGH signal (`P(B_high)`). But it **fails the gates and does
not generalize** (validation 0.367). So: a real but insufficient improvement for
routing.

## Phase 18 — Julia's useful capability

**B — Routing unsuitable, binary governance useful.**

3-class and hierarchical routing are not reliable (fails gates, degrades on
validation). But one narrower binary question is genuinely useful and robust:
**"Does this task require escalation?"** (see Phase 19).

## Phase 19 — Governance probe (RUN)

Fixed questions, HIGH-vs-rest, calibration and validation:

| Question                                                 | Calib acc / balAcc | Valid acc / balAcc              | Valid rec / spec |
| -------------------------------------------------------- | ------------------ | ------------------------------- | ---------------- |
| **G1 escalation** ("Does this task require escalation?") | 0.556 / **0.617**  | **0.733 / 0.775**               | 0.900 / 0.650    |
| G2 architecture boundary                                 | 0.622 / 0.467      | 0.700 / 0.550                   | 0.100 / 1.000    |
| G3 authorization/security                                | 0.622 / 0.467      | NOT_RUN (calib rec 0.000)       | —                |
| G4 operational consequences                              | 0.578 / 0.433      | P(true) L/M/H 0.117/0.068/0.079 | (inverted)       |

`P(true)` by class for **G1**: calibration LOW 0.500 / MED 0.463 / HIGH 0.795;
validation LOW 0.409 / MED 0.421 / **HIGH 0.840**. G1 is the **strongest and most
robust signal found in any phase** — it separates HIGH cleanly and _improves_ on
validation (balAcc 0.775, recall 0.900). G2/G3/G4 mostly fail to fire or invert.
So: **the governance question separates better than execution-class routing.**

## Phase 20 — Artifacts

- `tools/julia/calibrate_routing3.py` (extended: binaries, prompt variants,
  governance probe, dual-boundary, margin, hierarchies)
- `docs/reports/julia-routing-calibration-phase3.md` (this report)
- Phase 1/2 reports and frozen datasets untouched.

## Phase 21 — Quality gates

```
gofmt -l .            clean
go vet ./...          OK
go test -count=1 ./... ok
go test -race ./...   ok
go build ./...        OK
git diff --check      clean
python test_http_bridge.py (45) OK ; test_infer.py (26) OK
git diff --name-only -- decision/                 0 files
git diff --name-only -- internal/providers/julia/ 0 files
agentic-sop status                                 clean
```

## Phase 22 — Final decision

**`HOLD_ROUTING`.**

Julia does not provide sufficiently reliable LOW/MEDIUM/HIGH routing (H2 fails
the Phase 16 gates and degrades from 0.556 to 0.367 on validation), **but** the
narrower binary governance question G1 — "Does this task require escalation?" —
is a useful, robust HIGH-vs-rest signal (balAcc 0.617 calibration → **0.775
validation**, recall 0.900). That evidence supports a narrower binary use, not a
router. Not `GO_SHADOW`; not `NO_GO` (no correctness/reliability defect).

---

## Required Final Report

1. **Branch / HEAD:** `main` @ `4a0d65dbdff1e4373dfe92d9dfc8bfdf72c2970b`.
2. **Working-tree baseline:** clean; only Phase 3 evaluation artifacts uncommitted.
3. **Files changed:** `tools/julia/calibrate_routing3.py`, `docs/reports/julia-routing-calibration-phase3.md` (both new, uncommitted).
4. **Phase 1 calibration hash:** `b96a970c9b2a42b05f48bd7ff5af0c147a3728c770ff6ea1fe55b2314266f74b` (unmodified).
5. **Phase 2 validation hash:** `d0cbd40e67dbb4ca1e3896c2659f04ee860b9471950fd245c467a115c6cbc2b6` (unmodified).
6. **Binary LOW (A_low):** @0.5 acc 0.711, prec 0.750, rec 0.200, spec 0.967, balAcc 0.583; best T 0.09 → balAcc 0.617.
7. **Binary HIGH (B_high):** @0.5 acc 0.667, prec 0.500, rec 0.667, spec 0.667, balAcc 0.667; best T 0.65 → balAcc 0.717.
8. **Escalation binary (C_escalation):** @0.5 acc 0.622, prec 0.469, rec 1.000, spec 0.433, balAcc 0.717 (over-escalates).
9. **Binary probability distributions:** A_low 0.227/0.094/0.125; B_high 0.304/0.398/0.662 (ordinal); C_escalation 0.607/0.640/0.864.
10. **Default 0.5 results:** see §Phase 3/9/11 tables.
11. **Threshold-search results:** A_low 0.09, B_high 0.65, C_escalation 0.49 (evaluation-only).
12. **H1 results:** default 0.489; searched 0.511 (LOW 0.600 / MED 0.533 / HIGH 0.400).
13. **H2 results:** default 0.467; searched **0.556** (LOW 0.467 / MED 0.533 / HIGH 0.667).
14. **Dual-boundary results:** @0.5 acc 0.489 (6.7% conflict); @searched acc 0.556 (22.2% conflict).
15. **Conflict rate:** 6.7% @0.5; 22.2% @searched (> 10% target).
16. **Margin analysis:** monotone (LOW −0.077 > MED −0.304 > HIGH −0.537) but overlapping; best band acc 0.489.
17. **Prompt-variant results:** LOW_A balAcc 0.500, LOW_B 0.367 (both severity-inverted); HIGH_A 0.617, HIGH_B 0.550.
18. **Best calibration candidate:** H2 (HIGH-first, A_low/B_high, 0.09/0.65) — acc 0.556, macroR 0.556.
19. **Frozen candidate definition:** H2, A_low, B_high, LOW 0.09 / HIGH 0.65, HIGH precedence on conflict, single deterministic run.
    20–24. **Calibration accuracy 0.556 / macroR 0.556 / LOW 0.467 / MED 0.533 / HIGH 0.667.**
    25–29. **Validation accuracy 0.367 / macroR 0.367 / LOW 0.300 / MED 0.200 / HIGH 0.600.**
20. **Validation confusion:** LOW [3,2,5], MED [4,2,4], HIGH [1,1,8].
21. **Predicted distributions:** calibration 10/18/17; validation 6/9/15.
22. **Calibration→validation deltas:** acc −0.189, macroR −0.189, LOW −0.167, MED −0.333, HIGH −0.067.
23. **vs Phase 1:** H2 0.556 vs direct 0.356 (no HIGH collapse; all classes present).
24. **vs Phase 2:** H2 0.556 vs order-ensemble 0.444 / SCORE 0.511; H2 is the best 3-class result but still < gates.
25. **Did binary decomposition materially improve routing?** Partly on calibration; it failed the gates and did not generalize.
26. **Narrower governance value?** **Yes** — G1 escalation: balAcc 0.617 calib → 0.775 valid, recall 0.900.
27. **Remaining bias/failure patterns:** LOW suppression (A_low fires 4/45); severity-driven `P(true)` (LOW wording still yields HIGH-highest); wording fragility; HIGH dominance in most choice methods; threshold overfit.
28. **No label leakage:** confirmed (inputs are `state` + fixed `question` only).
29. **Provider-neutral contract unchanged:** confirmed (0 files under `decision/`).
30. **Julia provider production code unchanged:** confirmed (0 files under `internal/providers/julia/`).
31. **`agentic-sop` untouched:** confirmed clean.
32. **Julia not made authoritative:** confirmed (evaluation tooling only).
33. **Clef work preserved:** yes (committed in `79569b8`; untouched).
34. **Final decision:** **`HOLD_ROUTING`**.
