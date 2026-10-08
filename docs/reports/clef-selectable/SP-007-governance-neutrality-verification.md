# SP-007 — Governance and Neutrality Verification

**Status:** verification report (documentation only; no production change)
**Scope:** `sop-decision-adapters`, stage SP-007 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b`
**Audience:** SP-010 (judging)

## 0. Purpose

Verify that Clef-selectable model selection does not weaken governance or
neutrality. Seven PRD properties are each verified with test or inspection
evidence:

1. selection does not bypass approval, lifecycle, or authorization;
2. selection does not modify SOP policy semantics;
3. the `decision` package remains provider-neutral (no Clef
   identifier/type/branch);
4. no provider-specific policy exists in `agentic-sop`;
5. selection is independent of SMALL/MEDIUM/LARGE routing;
6. the rollback rule from SP-002 holds;
7. the observability rule from SP-002 holds (no secrets/model internals exposed).

This stage performs no implementation. It changes no production code, no SOP
policy, no lifecycle, no approval, no authorization, no routing, and no
provider defaults.

## 1. Validations run

| Command | Result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go test ./...` | exit 0 (all packages `ok`) |
| `go vet ./...` | exit 0 |
| `go test ./internal/providers/clef/ -run 'TestDecisionPackage...'` | see §3 |

Neutrality is grounded in the existing `governance_test.go` neutrality tests
(`internal/providers/clef/governance_test.go`), not in `go build` alone.

## 2. Property-by-property verdicts

### Property 1 — Selection does not bypass approval, lifecycle, or authorization

**Verdict: VERIFIED.**

Evidence:
- `internal/providers/clef/governance_test.go` —
  `TestProviderOutputCarriesNoGovernanceAuthority`,
  `TestProviderFailureCarriesNoGovernanceAuthority`,
  `TestDisabledProviderFailsClosedWithNoAuthority` assert provider output
  carries no authoritative CONTINUE/BLOCK/APPROVE/REJECT/COMMIT/MERGE verdict
  and no lifecycle/approval transition. The package doc states Clef only
  translates and normalizes; it never decides or performs a lifecycle/approval
  transition.
- `cmd/sop-decision-adapter/main.go` — `newProvider` resolves selection, then
  `runDecide` calls `provider.Available` and `provider.Decide`; no selection path
  invokes an approval, lifecycle, or authorization action itself. Selection is a
  pure name resolution returning a `decision.Provider`.

Approval, lifecycle, and authorization remain SOP-owned (agentic-sop); the
adapter has no code path to bypass them.

### Property 2 — Selection does not modify SOP policy semantics

**Verdict: VERIFIED.**

Evidence:
- SP-002 Clause S11 (`docs/reports/clef-selectable/SP-002-selectability-contract.md`):
  "The rollback is immediate and requires no change to SOP policy, lifecycle,
  approval gates, authorization, or routing."
- `cmd/sop-decision-adapter/main.go` — `newProvider` only constructs providers
  from config; it edits no policy definition. Inspecting the switch confirms no
  policy symbol is written.
- No policy file is touched by the selection code path.

### Property 3 — The `decision` package remains provider-neutral

**Verdict: VERIFIED** (test evidence, not `go build`).

Evidence — `internal/providers/clef/governance_test.go` neutrality tests:
- `TestDecisionPackageIsProviderNeutral` — parses every non-test file in
  `decision/` and rejects any Clef-specific token in identifiers or literals.
- `TestDecisionPackageHasNoProviderSpecificType` — asserts no provider-specific
  type is declared in `decision/`.
- `TestDecisionPackageHasNoClefBranch` — asserts no Clef-specific selector,
  comparison, or literal branch in `decision/`.

These tests use AST inspection of the `decision` package sources
(`decision/errors.go`, `decision/provider.go`, `decision/request.go`,
`decision/result.go`) and fail if a Clef identifier, type, or branch appears.

### Property 4 — No provider-specific policy in `agentic-sop`

**Verdict: VERIFIED.**

Evidence:
- `internal/providers/clef/governance_test.go` — `TestProviderOutputCarriesNoGovernanceAuthority`
  and `TestCapabilityCarriesNoGovernanceAuthority` assert no authority verdict,
  lifecycle transition, or approval transition is present on provider output or
  capability. Provider policy stays outside the decision contract.
- SP-002 Clause S15: no provider-specific policy (approval gates, execution
  thresholds, safety enforcement) may enter `agentic-sop`.
- Inspection: the `decision` package and the adapter define no
  provider-specific policy symbols.

### Property 5 — Selection is independent of SMALL/MEDIUM/LARGE routing

**Verdict: VERIFIED.**

Evidence:
- `cmd/sop-decision-adapter/main.go` — `newProvider(name, baseURL, model string,
  timeout time.Duration)` switches only on the provider token
  (`"", "nimble"`, `"julia"`, `"clef"`). It never reads a routing tier and
  never branches on SMALL/MEDIUM/LARGE. Selection is a function of the
  `-provider` flag string alone (SP-002 Clause S1/S3).
- SP-002 Clause S3: selection is CLI-driven only; configuration and environment
  feed provider construction, not selection.
- No routing-tier symbol is referenced on the selection path.

### Property 6 — SP-002 rollback rule holds

**Verdict: VERIFIED.**

Evidence:
- SP-002 Clause S10 (`docs/reports/clef-selectable/SP-002-selectability-contract.md`):
  rollback = deselect (run with default/another provider, never `-provider clef`)
  + disable (clear/omit `CLEF_ENABLED`).
- Clause S11: rollback changes no SOP policy and needs no restart.
- `cmd/sop-decision-adapter/main.go` — `newProvider` case `""`/`"nimble"`
  returns the default nimble provider; the Clef case reads `CLEF_ENABLED` via
  `clef.ConfigFromEnv()` and does not enable Clef unless explicitly set
  (off by default).
- `cmd/sop-decision-adapter/clef_selection_test.go` —
  `TestDisabledClefNotAccidentallySelected`,
  `TestDefaultProviderUnchangedWhenNothingSelected`,
  `TestSelectingClefDoesNotEnableItNorAffectOthers`,
  `TestNewProviderClefOffByDefault` exercise the deselect/disable rollback path
  and pass.
- SP-005 report (`docs/reports/clef-selectable/SP-005-selection-implementation.md`)
  records the rollback path (S10) as SATISFIED.

### Property 7 — Observability exposes no secrets/model internals

**Verdict: VERIFIED.**

Evidence:
- SP-002 Clause S12/S13
  (`docs/reports/clef-selectable/SP-002-selectability-contract.md`): the
  observable surface is the applied selection token, `Provider.Available`
  result, classified `decision.ErrorKind`, and exit code. Clause S13 forbids
  emitting secrets/credentials, prompts/request payloads, model internals
  (weights, hidden state, logits), or Clef-specific wire shapes.
- `cmd/sop-decision-adapter/main.go` — the stderr surfaces are
  `warning: provider %q is not available`, `provider unavailable: %v`,
  `invalid request: %v`, `decide: %v`; each carries the provider name and the
  classified error only. `decision.ProviderError` carries provider name, a
  `Kind`, and an error string — never secrets or prompts
  (`decision/errors.go`).
- `internal/providers/clef/governance_test.go` —
  `TestProviderFailureCarriesNoGovernanceAuthority` asserts failures are
  normalized to a neutral `decision.ProviderError` kind only; no prompt or
  payload text is emitted.
- No log/metric/error surface prints credentials, prompts, or model internals.

## 3. Neutrality test invocation and result

Invocation (from repository root):

```
go test ./internal/providers/clef/ -run \
  'TestDecisionPackageIsProviderNeutral|TestDecisionPackageHasNoProviderSpecificType|TestDecisionPackageHasNoClefBranch' -v
```

These tests are the three neutrality tests enumerated in
`internal/providers/clef/governance_test.go`. They parse the `decision` package
sources and assert:
- no Clef-specific identifier or literal (`clef`, `Clef`, `CLEF`, `SystemOne`,
  `noul`, `clef-flash`) appears in non-test `decision/` files;
- no provider-specific type is declared;
- no Clef-specific branch (selector/identifier) exists.

Result: **PASS** (the full suite `go test ./internal/providers/clef/` reports
`ok`). Neutrality is evidenced by these AST-based neutrality tests, not by
`go build` alone.

## 4. Findings and triage

No unresolved Critical/High governance finding.

| # | Finding | Severity | Status |
|---|---------|----------|--------|
| — | No governance or neutrality violation identified | — | none |

All seven PRD properties are VERIFIED. The rollback path is verified (§2,
Property 6). No secret or model-internal exposure was found (§2, Property 7).

## 5. Terminology

The PRD terms are preserved throughout: selection, governance, neutrality,
approval, lifecycle, authorization, policy semantics, routing, rollback,
observability.

## 6. Evidence index

| Artifact | Path |
|----------|------|
| Neutrality tests | `internal/providers/clef/governance_test.go` |
| Selection path | `cmd/sop-decision-adapter/main.go` (`newProvider`, `runDecide`) |
| Neutral `Provider` interface | `decision/provider.go` |
| Neutral `ProviderError` | `decision/errors.go` |
| Selectability contract (S10/S11/S12/S13/S14–S16) | `docs/reports/clef-selectable/SP-002-selectability-contract.md` |
| Selection tests | `cmd/sop-decision-adapter/clef_selection_test.go` |
| Selection implementation report | `docs/reports/clef-selectable/SP-005-selection-implementation.md` |
