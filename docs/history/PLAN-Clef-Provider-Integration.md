# PLAN --- Clef Provider Integration

**Repository:** `~/agentic-workspace/projects/sop-decision-adapters`\
**Scope:** `sop-decision-adapters` only (adapter-side).\
**Status:** PLANNED.\
**Prerequisite:** `agentic-sop` Provider-Neutral Decision Integration Seam
(SEAM-001..SEAM-008) is closed at **READY_WITH_NOTES** with:

- CRITICAL findings: 0
- HIGH findings: 0
- MEDIUM findings: 0
- blocking residuals: 0

The integration baseline is:

`agentic-sop` HEAD:

`11539138225a84a3bd476600cb1d1d8acabffc21`

Authoritative inputs:

- `agentic-sop/docs/reports/decision-integration/SEAM-002-decision-contract.md`
- `agentic-sop/docs/reports/decision-integration/SEAM-008-readiness-decision.md`
- `agentic-sop/docs/reports/clef-readiness/AS-CLEF-008-adapter-contract.md`
- `docs/reports/clef-provider/CLEF-architecture-analysis.md`

---

## Project

Clef Provider Integration

## Summary

Implement a Clef decision provider in `sop-decision-adapters` that satisfies the
already-proven provider-neutral decision contract exposed by `agentic-sop`
without introducing Clef-specific policy, lifecycle, approval, authorization,
or execution-routing behavior into `agentic-sop`.

The provider is reached through the existing in-process provider-neutral
`decision.Provider` seam, selected through the existing `newProvider`
construction-time switch in `cmd/sop-decision-adapter/main.go`:

```text
agentic-sop (or any caller)
    |
    | in-process decision.Provider
    v
internal/providers/clef
    |
    | provider-specific transport
    v
Clef runtime (/v1/systemone)
```

A generic `sop-decision-adapters serve` + provider-neutral JSON process boundary is
OUT OF SCOPE for this plan and may be handled by a future provider-runtime plan.

The current evidence supports Ollama `POST /v1/systemone` with `clef-flash` as
the primary verified Clef transport.

`mlx-community/clef-4bit` through oMLX remains an evaluation target and MUST NOT
be treated as equivalent unless CLEF-002 demonstrates fidelity-preserving Clef
decision semantics.

Clef remains OFF and non-default throughout this plan.

This plan may establish readiness for a later selectable-provider review, but
does not itself authorize making Clef selectable, live, or default.

---

## Governing Invariants

> Providers produce evidence. SOP determines what that evidence means.

> Providers evaluate. SOP governs.

Provider output is evidence only.

Clef MUST NOT gain authority to:

- transition task state;
- transition plan state;
- approve or reject work;
- satisfy human approval;
- authorize execution;
- authorize commit;
- authorize merge;
- bypass validation;
- bypass review;
- select SOP lifecycle outcomes;
- alter SMALL/MEDIUM/LARGE execution routing.

---

## Capabilities

### Go build/test toolchain — EXISTS

Evidence: `go version go1.27.1 darwin/arm64`; baseline `go build ./...`, `go vet ./...`, and `go test ./...` pass offline at the planning baseline.

Owner: operator-supplied development environment.

---

### agentic-sop provider-neutral decision seam — EXISTS

Evidence: `agentic-sop` HEAD `11539138225a84a3bd476600cb1d1d8acabffc21` provides the provider-neutral request/result boundary, fail-closed validation, SOP-owned evidence interpretation, a single policy path through `autonomy.Decide`, monotonic escalation, authoritative human approval, and provider-independent governance.

Owner: `agentic-sop` (read-only for this plan).

---

### Ollama SystemOne decision endpoint serving Clef — EXISTS

Evidence: local Ollama exposes `clef-flash` with decision capability; `POST /v1/systemone` has produced bounded choice evidence with per-choice probabilities and confidence, and score and null-like operations have also been observed.

CLEF-002 MUST reproduce and formally classify these capabilities before implementation depends on them.

---

### Apple Silicon MLX runtime / oMLX with clef-4bit — UNKNOWN

Evidence: oMLX is available and exposes OpenAI-compatible generation/chat interfaces; `mlx-community/clef-4bit` is a candidate model; native Clef decision semantics equivalent to `/v1/systemone` have NOT been established.

Status: UNKNOWN

Gap: oMLX does not natively expose Clef decision semantics; whether a fidelity-preserving translation exists is not established.

No implementation may assume compatibility; the capability REQUIRES EXPERIMENT.

---

## Goal

Enable `sop-decision-adapters` to provide Clef-generated decision evidence to
the live provider-neutral `agentic-sop` seam without importing SOP internals
and without granting Clef any governance authority.

---

## Non-Goals

- Do not modify `agentic-sop`.
- Do not modify the provider-neutral contract.
- Do not change SOP policy.
- Do not change lifecycle semantics.
- Do not change approval semantics.
- Do not change authorization semantics.
- Do not change SMALL/MEDIUM/LARGE routing.
- Do not make Clef selectable, live, or default.
- Do not assume oMLX provides SystemOne semantics.
- Do not treat generic chat generation as equivalent to Clef decision semantics.
- Do not introduce Clef-specific types into the public `decision` package.
- Do not expand this project into general LLM benchmarking.

---

## Architecture Invariant

```text
agentic-sop decision seam
        |
        | in-process decision.Provider
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

The integration boundary is the existing in-process `decision.Provider` seam,
selected through the existing `newProvider` switch in
`cmd/sop-decision-adapter/main.go`. No separate `serve` command and no generic wire
layer are part of this plan.

SOP owns interpretation.

The adapter owns provider-specific transport and translation.

---

## Configuration Invariant

Transport configuration and provider enablement are separate concerns.

`CLEF_*` configuration may define:

- endpoint;
- model;
- timeout;
- transport parameters;
- provider-specific options.

Configuration defaults MUST NOT implicitly enable or select Clef.

Clef MUST require an explicit operator selection/enablement action.

Therefore:

```text
configured != enabled
enabled != default
selectable != default
```

No endpoint/model default may make Clef active by itself.

---

## Planning and Discovery Constraint

Repository discovery performed inside a task is work performed by that task,
not a prerequisite capability.

The following MUST NOT be modeled as pre-verified capabilities:

- repository inspection;
- package discovery;
- provider-interface discovery;
- request/result-shape discovery;
- configuration discovery;
- capability discovery;
- transport discovery;
- test discovery;
- report creation;
- inspection of the read-only `agentic-sop` reference repository.

Only externally supplied runtimes, permissions, tools, artifacts, or already
completed dependencies belong in `Requires`.

---

## CLEF-001 — Capture Adapter Repository Truth

Establish the exact current state of `sop-decision-adapters`.

Inspect and record each of the following from the repository itself (source is
authoritative; do not infer from documentation):

- branch;
- HEAD;
- upstream relationship;
- working-tree state;
- Go toolchain;
- build/test baseline;
- the current provider interface;
- the current request/result types;
- existing provider implementations, including Nimble and Julia where present;
- the repository's current provider registration/factory and CLI selection
  mechanism;
- configuration and provider enablement behavior;
- the current capability representation;
- current timeout/cancellation conventions;
- current process/HTTP transport abstractions;
- the existing test structure;
- the current relationship to `agentic-sop`.

Produce a current-state architecture diagram.

Every item above is a discovery target that this task inspects and records; none is
an externally supplied prerequisite. The repository working tree, its Go toolchain,
its git metadata, and its documentation already exist in this checkout: record them
as EXISTS, do not model them as MISSING or UNKNOWN, and do not create any stage that
requires a capability this task exists to discover.

### Dependencies

None

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-provider/CLEF-001-repository-truth.md`

### Acceptance Criteria

- Repository state is recorded.
- The provider interface and request/result types are identified with file/symbol evidence.
- Existing provider implementations, including Nimble and Julia where present, are inspected and recorded.
- The current provider registration/factory and CLI selection mechanism is recorded.
- Configuration and provider enablement behavior are recorded.
- Timeout/cancellation conventions are recorded.
- The existing test structure is recorded.
- The relationship to `agentic-sop` is recorded.
- A current-state diagram is included.
- Documentation/source divergence is recorded.
- No production code is changed.

### Evidence Requirements

The report must cite exact file and symbol locations.

### Stop Conditions

Stop if the provider interface or integration boundary cannot be established
from authoritative repository evidence.

Record UNKNOWN rather than guessing.

### Production-Change Scope

None.

---

## CLEF-002 — Verify Clef/oMLX Capability and Transport

Determine with evidence which Clef capabilities are available on each candidate
runtime.

Investigate:

#### Ollama / clef-flash

- choice;
- score;
- probability;
- confidence;
- null/indeterminate;
- request shape;
- response shape;
- unsupported operations;
- runtime/model-specific behavior.

#### oMLX / mlx-community/clef-4bit

Determine whether oMLX can expose equivalent Clef decision semantics.

Explicitly distinguish:

- native decision interface;
- faithful translation;
- generic generation approximation.

Generic generation alone MUST NOT be classified as equivalent.

Classify each capability as:

- SUPPORTED
- UNSUPPORTED
- UNKNOWN / REQUIRES EXPERIMENT

### Dependencies

- CLEF-001

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-provider/CLEF-002-capability-transport.md`
- bounded request/response captures for each capability claimed SUPPORTED

### Acceptance Criteria

- Ollama Clef capabilities are individually classified.
- Every SUPPORTED classification has captured evidence.
- oMLX compatibility is explicitly classified.
- No untested behavior is presented as supported.
- Whether `clef-flash` and `mlx-community/clef-4bit` represent the same base
  model is either established or recorded UNKNOWN.
- Exactly one primary transport is selected.
- The primary transport is selected based on semantic fidelity and evidence,
  not convenience.
- No production provider implementation is created.

### Evidence Requirements

Capability table plus bounded captures.

### Stop Conditions

Stop if no Clef runtime can produce bounded choice evidence containing:

- an allowed choice; and
- usable confidence semantics required by the provider mapping.

Do not fabricate decision semantics using generic chat output.

### Production-Change Scope

Reports and probe artifacts only.

---

## CLEF-003 — Define Clef-to-Provider Contract Mapping

Define the exact mapping between:

```text
SEAM request
    ->
local adapter request
    ->
Clef operation
    ->
Clef response
    ->
provider-neutral result
```

Specify:

- request mapping;
- result mapping;
- allowed-choice handling;
- confidence semantics;
- status derivation;
- unsupported semantics;
- indeterminate semantics;
- transport failure semantics;
- malformed-result semantics.

Statuses must map into the existing provider-neutral contract:

- OK
- UNSUPPORTED
- ERROR
- INDETERMINATE

Explicitly enumerate forbidden authority fields.

### Dependencies

- CLEF-002

### Requires

- Go build/test toolchain
- agentic-sop provider-neutral decision seam

### Deliverables

- `docs/reports/clef-provider/CLEF-003-adapter-contract-mapping.md`

### Acceptance Criteria

- Every provider-neutral request field has a defined mapping.
- Every supported Clef result form has a defined mapping.
- Invalid choice handling is explicit.
- Invalid confidence handling is explicit.
- Unsupported behavior is explicit.
- Indeterminate behavior is explicit.
- Runtime/model/transport failures are mapped.
- No lifecycle, approval, authorization, commit, merge, policy, scheduler, or
  execution-model authority appears in the adapter result.
- The existing provider-neutral contract can represent all required Clef
  outcomes.

### Stop Conditions

Stop if Clef requires:

- SOP-specific policy;
- lifecycle authority;
- approval authority;
- new provider-specific SOP semantics;
- a provider-neutral contract change.

### Production-Change Scope

None.

---

## CLEF-004 — Implement Clef Provider

Implement the Clef provider behind the existing local `decision.Provider`
abstraction.

Implement the provider-neutral wire layer required by the existing
`agentic-sop` process boundary.

Clef-specific implementation stays entirely adapter-side.

### Dependencies

- CLEF-003

### Requires

- Go build/test toolchain
- agentic-sop provider-neutral decision seam

### Deliverables

Expected implementation areas:

- `internal/providers/clef/clef.go`
- `internal/providers/clef/translate.go`
- `internal/providers/clef/transport.go`
- `internal/providers/clef/config.go`
- `internal/providers/clef/capability.go`
- provider-neutral wire layer under an appropriate internal package
- provider-neutral `serve` CLI/subcommand
- `docs/specs/clef-provider.md`
- configuration documentation
- `.env.example` updates where appropriate

Exact paths may be adjusted to match current repository architecture discovered
in CLEF-001.

### Acceptance Criteria

- Clef implements the existing local provider abstraction.
- Clef-specific transport types remain private/internal.
- The selected CLEF-002 transport is used.
- Unverified transports are not enabled.
- Choice must belong to the request's allowed set.
- Confidence/probabilities must be valid and bounded.
- NaN/Inf are rejected.
- Unsupported operations return UNSUPPORTED.
- Indeterminate outcomes remain indeterminate.
- Transport failures do not become successful evidence.
- The wire layer is provider-neutral.
- The wire layer contains no Clef-specific policy branch.
- No provider-specific type is added to the public `decision` package.
- Existing providers remain unchanged except where a generic registry/factory
  extension is strictly necessary.
- Clef is OFF by default.
- Transport defaults MUST NOT implicitly enable Clef.
- Provider selection requires explicit operator action.
- No `agentic-sop` change is required.
- No execution-model routing is changed.

### Evidence Requirements

Implementation diff plus a bounded request-in/result-out demonstration.

### Stop Conditions

Stop if implementation requires:

- importing `agentic-sop` internals;
- changing the provider-neutral SOP contract;
- exposing lifecycle or approval authority;
- adding Clef-specific behavior to SOP-facing policy;
- coupling Clef to SMALL/MEDIUM/LARGE routing.

### Production-Change Scope

Adapter-side provider, generic wire plumbing, and documentation only.

---

## CLEF-005 — Provider Contract and Failure Tests

Prove offline that the Clef implementation satisfies the provider-neutral
contract and fails closed.

No real model is required.

### Dependencies

- CLEF-004

### Requires

- Go build/test toolchain

### Deliverables

- Clef translation tests
- fake-transport tests
- provider-contract tests
- failure tests
- governance-boundary tests

### Acceptance Criteria

Test:

- request translation;
- response translation;
- confidence bounds;
- probability bounds;
- NaN/Inf rejection;
- allowed-choice membership;
- out-of-set choice rejection;
- unsupported capability;
- malformed response;
- empty response;
- transport error;
- timeout;
- cancellation;
- indeterminate result;
- ERROR result;
- UNSUPPORTED result;
- provider-neutral wire behavior.

Prove provider output contains no governance authority.

Provider output must not encode authoritative:

- CONTINUE
- BLOCK
- APPROVE
- REJECT
- COMMIT
- MERGE
- lifecycle transition
- approval transition

unless a literal string happens to exist as domain data, in which case it must
remain ordinary data and carry no authority.

Default:

```text
go test ./...
```

must remain fully offline.

### Stop Conditions

Stop if any failure case cannot be made fail-closed without broadening the SOP
contract or introducing provider-specific policy.

### Production-Change Scope

Tests plus narrowly required implementation hardening only.

No semantic contract widening.

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

This is a bounded test-authoring task, not an architecture-discovery task. Bind the tests to
the existing Clef production API and begin writing as soon as the authoritative inputs below
(and, only if needed, the single named precedent) have been read.

Authoritative inputs (read; do not modify):

- `internal/providers/clef/config.go` — `ConfigFromEnv`, `Config{BaseURL,Model,Timeout,Enable}`,
  `Config.Enabled`, `Config.WithDefaults`, `DefaultBaseURL`, `DefaultModel`, `DefaultTimeout`.
- `internal/providers/clef/capability.go` — `Capability`, `(*Provider).Capability`, `Supports`,
  `UnsupportedError`, `ErrUnsupported`.
- `internal/providers/clef/clef.go` — only if a `ProviderName`/`New` reference is required.

Mutation targets (create):

- `internal/providers/clef/config_test.go`
- `internal/providers/clef/capability_test.go`

Allowed precedent (at most one; optional):

- `internal/providers/nimble/config_test.go` — the `t.Setenv`-based hermetic
  `ConfigFromEnv`/`WithDefaults` table pattern. Do not survey other providers.

Execution contract:

1. Read `config.go`, then `capability.go`.
2. Read the single named precedent only if the test style is unclear.
3. Create `config_test.go`, then `capability_test.go`.
4. Run `go test ./internal/providers/clef/...`.
5. Harden only failures attributable to this task.
6. Run the required verification.
7. Finish.

After the authoritative inputs and the named precedent have been inspected, additional
repository discovery is not part of normal execution; it requires naming the exact acceptance
criterion that cannot otherwise be implemented. Do not inspect `transport.go`, `translate.go`,
`internal/wire`, `cmd/`, or other providers — those belong to other tasks. The tests must
remain offline (no network, no model).

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

This is a bounded test-authoring task, not an architecture-discovery task. Bind the tests to
the existing Clef translation API and begin writing as soon as the authoritative inputs below
(and, only if needed, the single named precedent) have been read.

Authoritative inputs (read; do not modify):

- `internal/providers/clef/translate.go` — `BuildSystemOneRequest`, `translateQuestion`,
  `instructionsFor`, `NormalizeSystemOneResponse`, `normalizeAnswer`/`normalizeChoiceAnswer`/
  `normalizeBooleanAnswer`/`normalizeScoreAnswer`, `inUnitRange`.
- `internal/providers/clef/systemone.go` — the SystemOne request/response wire types the
  translation reads and produces.
- `decision/request.go` — `DecisionRequest`/`Question` field and validation semantics the
  translation must cover.

Mutation target (create):

- `internal/providers/clef/translate_test.go`

Allowed precedent (at most one; optional):

- `internal/providers/nimble/translate_test.go` — `TestBuildSystemOneRequest*` /
  `TestNormalizeSystemOneResponse*` shape. Do not survey other providers.

Execution contract:

1. Read `translate.go`, then `systemone.go`.
2. Read the single named precedent only if the test style is unclear.
3. Create `translate_test.go`.
4. Run `go test ./internal/providers/clef/...`.
5. Harden only failures attributable to this task.
6. Run the required verification.
7. Finish.

After the authoritative inputs and the named precedent have been inspected, additional
repository discovery is not part of normal execution; it requires naming the exact acceptance
criterion that cannot otherwise be implemented. Do not inspect `transport.go`, `internal/wire`,
`cmd/`, `agentic-sop`, or unrelated providers.

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

This is a bounded test-authoring task, not an architecture-discovery task. Bind the tests to
the existing Clef transport and provider-failure mapping and begin writing as soon as the
authoritative inputs below (and, only if needed, the single named precedent) have been read.

Authoritative inputs (read; do not modify):

- `internal/providers/clef/transport.go` — the unexported `transport` interface, `httpTransport`,
  `httpStatusError`, `malformedResponseError`, `isMalformedResponse`, `isUnavailable`.
- `internal/providers/clef/clef.go` — `Provider.Decide`, `classifyTransportError`, `providerErr`
  and the fail-closed `decision.ErrorKind` mapping.
- `internal/providers/clef/systemone.go` — the wire types a fake transport returns.

Mutation targets (create):

- `internal/providers/clef/transport_test.go`
- `internal/providers/clef/clef_test.go`

Allowed precedent (at most one; optional):

- `internal/providers/nimble/transport_test.go` — the `httptest`-server + failure/classification
  pattern (`TestHTTPTransport*`, `TestClassifyTransportError`). Do not survey other providers.

Execution contract:

1. Read `transport.go`, then `clef.go`.
2. Read the single named precedent only if the fake-transport style is unclear.
3. Create `transport_test.go`, then `clef_test.go`.
4. Run `go test ./internal/providers/clef/...`.
5. Harden only failures attributable to this task (narrowly required implementation hardening
   only; no contract widening).
6. Run the required verification.
7. Finish.

After the authoritative inputs and the named precedent have been inspected, additional
repository discovery is not part of normal execution; it requires naming the exact acceptance
criterion that cannot otherwise be implemented. Do not inspect `translate.go`, `internal/wire`,
`cmd/`, `agentic-sop`, or unrelated providers.

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

This is a bounded test-authoring task, not an architecture-discovery task. It is implemented
directly from the acceptance criteria and the existing Clef/`decision` surfaces below; no
existing-provider precedent is required, and none is authorized.

Authoritative inputs (read; do not modify):

- `internal/providers/clef/clef.go` and `internal/providers/clef/capability.go` — the adapter's
  outputs (`decision.DecisionResult`, `decision.ProviderError`, `Capability`) whose content must
  carry no governance authority.
- `decision/result.go`, `decision/errors.go`, `decision/provider.go` — the public neutral
  surfaces that must contain no Clef-specific type or branch.

Mutation targets (create):

- `internal/providers/clef/governance_test.go`
- a neutrality assertion (no Clef identifier in the public `decision` package)

Execution contract:

1. Read `clef.go`, `capability.go`, and the three `decision` files above.
2. Create `governance_test.go` asserting the boundary directly from the acceptance criteria.
3. Run `go test ./internal/providers/clef/...`.
4. Harden only failures attributable to this task.
5. Run the required verification.
6. Finish.

The neutrality assertion is bounded to the `decision/` package only. No broad architecture
discovery is authorized, and the public `decision` contract must not be modified. Additional
discovery requires naming the exact acceptance criterion that cannot otherwise be implemented.

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

Verify the implemented provider against a real local Clef runtime.

The primary runtime is the transport verified in CLEF-002.

At planning time this is expected to be:

```text
Ollama
  +
clef-flash
  +
POST /v1/systemone
```

The `mlx-community/clef-4bit` + oMLX path MUST be exercised only if CLEF-002
proved a fidelity-preserving decision transport.

The task name intentionally does not claim that clef-4bit/oMLX has already been
verified.

### Dependencies

- CLEF-011
- CLEF-014

### Requires

- Go build/test toolchain
- Ollama SystemOne decision endpoint serving Clef

### Deliverables

- opt-in live integration test
- `docs/reports/clef-provider/CLEF-006-local-runtime-verification.md`

### Acceptance Criteria

The live test:

- is disabled by default;
- requires explicit opt-in;
- skips cleanly when local runtime prerequisites are absent;
- validates contract shape rather than fixed model judgment;
- verifies returned choice belongs to allowed choices;
- verifies confidence bounds when confidence is provided.

Record:

- request success rate;
- latency p50;
- latency p95;
- choice validity;
- confidence availability;
- timeout behavior;
- malformed/indeterminate behavior;
- repeatability.

If oMLX/clef-4bit is not viable, record:

```text
NOT EXERCISED — see CLEF-002
```

Do not imply clef-4bit verification occurred.

### Stop Conditions

Stop if the selected primary runtime cannot produce valid provider-neutral
evidence.

### Production-Change Scope

Opt-in tests and report only.

---

## CLEF-007 — Shadow Evaluation Harness

Build an adapter-side evaluation harness for observational comparison.

The harness may compare:

- Clef;
- deterministic expected outcomes;
- recorded SOP outcomes;
- Nimble where semantically comparable.

Shadow evaluation MUST NOT influence the primary decision.

### Dependencies

- CLEF-006

### Requires

- Go build/test toolchain

### Deliverables

- adapter-side shadow/evaluation harness
- harness tests
- `docs/reports/clef-provider/CLEF-007-shadow-evaluation.md`

### Acceptance Criteria

Record:

- agreement rate;
- disagreement cases;
- invalid-choice rate;
- indeterminate rate;
- latency p50/p95;
- timeout/failure rate;
- confidence distribution.

A shadow:

- timeout;
- error;
- disagreement;
- malformed response

must never alter the primary result.

No shadow output is fed into SOP policy.

### Stop Conditions

Stop if shadow evaluation cannot remain purely observational.

### Production-Change Scope

Evaluation tooling only.

No live-path semantic change.

---

## CLEF-008 — Benchmark and Compare Decision Quality

Run a focused decision-provider benchmark.

Primary matrix:

- Clef via verified primary transport;
- Nimble where semantically comparable.

Conditional matrix:

- `mlx-community/clef-4bit` via oMLX only if CLEF-002 proved faithful decision
  semantics;
- Clef 8-bit only if needed as a quantization/fidelity comparison.

### Dependencies

- CLEF-007

### Requires

- Go build/test toolchain

### Deliverables

- benchmark harness
- benchmark corpus
- `docs/reports/clef-provider/CLEF-008-benchmark.md`

### Acceptance Criteria

Measure:

- choice validity;
- agreement with expected outcomes;
- confidence availability;
- confidence calibration where ground truth exists;
- latency;
- timeout rate;
- failure rate;
- indeterminate rate.

Quantization/fidelity impact must be recorded where measurable.

This task MUST NOT become a general LLM benchmark.

Benchmark results MUST NOT modify SOP policy.

### Stop Conditions

Stop if meaningful comparison would require changing:

- provider contract;
- SOP policy;
- governance;
- execution-model routing.

### Production-Change Scope

Benchmark/evaluation tooling only.

---

## CLEF-009 — Full Verification and Security Review

Perform complete repository verification and architecture review.

### Dependencies

- CLEF-008

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-provider/CLEF-009-verification.md`

### Acceptance Criteria

- All required gates (`gofmt -l .`, `go vet ./...`, `go test -count=1 ./...`, `go test -race -count=1 ./...`, `go build ./...`, `git diff --check`) are run with exact commands and outcomes recorded.
- Each architecture guard is verified and recorded with test or inspection evidence.
- The security review covers transport invocation, endpoint/URL validation, environment and secret handling, bounded output, and fail-closed error handling.
- The verification record `docs/reports/clef-provider/CLEF-009-verification.md` is produced.
- No task is marked complete with a failing required gate or an unresolved Critical or High finding.
- No SOP policy, provider contract, governance, approval, lifecycle, or execution-model routing is changed.

### Required Gates

Run:

```bash
gofmt -l .
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
go build ./...
git diff --check
```

All required gates must pass.

### Architecture Guards

Verify:

1. No Clef-specific SOP policy exists.
2. Provider/model name does not drive SOP policy.
3. Clef cannot transition lifecycle state.
4. Clef cannot create or satisfy approval.
5. Clef cannot authorize execution.
6. Clef cannot authorize commit or merge.
7. Invalid evidence fails closed.
8. Provider failure cannot produce authorization.
9. Providers are reached only through the existing in-process `decision.Provider`
   seam; no separate serve/wire architecture is introduced.
10. oMLX/Ollama/SystemOne details remain adapter-side.
11. Decision-provider selection remains independent of SMALL/MEDIUM/LARGE
    routing.
12. Clef remains OFF and non-default.

### Security Review

Review:

- shell/command injection;
- use of `sh -c`;
- URL/endpoint validation;
- argument handling;
- environment leakage;
- secret leakage;
- unbounded response handling;
- malformed JSON;
- timeout/cancellation cleanup;
- goroutine/process leaks;
- diagnostic injection;
- provider output being interpreted as an instruction.

### Stop Conditions

Any required gate failure blocks completion.

Any Critical/High architecture or security finding blocks readiness.

### Production-Change Scope

Verification hardening and tests only.

No policy or contract widening.

---

## CLEF-010 — Selectable-Provider Readiness Review

Produce the final adapter readiness decision.

This task determines whether the Clef adapter is ready to enter a separate
**selectable-provider review**.

It does NOT authorize making Clef selectable.

It does NOT enable Clef.

It does NOT make Clef live.

It does NOT make Clef default.

Final task result must be exactly one of:

- READY
- READY_WITH_NOTES
- NOT_READY

Interpretation:

#### READY

The adapter is ready to enter a separate selectable-provider review.

#### READY_WITH_NOTES

The adapter is ready to enter a separate selectable-provider review, with
non-blocking notes that must be carried forward.

#### NOT_READY

Adapter-side deficiencies remain.

### Dependencies

- CLEF-009

### Requires

- Go build/test toolchain

### Deliverables

- `docs/reports/clef-provider/CLEF-010-readiness-decision.md`

### Acceptance Criteria

The report must explicitly answer:

> Can Clef now be considered READY_FOR_SELECTABLE_REVIEW through the existing
> provider-neutral seam without requiring Clef-specific policy, lifecycle,
> approval, authorization, or execution-routing changes in agentic-sop?

Record:

- repository state;
- CLEF-001..CLEF-016 evidence;
- verified transport;
- in-process provider-neutral seam (`decision.Provider` selected via `newProvider`);
- failure semantics;
- governance boundary;
- verification results;
- security findings;
- benchmark findings;
- residual risks;
- readiness result.

The report must confirm:

- Clef remains OFF;
- Clef remains non-default;
- no `agentic-sop` change is required;
- SMALL/MEDIUM/LARGE routing remains unchanged;
- making Clef selectable requires separate human approval.

If READY or READY_WITH_NOTES, recommend exactly one next action:

```text
Open selectable-provider review.
```

Do not enable Clef.

### Stop Conditions

Return NOT_READY if Clef cannot preserve the architecture/governance guards.

### Production-Change Scope

None.

---

## Execution Policy

Execute along the decomposed graph:

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
(provider registration)      (config + capability tests)
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
                CLEF-006
                   |
                CLEF-007
                   |
                CLEF-008
                   |
                CLEF-009
                   |
                CLEF-010

CLEF-005  (historical BLOCKED evidence of the superseded broad task; its definition
           and execution history are preserved. It has no remaining dependents and
           does not block continuation; its remaining scope is carried by the
           decomposed tasks CLEF-016/CLEF-012/CLEF-013/CLEF-014.)
```

Use truthful lifecycle states:

- PLANNED
- ACTIVE
- LOCAL_DONE
- BLOCKED
- NOT_REQUIRED

Do not:

- skip a blocked gate;
- claim completion without evidence;
- overwrite unrelated user changes;
- amend unrelated commits;
- modify `agentic-sop`;
- change the provider-neutral contract;
- change SOP governance;
- change SOP policy;
- change approval behavior;
- change SMALL/MEDIUM/LARGE routing;
- introduce Clef-specific types into the public provider contract;
- assume oMLX semantics;
- make Clef selectable;
- make Clef live;
- make Clef default;
- push unless explicitly approved.

Human review is required between major stages, especially:

```text
CLEF-002 -> CLEF-003
CLEF-003 -> CLEF-004
CLEF-009 -> CLEF-010
```

---

## Expected End State

A successful plan leaves `sop-decision-adapters` with a Clef provider that:

- implements the existing adapter-side provider abstraction;
- is reached through the existing in-process provider-neutral `decision.Provider`
  seam (selected via `newProvider`);
- imports no SOP internal package;
- emits evidence only;
- validates choice and confidence;
- fails closed;
- keeps Clef/Ollama/oMLX/SystemOne details adapter-side;
- leaves SOP governance unchanged;
- leaves approval semantics unchanged;
- leaves lifecycle semantics unchanged;
- leaves SMALL/MEDIUM/LARGE routing unchanged;
- remains OFF;
- remains non-default;
- has been locally verified;
- has been shadow-evaluated;
- has been benchmarked;
- has passed full verification/security review;
- is, at most:

```text
READY_FOR_SELECTABLE_REVIEW
```

Making Clef selectable, live, or default requires a separate plan, separate
evidence, and explicit human approval.
