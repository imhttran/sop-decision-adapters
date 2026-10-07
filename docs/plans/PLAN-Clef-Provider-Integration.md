# PLAN --- Clef Provider Integration

**Repository:** `~/agentic-workspace/projects/sop-decision-adapters`\
**Scope:** `sop-decision-adapters` only (adapter-side).\
**Status:** PLANNED.\
**Prerequisite:** `agentic-sop` Provider-Neutral Decision Integration Seam
(SEAM-001..SEAM-008) is closed at **READY_WITH_NOTES** (0 CRITICAL, 0 HIGH,
0 MEDIUM, 0 blocking residuals; L1--L6 are non-blocking notes). The cross-module
contract is `agentic-sop` `docs/reports/decision-integration/SEAM-002-decision-contract.md`;
the readiness decision is `SEAM-008-readiness-decision.md`; the adapter-side
handoff is `docs/reports/clef-readiness/AS-CLEF-008-adapter-contract.md`. The
design input is this repository's
`docs/reports/clef-provider/CLEF-architecture-analysis.md`.

## Project

Clef Provider Integration

## Summary

Implement a Clef decision provider in `sop-decision-adapters` that satisfies the
already-proven, provider-neutral decision contract exposed by `agentic-sop`
(SEAM-002), without any Clef-specific policy, lifecycle, approval, authorization,
or execution-routing change in `agentic-sop`. The provider is reached over
`agentic-sop`'s provider-neutral JSON process boundary: the adapter exposes a
provider-neutral `serve` wire surface that reads the SEAM-002 request DTO on
stdin and writes the SEAM-002 result DTO on stdout, and implements Clef behind the
existing local `decision.Provider` abstraction. The design selects the **verified**
Ollama `POST /v1/systemone` Clef decision transport (live-verified with
`clef-flash`: choice + per-choice probabilities + confidence) as the primary
transport, and treats the task-named `mlx-community/clef-4bit` + oMLX path as an
**evaluation target / REQUIRES EXPERIMENT** (oMLX exposes only OpenAI-compatible
generation/chat, not Clef's decision semantics). The work proceeds through ten
sequential stages: repository truth (CLEF-001), Clef/oMLX capability and transport
verification (CLEF-002), adapter-to-provider contract mapping (CLEF-003), provider
implementation (CLEF-004), provider-contract and failure tests (CLEF-005), local
runtime verification (CLEF-006), shadow evaluation harness (CLEF-007), benchmark
and decision-quality comparison (CLEF-008), full verification and security review
(CLEF-009), and the selectable-provider readiness decision (CLEF-010). Clef is
OFF and non-default throughout; making it selectable or live requires separate
evidence and human approval.

## Capabilities

### Go build/test toolchain — EXISTS

- Evidence: `go version` returns `go version go1.27.1 darwin/arm64`; `go build ./...`, `go vet ./...`, and `go test ./...` all pass offline at HEAD `a419ce5`.
- Owner: operator-supplied development environment
- Location: Go executable available on `PATH`

### agentic-sop provider-neutral decision seam (SEAM-002) — EXISTS

- Evidence: `agentic-sop` HEAD `11539138225a84a3bd476600cb1d1d8acabffc21` contains the seam (`internal/decision`, `internal/decision/command`, `internal/cli/decision_provider.go`); `SEAM-008-readiness-decision.md` records READY_WITH_NOTES; `SEAM-002-decision-contract.md` fixes the request/result DTO and validation ownership.
- Owner: `agentic-sop` (external reference repository; read-only here)
- Location: `~/agentic-workspace/agentic-sop`

### Ollama SystemOne decision endpoint serving Clef — EXISTS

- Evidence: Ollama `0.40.0` reachable at `http://localhost:11434`; `POST /v1/systemone` with `model: "clef-flash"` returned `{"type":"choice","choice":"HIGH","probabilities":{...},"confidence":0.6630}` and, for a second probe, `{"type":"score",...}` and `{"type":"noul","noul":0.8176}`; `ollama show clef-flash` reports capability `decision`.
- Owner: operator-supplied local runtime
- Location: `http://localhost:11434` (Ollama)

### Apple-Silicon MLX runtime (oMLX) with clef-4bit — UNKNOWN

- Evidence: `omlx` `0.7.0` is installed and describes itself as an OpenAI-compatible server (no `/v1/systemone` route); `mlx-community/clef-4bit` is referenced in the HF cache but its weights are not materialized locally; whether a fidelity-preserving translation exists is not established.
- Owner: operator-supplied local runtime and model artifact
- Location: `/opt/homebrew/bin/omlx`; `~/.cache/huggingface/hub/models--mlx-community--clef-4bit`

## Goal

Make it possible for the live `agentic-sop` decision seam to consume a **Clef**
decision provider from `sop-decision-adapters` as bounded **evidence**, without
importing any `agentic-sop` Go type and without Clef gaining policy, lifecycle,
approval, commit, or merge authority.

Governing invariant:

> Providers produce evidence. SOP determines what that evidence means.

and

> Providers evaluate. SOP governs.

## Non-Goals

- Do not implement Clef in `agentic-sop` or modify `agentic-sop` in any way.
- Do not change the provider-neutral contract, SOP governance, SOP policy, or the
  SMALL/MEDIUM/LARGE execution-model routing.
- Do not add Clef, oMLX, Ollama, SystemOne, or any provider-specific branch to
  `agentic-sop` policy.
- Do not make Clef default, live, or selectable under this plan.
- Do not redesign the proven provider-neutral seam; L1--L6 are non-blocking.
- Do not assume oMLX exposes Ollama `/v1/systemone`, and do not treat
  OpenAI-compatible chat output as equivalent to Clef's decision semantics.
- Do not introduce provider-specific types into the public `decision` package.
- Do not expand into general LLM benchmarking.

## Architecture Invariant

No `agentic-sop` subsystem depends on Clef, and no Clef output influences a SOP
state transition. The desired boundary:

```text
agentic-sop decision seam  (internal/decision/command, provider-neutral JSON)
        |  request DTO (stdin)                     result DTO (stdout)
        v
sop-decision-adapters  serve  (provider-neutral SEAM-002 wire layer)
        |  local decision.Provider
        v
internal/providers/clef  (Clef transport, translation, capability handling)
        v
Ollama POST /v1/systemone  (verified)   |   oMLX OpenAI-compat (experiment only)
        v
Clef  —  evidence only
```

Provider output is evidence; SOP validation and `autonomy` policy determine what
it means. With the capability disabled or no provider configured, existing SOP
behavior is unchanged, and Clef remains OFF.

## Planning and discovery constraint

Repository inspection and discovery performed by tasks in this plan are work to be
performed by the task, not prerequisite capabilities. The following MUST NOT be
modeled as required capabilities that must already be VERIFIED before a task may
execute: repository/package inspection; provider-interface discovery;
adapter/transport discovery; request/result-shape discovery; configuration/default
discovery; capability/error discovery; test-location discovery; report creation or
inspection of this repository or of the read-only `agentic-sop` reference.

These discovery targets may begin with status UNKNOWN; the task converts UNKNOWN
into evidence-backed findings. A task may require a prerequisite capability only
when it represents an actual externally supplied runtime, permission, tool,
service, artifact, or completed dependency. Required report/plan paths are task
outputs, not prerequisite capabilities. Completed artifacts (including the
`agentic-sop` SEAM-002 contract) are evidence inputs, not capabilities requiring
rediscovery.

---

## CLEF-001 — Capture Adapter Repository Truth

Establish the exact current state of `sop-decision-adapters` before any change:
record branch, HEAD, upstream relationship, working-tree state, Go toolchain, and
the current build/test baseline; identify the provider interface, existing
adapters (Nimble, Julia), provider registration/factory, configuration,
request/result types, capability representation, timeout/cancellation
conventions, tests, command/process/HTTP transport abstractions, and the
repository's relationship to `agentic-sop`. Produce a concrete current-state
diagram. Do not infer architecture from old documentation when source disagrees.

### Dependencies

None

### Requires

- Go build/test toolchain

### Deliverables

- docs/reports/clef-provider/CLEF-001-repository-truth.md — current-state trace, boundary inventory, and diagram (evidence record)

### Acceptance Criteria

- Branch, HEAD, upstream relationship, working-tree state, Go toolchain, and the current build/test baseline are recorded.
- The provider interface, request/result types, and error model are identified with file/symbol evidence.
- Existing adapters (Nimble, Julia) and their transport abstractions are identified with file/symbol evidence.
- Provider selection/registration, configuration, capability representation, and timeout/cancellation conventions are identified.
- The tests (offline and opt-in) and the relationship to `agentic-sop` are recorded.
- A current-state diagram is produced, and any divergence between documentation and source is recorded.
- No production change is made; uncommitted user changes are preserved and not modified.

### Evidence requirements

The current-state record and diagram citing file/symbol locations at HEAD `a419ce5`.

### Stop conditions

Stop and report if the provider interface or the `agentic-sop` relationship cannot
be established from authoritative in-repository evidence; record UNKNOWN with the
remaining gap rather than asserting.

### Production-change scope

None (read-only trace and report).

---

## CLEF-002 — Verify Clef/oMLX Capability and Transport

Determine, with evidence, which Clef capabilities are actually available for the
target runtimes and which transport is viable. Investigate separately: the
Ollama-served Clef (`clef-flash`) over `POST /v1/systemone` and the
task-named `mlx-community/clef-4bit` served through oMLX. For each capability
(choice; score; probability/confidence; null/indeterminate handling; supported
request format; supported response format; model/runtime-specific behavior)
classify SUPPORTED, UNSUPPORTED, or UNKNOWN/REQUIRES EXPERIMENT. Determine
explicitly whether `mlx-community/clef-4bit` + oMLX can expose the decision
semantics needed, or whether (as the design analysis indicates) oMLX provides only
generic generation/chat semantics and a translation must be either demonstrated
faithful or rejected. Do not invent compatibility.

### Dependencies

- CLEF-001

### Requires

- Go build/test toolchain

### Deliverables

- docs/reports/clef-provider/CLEF-002-capability-transport.md — a per-capability classification (SUPPORTED/UNSUPPORTED/UNKNOWN) for both runtimes, the oMLX compatibility conclusion, and the selected primary transport with rationale
- Bounded probe captures (request/response) for every capability claimed SUPPORTED

### Acceptance Criteria

- Each of choice, score, probability/confidence, null/indeterminate handling, request format, and response format is classified for the Ollama-served Clef path with a captured probe as evidence.
- The oMLX path is classified explicitly: whether it exposes Clef decision semantics natively, whether a translation is faithful, or whether it is not viable for decision evaluation.
- No compatibility is asserted without a capture; any unexercised behavior is recorded UNKNOWN/REQUIRES EXPERIMENT.
- Whether `clef-flash` (Ollama) and `mlx-community/clef-4bit` are the same base model is recorded as established or UNKNOWN.
- Exactly one primary transport is selected with an evidence-based rationale; the design report's selection is confirmed or corrected with evidence.
- No production provider code is written; probes are bounded and read-only.

### Evidence requirements

The capability report with probe captures and the explicit oMLX compatibility conclusion.

### Stop conditions

Stop and report if no Clef runtime can produce bounded choice evidence (choice and
confidence) at all; do not fabricate decision semantics from generic generation
output.

### Production-change scope

None in production packages; bounded probe inputs/captures and the report only.

---

## CLEF-003 — Define Clef-to-Provider Contract Mapping

Fix the exact, field-by-field mapping from the SOP-proven SEAM-002 contract to the
adapter, and from the adapter to Clef, and back. Specify: the SEAM-002 request DTO
to the local adapter request and Clef operation; the Clef result to the
provider-neutral evidence (SEAM-002 result DTO); the status derivation
(`OK`/`UNSUPPORTED`/`ERROR`/`INDETERMINATE`); the confidence semantics; and the
forbidden content that must never cross the boundary (lifecycle transitions,
approval, authorization, commit/merge/scheduler authority, policy outcomes, and
execution-model/class selection). Define the failure mapping for every
Clef-specific failure into the generic provider contract, and confirm whether the
SOP contract can represent each failure without new SOP policy semantics.

### Dependencies

- CLEF-002

### Requires

- Go build/test toolchain
- agentic-sop provider-neutral decision seam (SEAM-002)

### Deliverables

- docs/reports/clef-provider/CLEF-003-adapter-contract-mapping.md — the request/result mapping tables, status/confidence rules, forbidden-content list, and the failure-mapping table

### Acceptance Criteria

- Every SEAM-002 request field is mapped to an adapter field and a Clef operation, with ownership noted.
- Every Clef result form (choice, no-confidence, out-of-set choice, invalid confidence, indeterminate, unsupported, error) is mapped to a SEAM-002 result with status and exit behavior.
- The forbidden content (lifecycle/approval/authorization/commit/merge/policy/execution-model authority) is explicitly enumerated and absent from the mapping.
- The failure-mapping table covers runtime unavailable, model unavailable, endpoint unavailable, timeout, cancellation, malformed response, invalid choice, invalid confidence, unsupported operation, indeterminate result, transport failure, and model load failure.
- The report states whether the generic provider contract can represent each failure without new SOP policy semantics, or records an explicit contract gap.
- The mapping keeps provider-specific types inside `sop-decision-adapters` and changes neither the local contract nor the SOP contract.

### Evidence requirements

The mapping document with the request, result, status, and failure tables.

### Stop conditions

Stop and report if Clef requires provider-specific SOP policy, lifecycle
authority, or approval authority, or if a real Clef failure cannot be represented
by the SOP contract and would require modifying `agentic-sop`.

### Production-change scope

None (contract/mapping document only).

---

## CLEF-004 — Implement Clef Provider

Implement the Clef provider behind the existing local `decision.Provider`
abstraction, plus the provider-neutral SEAM-002 wire layer that `agentic-sop`'s
process boundary reaches. The provider owns Clef request construction, the
selected transport, Clef response parsing, capability detection, provider-specific
pre-return validation, conversion into provider-neutral evidence, and
provider-specific diagnostics. The wire layer is provider-neutral (it accepts any
local provider) so that another provider can satisfy the same contract with no SOP
policy change. Clef is OFF at the wire layer: nothing is invoked unless the
capability is explicitly enabled and a provider selected. No Clef-specific branch,
identity, or type is introduced into the public `decision` package, into Nimble or
Julia, or into any SOP-facing policy.

### Dependencies

- CLEF-003

### Requires

- Go build/test toolchain
- agentic-sop provider-neutral decision seam (SEAM-002)

### Deliverables

- internal/providers/clef/clef.go — Provider wired to a transport and translation, implementing `decision.Provider`
- internal/providers/clef/translate.go — DecisionRequest/DecisionResult ⇄ Clef/SystemOne translation (choice-oriented), with confidence and allowed-choice enforcement
- internal/providers/clef/transport.go — `Transport` interface plus the selected Clef transport (private wire types)
- internal/providers/clef/config.go — `CLEF_*` environment configuration with defaults and CLI overrides
- internal/providers/clef/capability.go — capability/support rules including UNSUPPORTED and INDETERMINATE derivation
- internal/evidence/ — provider-neutral SEAM-002 request/result wire layer (request DTO → local request; local result/error → result DTO)
- cmd/sop-decision-adapter — a provider-neutral `serve` subcommand that selects a provider by name and speaks the SEAM-002 DTO on stdin/stdout
- docs/specs/clef-provider.md — the normative Clef adapter spec; updates to architecture, configuration reference, CLI reference, and `.env.example`

### Acceptance Criteria

- `internal/providers/clef` implements `decision.Provider` (`Name` returns a stable identifier, `Available` performs no inference, `Decide` validates, transports, and normalizes).
- The selected transport is the primary transport from CLEF-002; no unverified transport is enabled by default.
- Confidence and probabilities are validated against `[0,1]`; a choice outside the allowed set is rejected (fail closed) and never emitted as evidence.
- Unsupported kinds/operations produce an explicit UNSUPPORTED result; the adapter never returns a low-confidence guess as a stand-in.
- The `serve` subcommand reads the SEAM-002 request DTO on stdin and writes the SEAM-002 result DTO on stdout, and exits non-zero on a transport failure (no usable result).
- The `serve` wire layer accepts any local provider; it contains no Clef-specific branch.
- No provider-specific type appears in the public `decision` package; Nimble and Julia are unchanged.
- The capability is OFF by default; constructing the Clef provider is a no-op unless selected and enabled.
- No change is made to `agentic-sop`, its policy, or the SMALL/MEDIUM/LARGE execution-model routing.

### Evidence requirements

The implementation diff, the adapter spec, and a configuration/CLI update; a
manual `serve` invocation demonstrating request-in/result-out against a local
runtime or a fake transport.

### Stop conditions

Stop and report if implementing Clef requires importing an `agentic-sop` package,
exposing lifecycle/approval authority in the DTO, or changing the public
`decision` contract or `internal/model` routing.

### Production-change scope

New adapter-side code only: `internal/providers/clef`, `internal/evidence`, and the
CLI `serve` subcommand, plus documentation. No change to `decision/`, to Nimble or
Julia, to `agentic-sop`, or to execution-model routing.

---

## CLEF-005 — Provider Contract and Failure Tests

Prove, with tests that require no real model, that Clef produces exactly the
provider-neutral contract already proven with `agentic-sop`'s fake external
provider, and that every failure is fail-closed. Test request translation,
response translation, confidence/probability bounds, allowed choices, unsupported
capability, malformed responses, and error translation; prove the adapter requires
no SOP internal type, exposes no governance authority, and requires no
provider-specific policy.

### Dependencies

- CLEF-004

### Requires

- Go build/test toolchain

### Deliverables

- Unit tests for `internal/providers/clef` (translation, bounds, allowed choices, unsupported capability, malformed responses, error translation) using a fake transport
- Provider-contract tests for `internal/evidence` proving the SEAM-002 result DTO is produced for each status and validates under the SEAM-002 rules
- Failure/governance tests proving the adapter cannot emit a governance action and that invalid/failed results fail closed

### Acceptance Criteria

- Unit tests exist and pass for request translation, response translation, confidence/probability `[0,1]` bounds (including NaN/Inf rejection), allowed-choice membership, unsupported capability, malformed responses, and error translation.
- Provider-contract tests prove the result DTO matches SEAM-002 for OK, UNSUPPORTED, ERROR, and INDETERMINATE, and that a second provider satisfies the same wire unchanged.
- Tests prove the adapter output contains no `CONTINUE`/`BLOCK`/`APPROVE`/`REJECT`/`COMMIT`/`MERGE` or lifecycle/approval/authorization value.
- Tests prove an out-of-set choice, out-of-range confidence, malformed response, or transport failure never becomes a success.
- The default `go test ./...` suite passes fully offline with no model, network, Ollama, oMLX, or MLX runtime.

### Evidence requirements

The passing offline test suite and the specific assertions for each failure case.

### Stop conditions

Stop and report any case that cannot be made fail-closed without broadening the
contract or adding provider-specific policy.

### Production-change scope

Tests only (plus any bounded hardening the tests require); no policy-semantic
widening and no contract change.

---

## CLEF-006 — Local Clef-4bit Runtime Verification

Verify the selected Clef transport against the real local runtime. Drive the
verified Ollama SystemOne path with Clef (`clef-flash`), and additionally exercise
the `mlx-community/clef-4bit` + oMLX path only if CLEF-002 demonstrated a
fidelity-preserving transport. Measure request success rate, latency (p50/p95),
choice validity, confidence availability, timeout behavior, malformed/indeterminate
behavior, and repeatability. The live test is opt-in and SKIPS when the runtime is
not configured; it asserts the contract, not fixed judgments.

### Dependencies

- CLEF-005

### Requires

- Go build/test toolchain
- Ollama SystemOne decision endpoint serving Clef

### Deliverables

- tests/integration_clef_test.go — an opt-in (`CLEF_INTEGRATION_TEST=1`) live contract test that SKIPS when the runtime is absent
- docs/reports/clef-provider/CLEF-006-local-runtime-verification.md — measured success rate, latency, choice validity, confidence availability, timeout/malformed/indeterminate behavior, and repeatability

### Acceptance Criteria

- The live test is skipped by default and only runs when `CLEF_INTEGRATION_TEST` is set; it SKIPS (never fails the suite) when the runtime is not configured.
- The live run asserts the contract (provider available, a valid choice in the allowed set, confidence in range when present) and does not assert fixed judgments.
- Request success rate, latency p50/p95, choice validity, confidence availability, timeout behavior, malformed/indeterminate behavior, and repeatability are recorded.
- The oMLX/`clef-4bit` path is exercised only if CLEF-002 deemed it viable; otherwise the report records it as not exercised with the CLEF-002 reason.
- The default offline suite is unaffected and still requires no runtime.

### Evidence requirements

The live test, its recorded outcomes, and the measured runtime report.

### Stop conditions

Stop and report if the verified transport cannot produce valid provider-neutral
evidence on the real runtime, or if the runtime is unavailable and no bounded
alternative is available.

### Production-change scope

An opt-in test and a report; no production behavior change and no default enablement.

---

## CLEF-007 — Shadow Evaluation Harness

Build an adapter-side evaluation harness that compares Clef decisions against
deterministic expected cases, existing SOP outcomes, and optionally Nimble, and
reports disagreements. The harness is observational: it returns the primary result
only and records the shadow result; a shadow error or disagreement is captured,
never surfaced as a primary failure and never fed back into SOP policy. Because the
SOP seam provides OFF and enabled (no runtime shadow mode), shadow evaluation is
performed here, adapter-side, and Clef remains non-authoritative.

### Dependencies

- CLEF-006

### Requires

- Go build/test toolchain

### Deliverables

- An adapter-side shadow/evaluation command or harness under `internal/eval` or `tools/` that runs a primary and a shadow provider and records a disagreement report
- docs/reports/clef-provider/CLEF-007-shadow-evaluation.md — the harness design and a first evaluation record

### Acceptance Criteria

- The harness records agreement rate, disagreement cases, invalid-choice rate, indeterminate rate, latency p50/p95, timeout/failure rate, and confidence distribution.
- The harness returns the primary result only; the shadow result is observational and never returned as the decision.
- A shadow error, timeout, or disagreement never alters the primary result and never influences SOP policy.
- The harness runs offline with fakes and, optionally, against a real Clef runtime when configured.
- No SOP policy, `agentic-sop`, or execution-model routing is changed.

### Evidence requirements

The harness, its tests, and the evaluation record.

### Stop conditions

Stop and report if shadow evaluation cannot be made observational (i.e., if it
would need to affect the primary decision or SOP policy).

### Production-change scope

Adapter-side evaluation tooling and reports; no change to the live decision path
and no default enablement.

---

## CLEF-008 — Benchmark and Compare Decision Quality

Design and run a focused benchmark that compares Clef decision-provider
suitability across the supported matrix: Clef via Ollama SystemOne (primary),
Clef 4-bit via oMLX (only if CLEF-002 proved a faithful translation), optionally
Clef 8-bit as a quantization/fidelity baseline, and Nimble where comparable.
Capture choice validity, confidence availability, confidence calibration where
ground truth exists, agreement with expected/SOP outcomes, latency, and
timeout/failure rate. Do not expand into general LLM benchmarking.

### Dependencies

- CLEF-007

### Requires

- Go build/test toolchain

### Deliverables

- A benchmark harness and corpus for decision-provider suitability
- docs/reports/clef-provider/CLEF-008-benchmark.md — the matrix, metrics, and comparison, including quantization effects where measurable

### Acceptance Criteria

- The benchmark matrix covers Clef via Ollama SystemOne, and includes Nimble where comparable; it includes Clef 4-bit/oMLX only if a faithful translation was proven.
- Choice validity, confidence availability, calibration (where ground truth exists), agreement, latency, and timeout/failure rate are measured and recorded.
- Quantization/fidelity is compared where evidence supports it, or recorded as not measurable with the reason.
- The benchmark is scoped to decision-provider suitability; it does not expand into general LLM benchmarking.
- No benchmark result changes SOP policy, the contract, or execution-model routing.

### Evidence requirements

The benchmark harness, its corpus, and the comparison report.

### Stop conditions

Stop and report if the benchmark would require changing the contract, policy, or
execution-model routing, or if no comparable provider is available.

### Production-change scope

Adapter-side benchmark tooling and reports; no change to the live decision path.

---

## CLEF-009 — Full Verification and Security Review

Run the complete repository verification gates and the provider-neutrality and
governance guards, recording exact commands and outcomes; perform a security review
of the adapter-side surface. Verify: no Clef-specific policy branch; no provider
name or model name drives SOP behavior; Clef output cannot transition lifecycle
state, create an approval, or authorize a commit/merge; malformed/unknown/
indeterminate evidence fails closed; provider absence and provider failure preserve
governed behavior; decision-provider selection remains independent of
SMALL/MEDIUM/LARGE routing; another provider satisfies the same contract with no SOP
policy change; oMLX/Ollama/SystemOne details remain adapter-side; and Clef is OFF
and non-default.

### Dependencies

- CLEF-008

### Requires

- Go build/test toolchain

### Deliverables

- docs/reports/clef-provider/CLEF-009-verification.md — the verification and security-review record with exact commands and results

### Acceptance Criteria

- `gofmt -l .`, `go vet ./...`, `go test -count=1 ./...`, `go test -race -count=1 ./...`, `go build ./...`, and `git diff --check` are run with exact commands and results recorded.
- Each architecture guard (1--12) is verified and recorded with test/inspection evidence.
- Security review covers transport invocation (no shell), bounded output, environment handling, secret non-leakage, and fail-closed error handling.
- The verification record is produced; no gate is claimed passing without a recorded outcome.
- No task is marked complete with a failing required gate.

### Evidence requirements

The verification record with exact commands and outcomes, and the security-review findings.

### Stop conditions

Stop and report any failing required gate; do not mark the task complete.

### Production-change scope

Tests/guards and reports only (no policy widening, no contract change, no default
enablement).

---

## CLEF-010 — Selectable-Provider Readiness Decision

Produce exactly one readiness result (READY, READY_WITH_NOTES, or NOT_READY) for
making Clef a **selectable** decision provider in `agentic-sop`, answering
explicitly whether Clef can be selected through the provider-neutral seam without
any Clef-specific policy, lifecycle, approval, authorization, or execution-routing
change in `agentic-sop`. Do not include `DEFAULT` as a state: making Clef live,
selectable, or default is a separate decision requiring separate evidence and
human approval. If READY or READY_WITH_NOTES, name exactly one next action; if
NOT_READY, report the smallest adapter-side correction required.

### Dependencies

- CLEF-009

### Requires

- Go build/test toolchain

### Deliverables

- docs/reports/clef-provider/CLEF-010-readiness-decision.md — the final readiness report

### Acceptance Criteria

- Exactly one readiness result is produced: READY, READY_WITH_NOTES, or NOT_READY.
- The report explicitly answers whether Clef can be selectable through the provider-neutral seam without provider-specific behavior or policy authority in `agentic-sop`.
- The report includes repository state, CLEF-001..CLEF-010 task states with one-line evidence, the verified transport and wire contract, the failure semantics, the governance boundary, verification commands and outcomes, changes, the readiness decision, and exactly one next action.
- The report confirms Clef is OFF and non-default, that `decision/` and `agentic-sop` are unchanged, and that SMALL/MEDIUM/LARGE execution-model routing is unchanged.
- No `DEFAULT` state is introduced; making Clef default is explicitly deferred to separate evidence and human approval.

### Evidence requirements

The readiness report with cited evidence from CLEF-001..CLEF-009.

### Stop conditions

Stop and report if Clef cannot be shown to preserve the guards; classify NOT_READY
with the smallest bounded adapter-side correction.

### Production-change scope

None (readiness report only).

---

## Execution Policy

Include the plan-wide planning and discovery constraint in every task's planning
input. Model repository discovery in task objectives and acceptance criteria, never
as a prerequisite `Requires` entry. Preserve completed evidence inputs; do not
rediscover them as prerequisite capabilities.

Execute tasks sequentially:

```text
CLEF-001
    |
CLEF-002
    |
CLEF-003
    |
CLEF-004
    |
CLEF-005
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
```

Use truthful task states: `PLANNED`, `ACTIVE`, `LOCAL_DONE`, `BLOCKED`,
`NOT_REQUIRED`.

Do not: skip a blocked gate; claim completion without evidence; overwrite unrelated
user changes; amend unrelated commits; modify `agentic-sop`; change the
provider-neutral contract; change SOP governance, policy, or the SMALL/MEDIUM/LARGE
execution-model routing; introduce Clef-specific types into the public `decision`
package; make Clef default, live, or selectable; push unless the repository
workflow explicitly authorizes it.

## Expected End State

A successful run leaves `sop-decision-adapters` with a Clef decision provider that:

- implements the existing local `decision.Provider` abstraction;
- is reached by `agentic-sop` over the provider-neutral SEAM-002 JSON process
  boundary, through a provider-neutral `serve` wire layer;
- produces evidence only, with choice and confidence, and never a governance
  action;
- fails closed on malformed, invalid, unsupported, indeterminate, unavailable,
  timed-out, or cancelled outcomes;
- keeps all oMLX/Ollama/SystemOne/Clef details adapter-side;
- leaves `decision/`, Nimble, Julia, `agentic-sop`, and SMALL/MEDIUM/LARGE routing
  unchanged;
- is OFF and non-default, with shadow evaluation and benchmarking performed
  adapter-side;
- is, at most, READY_FOR_SELECTABLE_REVIEW — making Clef selectable, live, or
  default requires separate evidence and human approval.
