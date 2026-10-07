# PLAN --- Clef Selectable-Provider Review

**Repository:** `~/agentic-workspace/projects/sop-decision-adapters`\
**Scope:** `sop-decision-adapters` only (adapter-side).\
**Status:** PLANNED.\
**Prerequisite:** the **Clef Provider Integration** plan is COMPLETE and archived
(`docs/history/PLAN-Clef-Provider-Integration.md`); its readiness result was
**READY_WITH_NOTES**, and it is **not** reactivated by this plan.

This plan does **not** assume the answer is YES. It may conclude
`NOT_SELECTABLE_READY`.

Authoritative inputs (read-only):

- `docs/reports/clef-provider/CLEF-009-verification.md`
- `docs/reports/clef-provider/CLEF-010-readiness-decision.md`
- `docs/reports/clef-provider/CLEF-008-benchmark.md`
- `docs/reports/clef-provider/CLEF-006-local-runtime-verification.md`
- `docs/reports/clef-provider/CLEF-architecture-analysis.md`

Carried-forward open findings (from CLEF-009, unchanged):

- **M-1** (Medium): `BaseURL` is not parsed/validated at construction; a malformed
  URL fails closed at request creation.
- **L-1** (Low): HTTP is allowed for the local backend.
- **L-2** (Low): successful response bodies have no explicit hard byte cap.
- **I-1** (Info): invalid `CLEF_TIMEOUT` falls back to the default timeout.

Live evidence state carried forward:

- adapter/seam behavior: **VERIFIED_STATICALLY / VERIFIED_BY_TEST**
- live Clef `/v1/systemone`: **NOT_EXERCISED_LIVE / UNAVAILABLE** (CLEF-008)
- `mlx-community/clef-4bit` via oMLX: **NOT_EXERCISED**
- Nimble benchmark cell: exercised separately

---

## Project

Clef Selectable-Provider Review

## Summary

Determine whether Clef may become an **explicitly selectable** provider, and
under what conditions, without changing `agentic-sop`, without changing any
default, and without granting Clef governance authority.

"Selectable" is defined operationally by this plan (SP-002). This plan builds, at
most, a **minimal, opt-in** selection surface for the adapter/CLI and the tests
that prove default-preservation; the final task (SP-010) issues one of
`SELECTABLE_READY`, `SELECTABLE_READY_WITH_CONDITIONS`, or `NOT_SELECTABLE_READY`.

Clef remains **OFF** and **non-default** throughout this plan. No production
enablement occurs merely because tests pass.

The provider-neutral `decision.Provider` seam remains authoritative; selection
stays behind it.

---

## Governing Invariants

> Providers produce evidence. SOP determines what that evidence means.

> Providers evaluate. SOP governs.

Clef MUST NOT gain authority to:

- transition task or plan state;
- approve or reject work;
- satisfy human approval;
- authorize execution, commit, or merge;
- bypass validation or review;
- select SOP lifecycle outcomes;
- alter SMALL/MEDIUM/LARGE execution routing.

Making Clef selectable, live, or default requires explicit human authorization
under this plan's final decision; it is never implied by passing tests.

---

## Non-Goals

- Do not modify `agentic-sop`.
- Do not modify the provider-neutral `decision` contract or introduce Clef-specific
  types/branches into it.
- Do not change SOP policy, lifecycle, approval, or authorization semantics.
- Do not change SMALL/MEDIUM/LARGE routing.
- Do not change any existing provider default.
- Do not make Clef live or default.
- Do not add hidden or automatic provider selection.
- Do not fix M-1/L-1/L-2/I-1 inside this plan unless SP-003 classifies a finding as
  a precondition (then the minimal hardening is in scope).
- Do not expand into general LLM benchmarking.

---

## Architecture Invariant

```text
agentic-sop decision seam
        |
        | in-process decision.Provider  (provider-neutral)
        v
newProvider selection switch  (cmd/sop-decision-adapter/main.go)
        |
        | explicit, opt-in
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

Selection is exposed only through the existing construction-time seam. No
`serve` process boundary, registry, or factory is introduced.

SOP owns interpretation. The adapter owns provider-specific transport and
translation.

---

## Selection Invariant

```text
configured != enabled
enabled    != default
selectable != default
selectable != enabled-by-default
```

- Selection is **explicit and opt-in** only.
- Absence of selection leaves every existing default unchanged (Nimble remains the
  CLI default; no provider is auto-enabled).
- A disabled Clef provider MUST NOT be selectable accidentally (for example by an
  empty/absent value or by configuration defaults).
- Selecting Clef MUST NOT enable it; enablement remains the separate
  `CLEF_ENABLED` gate.
- Selecting Clef MUST NOT alter any other provider's availability or default.

---

## Planning and Discovery Constraint

Repository discovery performed inside a task is work performed by that task, not a
prerequisite capability. The repository working tree, its Go toolchain, its git
metadata, and its existing reports already exist and are recorded as EXISTS.

Only externally supplied runtimes, permissions, tools, artifacts, or already
completed dependencies belong in `Requires`.

---

## SP-001 — Selectability Baseline

Establish the exact current state relevant to provider selectability.

Inspect and record from the repository itself (source is authoritative; do not
infer from documentation):

- the current provider-selection mechanism (`newProvider`) and its exact surface;
- every existing provider's default/selection behavior (including Nimble and Julia);
- the enablement gates (`CLEF_ENABLED`) and configuration defaults (`CLEF_MODEL`,
  `OLLAMA_BASE_URL`, `CLEF_TIMEOUT`);
- where a `clef` selection currently resolves and what happens when Clef is
  disabled;
- the `decision.Provider` seam and the current `clef_selection_test.go` coverage;
- the carried-forward findings M-1/L-1/L-2/I-1 with their code locations;
- the CLEF-010 readiness statement and the live-evidence state.

Also record the "absence" behavior: what a caller sees when no provider is
selected, and when an unknown provider is requested.

### Dependencies

None

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-001-selectability-baseline.md`

### Acceptance Criteria

- The selection mechanism and its exact surface are recorded with file/symbol
  evidence.
- Existing provider defaults and selection behavior are recorded.
- Enablement/config defaults and the disabled-Clef behavior are recorded.
- Absence/unknown-selection behavior is recorded.
- M-1/L-1/L-2/I-1 are recorded with their code locations.
- No production code is changed.

### Stop Conditions

Stop if the selection mechanism or the provider-neutral boundary cannot be
established from authoritative repository evidence. Record UNKNOWN rather than
guessing.

### Production-Change Scope

None.

---

## SP-002 — Selectability Contract

Define, precisely and testably, what "selectable" means for this repository.

Define:

- the exact selection surface(s) (CLI flag and/or configuration) and their exact
  tokens;
- whether selection is CLI-driven, config-driven, env-driven, or a combination;
- that selection is explicit opt-in only and that an absent/empty value selects
  nothing;
- that selection is independent of enablement (`CLEF_ENABLED`) and of
  configuration defaults;
- the default-preservation rule for all existing providers;
- the **rollback** rule: how Clef is disabled immediately (deselect + disable)
  without changing SOP policy;
- the **observability** rule: the minimum logging/metrics a selectable provider
  must surface, without exposing secrets, prompts, or model internals;
- the neutral-boundary rule: no Clef-specific branch/type enters `decision`, and no
  provider-specific policy enters `agentic-sop`.

The contract MUST be stated so that SP-005/SP-006 can implement and test it
mechanically, and so that SP-010 can judge it.

### Dependencies

- SP-001

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-002-selectability-contract.md`

### Acceptance Criteria

- The selection surface(s) and exact tokens are defined.
- Opt-in and absence-of-selection semantics are defined.
- Default-preservation is defined for all existing providers.
- The rollback procedure is defined.
- The minimum observability is defined and is secret/model-internal safe.
- The neutral-boundary rule is restated as a testable contract.
- No implementation is performed.

### Stop Conditions

Stop if the contract cannot be expressed without a policy, lifecycle, approval,
authorization, or routing change.

### Production-Change Scope

None.

---

## SP-003 — Security-Hardening Decision

Decide explicitly, for each carried-forward finding, whether it is:

```text
PRECONDITION_TO_SELECTION
ACCEPTABLE_WITH_CONDITION
POST_SELECTION_HARDENING
NOT_APPLICABLE
```

For each of M-1, L-1, L-2, I-1, record: the concrete code evidence, the impact
under selectability, the classification, and any condition. Do not silently waive
any finding. If a precondition is declared, name the minimal change and its
scope.

### Dependencies

- SP-001
- SP-002

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-003-hardening-decision.md`

### Acceptance Criteria

- M-1, L-1, L-2, and I-1 each receive exactly one classification with rationale and
  code evidence.
- Any `PRECONDITION_TO_SELECTION` finding names the minimal, in-scope change.
- No finding is silently waived.
- No production code is changed by this task.

### Stop Conditions

Stop if a finding's disposition cannot be established from evidence. Record
UNKNOWN rather than guessing.

### Production-Change Scope

None.

---

## SP-004 — Runtime-Readiness Decision

Decide whether live Clef evidence is required before selectability, and record the
distinction between:

- **implementation readiness** (the adapter/seam code is correct and tested);
- **selection readiness** (the selection surface is explicit, opt-in, and
  default-preserving);
- **runtime quality readiness** (the live Clef backend serves valid, fail-closed
  decisions).

Record the current live-evidence state (live Clef `/v1/systemone`
NOT_EXERCISED_LIVE/UNAVAILABLE; oMLX NOT_EXERCISED) and state whether that absence
blocks selectability, and why. Do not fabricate live results. Where the plan
requires live evidence and it is unavailable, record `NOT_EXERCISED` /
`UNAVAILABLE` with the reason.

### Dependencies

- SP-001
- SP-002
- SP-003

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-004-runtime-readiness-decision.md`

### Acceptance Criteria

- The three readiness dimensions are distinguished and each is stated.
- An explicit decision is recorded: is live evidence required before
  selectability (yes/no), with rationale.
- Live-evidence absence is recorded as UNAVAILABLE/NOT_EXERCISED, never as a
  measured zero or as success.
- If live evidence is required, the SP-008 obligation is stated explicitly.

### Stop Conditions

Stop if the live-runtime requirement cannot be decided without changing the
provider contract or SOP policy.

### Production-Change Scope

None.

---

## SP-005 — Minimal Selection Implementation

Implement the minimal selection surface defined by SP-002, within the adapter/CLI
only.

Boundaries:

- selection MUST be explicit and opt-in; an absent/empty value MUST NOT select
  Clef;
- selecting Clef MUST NOT enable it (enablement stays behind `CLEF_ENABLED`);
- every existing provider default MUST remain unchanged;
- no Clef-specific branch/type may enter the `decision` package;
- no provider-specific policy may enter `agentic-sop`;
- the disabled/rollback path MUST be implemented per SP-002;
- if SP-003 declared preconditions, implement only those minimal hardenings.

If SP-002 or SP-004 concludes selection should not be exposed, this task records
that outcome and makes no selection change.

### Dependencies

- SP-002
- SP-003
- SP-004

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-005-selection-implementation.md`
- `cmd/sop-decision-adapter/main.go` (selection surface, if SP-002 requires it)

### Acceptance Criteria

- The selection surface matches SP-002 exactly.
- Default and existing-provider behavior are unchanged.
- Disabled Clef cannot be selected accidentally.
- Enablement remains separate from selection.
- No Clef-specific code exists in `decision`.
- Build/test/vet pass.
- No `agentic-sop` change.

### Stop Conditions

Stop if implementing the contract would require a policy, lifecycle, approval,
authorization, routing, or default change.

### Production-Change Scope

Adapter/CLI selection surface and tests only.

---

## SP-006 — Selection and Default-Preservation Tests

Add focused tests proving, at minimum:

- explicit selection of Clef works through the supported surface;
- the default provider remains unchanged when nothing is selected;
- a disabled Clef provider cannot be selected accidentally;
- absence/empty/unknown selection selects nothing (or fails closed);
- selecting Clef does not enable it and does not affect other providers;
- failure paths remain fail-closed (unavailable, timeout, malformed, invalid
  choice, unsupported operation, provider failure).

### Dependencies

- SP-005

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-006-selection-tests.md`
- `cmd/sop-decision-adapter/clef_selection_test.go` (extended as required)

### Acceptance Criteria

- Each behavior above has a deterministic, model-free test.
- Tests do not require a live runtime.
- Default-preservation and no-accidental-selection are explicitly asserted.
- Build/test/vet pass.

### Stop Conditions

Stop if a required behavior cannot be tested deterministically without a live
runtime.

### Production-Change Scope

Tests only.

---

## SP-007 — Governance and Neutrality Verification

Verify that selection does not weaken governance or neutrality:

- selection does not bypass approval, lifecycle, or authorization;
- selection does not modify SOP policy semantics;
- the `decision` package remains provider-neutral (no Clef identifier/type/branch);
- no provider-specific policy exists in `agentic-sop`;
- selection is independent of SMALL/MEDIUM/LARGE routing;
- the rollback and observability rules from SP-002 hold (no secrets/model internals
  exposed).

### Dependencies

- SP-005
- SP-006

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-007-governance-neutrality-verification.md`

### Acceptance Criteria

- Each governance/neutrality property is verified with test or inspection evidence.
- Neutrality is proven by the existing `governance_test.go` neutrality tests (not
  by `go build` alone).
- The rollback path is verified.
- No unresolved Critical/High governance finding.

### Stop Conditions

Stop if any governance/neutrality property cannot be verified from evidence.

### Production-Change Scope

None.

---

## SP-008 — Live/Runtime Verification (if required)

If SP-004 determined that live evidence is required before selectability, perform
opt-in live verification of the selectable path against the Clef runtime over the
CLEF-002-selected transport (`POST /v1/systemone`, `clef-flash`).

The live check MUST be disabled by default, explicitly opted in, and skip cleanly
when the runtime is absent. It MUST validate contract shape (choice membership,
confidence bounds, classified failures) and MUST NOT assert a fixed model verdict.

If SP-004 determined live evidence is NOT required, record that decision and its
basis, and record the live state as NOT_EXERCISED / UNAVAILABLE without fabrication.

### Dependencies

- SP-004
- SP-005

### Requires

- Go build/test toolchain
- (conditionally) Ollama SystemOne decision endpoint serving Clef

### Deliverables

- `docs/reports/clef-selectable/SP-008-runtime-verification.md`
- `internal/providers/clef/live_selectable_test.go` (opt-in; only if required)

### Acceptance Criteria

- The live check (if added) is opt-in and skips cleanly without a runtime.
- If live evidence is required and unavailable, it is recorded as
  NOT_EXERCISED / UNAVAILABLE with the reason.
- No fabricated results.
- Failure remains fail-closed.

### Stop Conditions

Stop if the live check cannot remain purely observational and fail-closed.

### Production-Change Scope

Opt-in tests and report only.

---

## SP-009 — Full Verification and Security Review

Run the full repository verification and security review of the selectable-provider
work:

```bash
gofmt -l .
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
go build ./...
git diff --check
```

Review the selection surface for command/argument handling, secret/environment
handling, malformed/error/timeout handling, and fail-closed behavior. Record every
result with exact commands and outcomes. No task may be marked complete with a
failing required gate or an unresolved Critical/High finding.

### Dependencies

- SP-005
- SP-006
- SP-007
- SP-008

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-009-verification-security-review.md`

### Acceptance Criteria

- All required gates are run with exact commands and outcomes recorded.
- Selection-surface security is reviewed with concrete evidence.
- The architecture guards (neutral seam, no policy, no routing coupling,
  fail-closed) are each verified.
- No unresolved Critical/High finding.

### Stop Conditions

Any required gate failure or any Critical/High finding blocks completion.

### Production-Change Scope

Verification and tests only.

---

## SP-010 — Final Selectable-Provider Decision

Produce the final decision for whether Clef may become an explicitly selectable
provider. The result MUST be exactly one of:

```text
SELECTABLE_READY
SELECTABLE_READY_WITH_CONDITIONS
NOT_SELECTABLE_READY
```

The report MUST record: selection semantics, default-preservation, the architecture
boundary, governance, the per-finding dispositions (M-1/L-1/L-2/I-1), the runtime
decision (implementation vs selection vs runtime-quality readiness), failure
behavior, observability, rollback, the test boundary, and residual risks.

The report MUST confirm: Clef remains OFF; Clef remains non-default; no
`agentic-sop` change; SMALL/MEDIUM/LARGE routing unchanged; making Clef selectable,
live, or default requires separate human authorization.

If the result is not `NOT_SELECTABLE_READY`, recommend exactly one next action:

```text
Authorize Clef as an explicitly selectable provider.
```

This plan does not enable Clef, make Clef live, or make Clef default.

### Dependencies

- SP-009

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-selectable/SP-010-selectable-provider-decision.md`

### Acceptance Criteria

- Exactly one of the three classifications is returned.
- Every required element above is recorded with evidence.
- The confirmations (OFF / non-default / no agentic-sop change / routing unchanged /
  separate authorization) are stated.
- If not `NOT_SELECTABLE_READY`, exactly one next action is recommended.
- No production enablement occurs.

### Stop Conditions

Return `NOT_SELECTABLE_READY` if Clef cannot preserve the architecture or governance
guards, or if a precondition remains unmet.

### Production-Change Scope

None.

---

## Execution Policy

Execute along the dependency graph. The graph is acyclic:

```text
SP-001
  |
SP-002 ---------------+
  |                   |
SP-003 ---+           |
  |       |           |
SP-004 ---+           |
  |       |           |
  +--> SP-005 <-------+
         |
       SP-006
         |
       SP-007 ----+
         |        |
SP-008 --+        |
(if required)     |
         |        |
       SP-009 <---+
         |
       SP-010
```

Human review is required between major stages, especially:

```text
SP-004 -> SP-005
SP-009 -> SP-010
```

Do not:

- skip a blocked gate;
- claim completion without evidence;
- overwrite unrelated user changes;
- modify `agentic-sop`;
- change the provider-neutral contract;
- change SOP governance, policy, approval, or lifecycle semantics;
- change SMALL/MEDIUM/LARGE routing;
- introduce Clef-specific types into the public provider contract;
- change any provider default;
- make Clef live or default;
- push unless explicitly approved.

---

## Expected End State

A successful plan leaves `sop-decision-adapters` with an explicit, opt-in
selectability decision for Clef in which:

- Clef remains OFF and non-default;
- every existing provider default is unchanged;
- selection stays behind the provider-neutral `decision.Provider` seam;
- no provider-specific policy exists in `agentic-sop`;
- failure remains fail-closed;
- M-1/L-1/L-2/I-1 each have an explicit, evidence-based disposition;
- the live-runtime requirement is explicitly decided;
- rollback and minimum observability are defined;
- the final result is exactly one of `SELECTABLE_READY`,
  `SELECTABLE_READY_WITH_CONDITIONS`, or `NOT_SELECTABLE_READY`.

Making Clef selectable, live, or default beyond the plan's decision requires a
separate step and explicit human approval.
