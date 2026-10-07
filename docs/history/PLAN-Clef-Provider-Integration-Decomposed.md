# PROPOSAL — Clef Provider Integration (Decomposed)

**Repository:** `~/agentic-workspace/projects/sop-decision-adapters`
**Scope:** `sop-decision-adapters` only (adapter-side).
**Status:** PROPOSED — review document. **NOT activated. Not independently activatable.**
**Revision:** rev 2 — CLEF-015 (serve/wire) removed; `internal/wire` orphaned; CLEF-005 kept
verbatim and historically superseded; config/capability tests re-homed to CLEF-016.
**Supersedes (on approval):** `docs/plans/PLAN-Clef-Provider-Integration.md` (the active plan).

> This document is a **plan-repair / decomposition proposal**, not an executable plan
> replacement by itself. It contains the **full text of every new or changed stage** and an
> explicit disposition for every retained stage. Sections named `## Proposal …` are review
> notes; a plan compiler ignores unrecognized top-level sections, so they do not become tasks.

Basis of this proposal (repository + run evidence, read-only):

- SOP state at proposal time: `CLEF-001/002/003` LOCAL_DONE, `CLEF-004` LOCAL_DONE,
  `CLEF-005` BLOCKED, `CLEF-006…010` PLANNED; plan `plan-clef-provider-integration` ACTIVE.
- CLEF-004 success run: `PASSED`, model `glm-5.3-flash:cloud`, `changes_expected:false`, empty
  summary — it validated pre-existing working-tree output, it did not add code.
- CLEF-005 run: `NO_PROGRESS`/BLOCKED, model `glm-5.3-flash:cloud`, `iterations=20`,
  `discovery_inspections=10`, `repository_mutations=0`.
- Working tree: `internal/providers/clef/*` (6 files, no tests, no importers),
  `internal/wire/wire.go` (no importers), `.env.example` (Clef block), `docs/specs/clef-provider.md`,
  `internal/providers/julia/config_test.go` (test hermeticity fix).

---

## Project

Clef Provider Integration

## Summary

Implement a Clef decision provider in `sop-decision-adapters` that satisfies the
already-proven provider-neutral decision contract exposed by `agentic-sop` without
introducing Clef-specific policy, lifecycle, approval, authorization, or execution-routing
behavior into `agentic-sop`.

**Integration boundary (revised).** The provider-neutral integration boundary for this plan
is the **existing in-process `decision.Provider` seam**, selected through the existing
`newProvider` construction-time switch in `cmd/sop-decision-adapter/main.go`:

```text
agentic-sop (or any caller)
    |
    | in-process decision.Provider  (decision / DecisionRequest / Question / DecisionResult)
    v
internal/providers/clef
    |
    | provider-specific transport
    v
Clef runtime (/v1/systemone)
    |
    v
decision evidence only
```

This **revises** the active plan's Summary/Architecture-Invariant diagram that named a
`sop-decision-adapters serve` + "provider-neutral JSON" process boundary. A generic
`agentic-sop -> provider-neutral JSON -> sop-decision-adapters serve -> provider` runtime is
**outside this repaired Clef plan**; it may be pursued later as a separate provider-runtime
architecture project. No new wire layer, serve subcommand, registry, or factory is introduced
by this plan.

The current evidence supports Ollama `POST /v1/systemone` with `clef-flash` as the primary
verified Clef transport.

`mlx-community/clef-4bit` through oMLX remains an evaluation target and MUST NOT be treated as
equivalent unless CLEF-002 demonstrates fidelity-preserving Clef decision semantics.

Clef remains OFF and non-default throughout this plan.

**Decomposition rationale.** The original `CLEF-004` ("Implement Clef Provider") and
`CLEF-005` ("Provider Contract and Failure Tests") proved too broad for the bounded execution
window: multiple models (DeepSeek, Nemotron, GLM) spent their whole allowance in DISCOVER with
zero repository mutations. This proposal replaces the remaining oversized work with small,
**mutation-oriented** tasks whose expected execution pattern is:
_read the authoritative contract → inspect one precedent → mutate the target → run focused
tests → verify_.

## Capabilities

### Go build/test toolchain — EXISTS

- Owner: operator-supplied development environment.
- Evidence: `go build ./...`, `go vet ./...`, `go test ./...` pass offline at the proposal baseline.

### agentic-sop provider-neutral decision seam — EXISTS

- Owner: `agentic-sop` (read-only for this plan).
- Evidence: the provider-neutral request/result boundary, fail-closed validation, and SOP-owned
  evidence interpretation are established; the adapter side depends only on the local `decision`
  package through the in-process `decision.Provider` seam.

### Ollama SystemOne decision endpoint serving Clef — EXISTS

- Owner: local Ollama runtime.
- Evidence: `POST /v1/systemone` serving `clef-flash` produced bounded choice/score/noul evidence
  (CLEF-002, `docs/reports/clef-provider/CLEF-002-capability-transport.md`).

### In-process adapter implementation of the local provider seam — EXISTS (adapter-side)

- Owner: this repository.
- Evidence: `internal/providers/clef/{clef,config,capability,translate,systemone,transport}.go`
  implement `decision.Provider`, compile, and pass `go build`/`go vet`/`go test` in a clean
  environment. Not yet registered, unverified by tests.

### Apple Silicon MLX runtime / oMLX with clef-4bit — UNKNOWN

- Owner: local oMLX runtime.
- Evidence: oMLX exposes OpenAI-compatible generation; `clef-4bit` is loaded; native Clef
  decision semantics equivalent to `/v1/systemone` have NOT been established (CLEF-002 §5).
- Gap: no fidelity-preserving decision transport is demonstrated for oMLX/clef-4bit.

## Assumptions

### The already-present `internal/providers/clef` package is architecturally correct and reusable

- Evidence: it implements the existing `decision.Provider` seam, mirrors `internal/providers/nimble`,
  keeps Clef transport/types private, and is OFF by default via `CLEF_ENABLED`; it compiles and
  passes the baseline gates in a clean environment.
- Consequence: the decomposed tasks extend and test it rather than rewriting it; no new provider
  abstraction is designed.

### The in-process `decision.Provider` / `newProvider` seam is sufficient for Clef integration

- Evidence: the repository's only existing provider-selection mechanism is the construction-time
  `switch` in `cmd/sop-decision-adapter/main.go`; there is no existing wire layer, serve command,
  registry, or factory; CLEF-001 §11 records the concrete `agentic-sop` process integration as
  UNKNOWN.
- Consequence: no new serve/wire architecture is introduced for Clef; `internal/wire` is treated
  as an orphaned speculative artifact to be removed from the Clef implementation (not deleted
  during this revision).

---

## Proposal: Existing Artifact Classification (rev 2)

| Artifact                                                                            | Class                                                               | Evidence / rationale                                                                                                                                |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/providers/clef/{clef,config,capability,translate,systemone,transport}.go` | **KEEP**                                                            | Implements the existing `decision.Provider` seam; mirrors the nimble precedent; private transport + types; `CLEF_ENABLED` OFF by default; compiles. |
| `docs/specs/clef-provider.md`                                                       | **KEEP**                                                            | Matches the `docs/specs/*` convention; describe the in-process seam only.                                                                           |
| `.env.example` Clef block                                                           | **KEEP**                                                            | Commented, OFF-by-default entries (`CLEF_ENABLED`/`CLEF_MODEL`/`CLEF_TIMEOUT`); satisfies `configured != enabled`.                                  |
| `internal/wire/wire.go`                                                             | **REMOVE FROM CLEF IMPLEMENTATION / ORPHANED SPECULATIVE ARTIFACT** | See below. Not removed during this revision.                                                                                                        |
| `internal/providers/julia/config_test.go` (hermeticity change)                      | **OUT-OF-SCOPE FOR CLEF / ROUTE TO JULIA TRACK**                    | See below. Not modified, reverted, deleted, or committed.                                                                                           |
| `internal/decision/` (empty directory, no Go sources)                               | **REMOVE (recommended)**                                            | Empty, unreferenced; not a code artifact.                                                                                                           |
| `cmd/sop-decision-adapter/main.go`                                                  | **KEEP, missing wiring**                                            | Selection is the construction-time `switch` in `newProvider`; no `clef` case yet → CLEF-011.                                                        |
| Provider-neutral `serve`/wire boundary                                              | **OUT OF SCOPE**                                                    | Not introduced by this plan (rev 2). No CLEF-015.                                                                                                   |

### `internal/wire` disposition (rev 2)

**REMOVE FROM CLEF IMPLEMENTATION / ORPHANED SPECULATIVE ARTIFACT.**

Rationale:

- **No importers** — `grep -rn "internal/wire"` finds no importing package; `internal/wire` has no tests.
- **No repository precedent** — the repo has no existing wire layer, `serve` command, registry, or
  factory; its real, tested integration surface is the in-process `decision.Provider` seam.
- **Not necessary for Clef** — Clef already implements the existing `decision.Provider` seam; no
  additional neutral layer is required to satisfy the acceptance criteria.
- **Out of scope** — generic `serve`/wire architecture is a **separate provider-runtime
  architecture project**, not part of this repaired Clef plan.

Action: **do not remove the file during this revision.** Removal is a repository change that
requires explicit approval and must preserve all other user-owned working-tree changes. When
approved, removal happens as a normal adapter-side change (not in this proposal step).

### Julia test change disposition (rev 2)

**OUT-OF-SCOPE FOR CLEF / ROUTE TO JULIA TRACK.** `internal/providers/julia/config_test.go`
was modified (test-only hermeticity hardening: it neutralizes ambient `JULIA_*` before each
subtest). It changes **no Julia behavior**, but it is Julia-track work performed during Clef
work. It must **not** be modified, reverted, deleted, or committed as part of the Clef plan, and
**Clef completion must not depend on accepting it.** If the workspace validation is affected by
ambient `JULIA_*`, that is an environment concern, not a Clef acceptance requirement.

## Proposal: Old Task → Proposed Task Mapping (rev 2)

| Old task                                           | Disposition                                                                                                                                                                      | Proposed task(s)                                                     |
| -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| CLEF-001 Capture Adapter Repository Truth          | **KEEP (historical, LOCAL_DONE)**                                                                                                                                                | unchanged                                                            |
| CLEF-002 Verify Clef/oMLX Capability and Transport | **KEEP (historical, LOCAL_DONE)**                                                                                                                                                | unchanged                                                            |
| CLEF-003 Define Clef-to-Provider Contract Mapping  | **KEEP (historical, LOCAL_DONE)**                                                                                                                                                | unchanged                                                            |
| CLEF-004 Implement Clef Provider                   | **KEEP (historical, LOCAL_DONE, definition verbatim)** — residual deliverables re-homed                                                                                          | **CLEF-011** (registration), **CLEF-016/012/013/014** (tests)        |
| CLEF-005 Provider Contract and Failure Tests       | **KEEP VERBATIM (definition unchanged)** — status dispositioned **NOT_REQUIRED** (historically superseded); BLOCKED history, retries, reports and NO_PROGRESS evidence preserved | superseded by **CLEF-016**, **CLEF-012**, **CLEF-013**, **CLEF-014** |
| CLEF-006 Local Clef Runtime Verification           | **KEEP, REWORD dependencies**                                                                                                                                                    | **CLEF-006′** (deps `CLEF-011`, `CLEF-014`)                          |
| CLEF-007 Shadow Evaluation Harness                 | **KEEP, REWORD dependency**                                                                                                                                                      | **CLEF-007′** (dep `CLEF-006′`)                                      |
| CLEF-008 Benchmark and Compare Decision Quality    | **KEEP**                                                                                                                                                                         | unchanged                                                            |
| CLEF-009 Full Verification and Security Review     | **KEEP, REWORD** (drop wire/serve guard wording; in-process seam)                                                                                                                | **CLEF-009′**                                                        |
| CLEF-010 Selectable-Provider Readiness Review      | **KEEP, REWORD** (in-process seam wording)                                                                                                                                       | **CLEF-010′**                                                        |
| —                                                  | **ADD**                                                                                                                                                                          | **CLEF-011**, **CLEF-012**, **CLEF-013**, **CLEF-014**, **CLEF-016** |

**No task removed. No executed task's definition changed.** CLEF-005 is _not_ rewritten and does
_not_ require `--accept-changed`; it is dispositioned `NOT_REQUIRED` (see below).

## Proposal: Historical Supersession of CLEF-005 (supported mechanism)

SOP supports superseding CLEF-005 **without changing its definition**, via the domain transition
`Task.MarkNotRequired` (`internal/domain/task.go`): _"only a task that has not started (PLANNED,
READY) or is BLOCKED may take it, a reason is required, and the transition is appended to the
task's history so it is auditable. Earlier attempts are kept. It records no implementation and no
approval."_ CLEF-005 is BLOCKED → eligible.

The supported operator command is:

```bash
sop task skip CLEF-005 \
  --reason "historically superseded by decomposed tasks CLEF-016/CLEF-012/CLEF-013/CLEF-014" \
  --evidence .agent-sdlc/runs/CLEF-005/report.json \
  --evidence .agent-sdlc/runs/CLEF-005/classification.json
```

Effects (verified from source): status `BLOCKED → NOT_REQUIRED`; a `NotRequired` provenance record is
written; earlier attempts/reports/NO_PROGRESS evidence are **kept**; `NOT_REQUIRED` is terminal and
`IsSatisfied()` (so it stops blocking runnable work) but `IsCompleted()` is false — **no
implementation success is fabricated**; no approval gate is bypassed (a pending gate would refuse).
CLEF-005's plan **stage text is left verbatim**, so reconciliation sees it as _unchanged_.

**No lifecycle limitation encountered** — SOP can represent this safely without changing an
executed task's definition and without `--accept-changed`.

## Proposal: Retained Stage History (verbatim, not reproduced)

The following active-plan stages are **carried over verbatim** and are **not reproduced here** to
avoid content drift. Their full text remains authoritative in
`docs/plans/PLAN-Clef-Provider-Integration.md`:

| Stage    | State to preserve                                                             |
| -------- | ----------------------------------------------------------------------------- |
| CLEF-001 | LOCAL_DONE; run artifacts, report, evidence                                   |
| CLEF-002 | LOCAL_DONE; report, captures, completion evidence                             |
| CLEF-003 | LOCAL_DONE; contract-mapping report                                           |
| CLEF-004 | LOCAL_DONE; all attempts incl. FAILED attempts, `NO_PROGRESS`, diffs, retries |
| CLEF-005 | definition verbatim; BLOCKED → to be dispositioned NOT_REQUIRED               |
| CLEF-007 | PLANNED (carried verbatim; dependency re-pointed as below)                    |
| CLEF-008 | PLANNED (carried verbatim)                                                    |

## Proposal: Harness Backlog (separate; do NOT solve here)

Recorded for `agentic-sop`, **not** part of this plan: _"implementation convergence / bounded
discovery."_ Evidence: multiple implementation models (DeepSeek, Nemotron, GLM) remained in
DISCOVER until `NO_PROGRESS`; mutation count stayed 0; raising `SOP_OLLAMA_STALE_ITERATIONS` 5→8
did not reliably resolve it; continuation checkpoints did not reliably prevent rediscovery;
reproduced across CLEF-004 and CLEF-005. Suggested future investigation: DISCOVER→IMPLEMENT
transition, `implementNowAfter` behavior, late-stage discovery restrictions,
continuation-checkpoint reuse, and a stronger mutation-or-blocker requirement after bounded
discovery. **No `agentic-sop` change is made or proposed here.**

---

## Execution Policy

```text
CLEF-001  (LOCAL_DONE)
    |
CLEF-002  (LOCAL_DONE)
    |
CLEF-003  (LOCAL_DONE)
    |
CLEF-004  (LOCAL_DONE)
    |
    +-----------------------------+
    |                             |
CLEF-011                     CLEF-016
(register in selection)      (config + capability tests)
                                  |
                              CLEF-012
                             (translation + normalization tests)
                                  |
                              CLEF-013
                             (fake transport + failure tests)
                                  |
                              CLEF-014
                             (governance + neutrality tests)
    |                             |
    +--------------+--------------+
                   |
                CLEF-006′  (live runtime verification)
                   |
                CLEF-007′
                   |
                CLEF-008
                   |
                CLEF-009′
                   |
                CLEF-010′

CLEF-005  (historical BLOCKED → NOT_REQUIRED; superseded by CLEF-016/012/013/014)
```

One authoritative path to completion. Each implementation task names at most **two** precedents.

---

## CLEF-011 — Register Clef in Provider Selection

Make the existing `internal/providers/clef` package explicitly selectable through the existing
in-process selection seam, OFF by default.

### Dependencies

- CLEF-004

### Requires

- Go build/test toolchain

### Deliverables

- `cmd/sop-decision-adapter/main.go` — a `clef` case in `newProvider` (the documented selection
  seam) and the `availableProviders` list; explicit-opt-in wording in usage.

### Acceptance Criteria

- `-provider clef` constructs a Clef provider; an unknown provider still fails closed.
- Clef is OFF unless `CLEF_ENABLED` is explicitly set.
- No registry/factory/serve abstraction is introduced; existing nimble/julia selection is unchanged.
- A focused test proves off-by-default and explicit selection.

### Production-Change Scope

Adapter-side CLI selection only.

---

## CLEF-016 — Clef Configuration and Capability Tests

Prove offline that Clef configuration and capability reporting satisfy the provider-neutral
contract. No real model is required.

### Dependencies

- CLEF-004

### Requires

- Go build/test toolchain

### Deliverables

- `internal/providers/clef/config_test.go`
- `internal/providers/clef/capability_test.go`

### Acceptance Criteria

- `ConfigFromEnv` leaves Clef OFF by default (no `CLEF_ENABLED` ⇒ provider not enabled).
- Explicit `CLEF_ENABLED` opt-in enables Clef; blank/non-positive values fall back to defaults.
- Endpoint/model/timeout derive from `OLLAMA_BASE_URL`/`CLEF_MODEL`/`CLEF_TIMEOUT`.
- Capability reporting returns UNSUPPORTED for operations Clef does not implement.
- `go test ./...` remains fully offline.

### Production-Change Scope

Tests only.

---

## CLEF-012 — Clef Translation and Normalization Tests

Prove offline that Clef request/response translation preserves provider-neutral semantics.

### Dependencies

- CLEF-016

### Requires

- Go build/test toolchain

### Deliverables

- `internal/providers/clef/translate_test.go`

### Acceptance Criteria

- Request translation covers every `DecisionRequest`/`Question` field.
- Response normalization covers choice, score, and null-like forms.
- Confidence and probabilities are bounded; NaN/Inf are rejected.
- A choice outside the request's allowed set is rejected, not returned.
- Indeterminate outcomes remain indeterminate.

### Production-Change Scope

Tests only.

---

## CLEF-013 — Clef Fake-Transport and Failure Tests

Prove offline that Clef fails closed across transport and malformed-result conditions.

### Dependencies

- CLEF-012

### Requires

- Go build/test toolchain

### Deliverables

- `internal/providers/clef/transport_test.go`
- `internal/providers/clef/clef_test.go`

### Acceptance Criteria

- A fake transport drives the provider without any real model.
- Malformed, empty, timeout, and cancelled responses map to ERROR (never success).
- Transport failure never yields a success-shaped result or successful evidence.
- Unsupported operations return UNSUPPORTED; indeterminate outcomes stay indeterminate.
- `go test ./...` remains fully offline.

### Production-Change Scope

Tests plus narrowly required implementation hardening only. No contract widening.

---

## CLEF-014 — Clef Governance-Boundary and Provider-Neutrality Tests

Prove offline that Clef output carries no governance authority and no provider-specific type leaks
into the neutral layer.

### Dependencies

- CLEF-013

### Requires

- Go build/test toolchain

### Deliverables

- `internal/providers/clef/governance_test.go`
- a neutrality search/assertion (no Clef identifier in the public `decision` package)

### Acceptance Criteria

- Provider output contains no authoritative CONTINUE/BLOCK/APPROVE/REJECT/COMMIT/MERGE, lifecycle
  transition, or approval transition (literal domain data remains ordinary data).
- No provider-specific type is added to the public `decision` package.
- Provider-neutral surfaces contain no Clef-specific branch or identifier.
- `go test ./...` remains fully offline.

### Production-Change Scope

Tests only.

---

## CLEF-006 — Local Clef Runtime Verification

Verify the implemented provider against a real local Clef runtime (dependency re-pointed).

### Dependencies

- CLEF-011
- CLEF-014

### Requires

- Go build/test toolchain
- Ollama SystemOne decision endpoint serving Clef

### Deliverables

- opt-in live integration test (disabled by default; explicit opt-in; clean skip when absent)
- `docs/reports/clef-provider/CLEF-006-local-runtime-verification.md`

### Acceptance Criteria

- The live test is disabled by default, requires explicit opt-in, and skips cleanly when
  prerequisites are absent.
- It validates contract shape (not fixed model judgment), choice membership, and confidence bounds.
- Records request success rate, latency p50/p95, choice validity, confidence availability, timeout
  behavior, malformed/indeterminate behavior, and repeatability.
- If oMLX/clef-4bit is not viable, record `NOT EXERCISED — see CLEF-002`.

### Production-Change Scope

Opt-in tests and report only.

---

## CLEF-009 — Full Verification and Security Review

Complete repository verification and architecture review.

### Dependencies

- CLEF-008

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-provider/CLEF-009-verification.md`

### Acceptance Criteria

- Required gates (`gofmt -l .`, `go vet ./...`, `go test -count=1 ./...`, `go test -race -count=1 ./...`,
  `go build ./...`, `git diff --check`) run with exact commands and recorded outcomes.
- Each architecture guard is verified and recorded with test or inspection evidence. Guard #9 is
  worded to the in-process seam: "providers are reached only through the existing in-process
  `decision.Provider` seam; no separate serve/wire architecture is introduced."
- Security review covers transport invocation, endpoint/URL validation, environment and secret
  handling, bounded output, and fail-closed error handling.
- No SOP policy, provider contract, governance, approval, lifecycle, or execution-model routing is
  changed.

### Production-Change Scope

Verification hardening and tests only. No policy or contract widening.

---

## CLEF-010 — Selectable-Provider Readiness Review

Produce the final adapter readiness decision (`READY` / `READY_WITH_NOTES` / `NOT_READY`). This
task does not authorize making Clef selectable, live, or default.

### Dependencies

- CLEF-009

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-provider/CLEF-010-readiness-decision.md`

### Acceptance Criteria

- The report answers whether Clef is ready to enter a separate selectable-provider review through
  **the existing in-process `decision.Provider` seam** without Clef-specific
  policy/lifecycle/approval/authorization/routing changes in `agentic-sop`.
- It records repository state, CLEF-001..CLEF-010 evidence, verified transport, the provider-neutral
  boundary (in-process seam — no serve/wire layer), failure semantics, governance boundary,
  verification results, security findings, benchmark findings, residual risks, and the readiness result.
- It confirms Clef remains OFF and non-default, no `agentic-sop` change is required, SMALL/MEDIUM/LARGE
  routing is unchanged, and making Clef selectable requires separate human approval.
- If READY or READY_WITH_NOTES, the single recommended next action is:
  `Open selectable-provider review.`

### Production-Change Scope

None.

---

## Proposal: Unresolved Architecture Decisions

1. **Generic serve/wire runtime.** Explicitly **out of scope** for this plan (rev 2). If a
   provider-neutral JSON process boundary is wanted later, it is a separate provider-runtime
   architecture project; it must not be introduced under Clef integration.
2. **`internal/providers/julia/config_test.go`.** Routed to the Julia track (out of scope for
   Clef). Human call on the Julia track: accept as generic test hardening, or move it there.
   No Julia behavior may change for Clef's benefit (this change does not).

## Proposal: What Would Change in the Active Plan (if approved)

- **Added stages:** CLEF-011, CLEF-012, CLEF-013, CLEF-014, CLEF-016.
- **Changed stages (all currently PLANNED → no per-task approval needed):** CLEF-006 (dependencies),
  CLEF-007 (dependency), CLEF-009 (guard #9 wording), CLEF-010 (boundary wording).
- **Header wording:** the Summary / Architecture-Invariant diagram drops the
  `sop-decision-adapters serve` + "provider-neutral JSON" boundary in favour of the in-process
  `decision.Provider` / `newProvider` seam.
- **Unchanged (verbatim):** CLEF-001, CLEF-002, CLEF-003, CLEF-004, CLEF-005, CLEF-008.
- **No stage removed. No executed task's definition changed → no `--accept-changed`.**
- **Lifecycle transition (separate from the plan text):** CLEF-005 `BLOCKED → NOT_REQUIRED`
  (historically superseded), history preserved.

## Proposal: Application Sequence (after human approval — not run)

1. Human merges the added/changed stages above into
   `docs/plans/PLAN-Clef-Provider-Integration.md`, leaving CLEF-001…005 and CLEF-008 verbatim,
   and updates the Summary/Architecture-Invariant wording to the in-process seam.
2. Preview, read-only:
   `sop reconcile docs/plans/PLAN-Clef-Provider-Integration.md --list-changed`
3. Apply (added tasks + PLANNED-task changes; **no `--accept-changed`**, since no executed task's
   definition changes):
   `sop reconcile docs/plans/PLAN-Clef-Provider-Integration.md`
4. Disposition the historical blocked task (supported mechanism; preserves its definition and history):
   `sop task skip CLEF-005 --reason "historically superseded by CLEF-016/CLEF-012/CLEF-013/CLEF-014" --evidence .agent-sdlc/runs/CLEF-005/report.json --evidence .agent-sdlc/runs/CLEF-005/classification.json`
5. Confirm:
   `sop continue --check --json` → expect `RECONCILE_CLEAN`, runnable `CLEF-011` and `CLEF-016`.

No `--force`. No `--accept-changed`. The reconcile path preserves plan identity and all historical
task state. (Removal of `internal/wire` and any Julia-track decision are separate changes, applied
only after explicit approval, and are not part of this plan application.)

---

## Expected End State

A successful plan leaves `sop-decision-adapters` with a Clef provider that implements the existing
adapter-side provider abstraction, is reached through the existing in-process provider-neutral
`decision.Provider` seam without importing SOP internals, emits evidence only, validates choice and
confidence, fails closed, keeps Clef/Ollama/oMLX/SystemOne details adapter-side, leaves SOP
governance, approval, lifecycle, and routing unchanged, remains OFF and non-default, and is, at
most, `READY_FOR_SELECTABLE_REVIEW`. Making Clef selectable, live, or default requires a separate
plan, separate evidence, and explicit human approval.
