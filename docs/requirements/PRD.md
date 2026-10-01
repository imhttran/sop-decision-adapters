# PRD — sop-decision-adapters

Status: Phase 0 and Phase 1.1 **COMPLETE**; Phase 2 **PLANNED**
(see the [roadmap](../plans/PLAN.md)).
Owner: agentic-sop platform.
License: Apache-2.0.

## Problem

`agentic-sop` must make **decisions** — risk level, whether approval is
required, which execution model to run — while its policy, approval gates,
workflow state, and safety enforcement stay authoritative and auditable.

Today, any specific decision model (an LLM, an ONNX model, a statistical
classifier) is entangled with the SOP engine: its endpoint, its request shape,
and its response format leak into policy code. That coupling makes it hard to
swap models, compare models, or fail over safely, and it risks letting a model
implicitly become the source of policy.

## Goals

- Provide **one stable decision API** that `agentic-sop` (or any caller) depends
  on, independent of any specific model. The API lives in the public `decision`
  package and is importable by other Go modules.
- Allow multiple decision-model implementations behind that API, added
  incrementally (Nimble → Julia → CLM → JEV).
- Normalize heterogeneous provider responses into a single `DecisionResult`.
- Keep providers **policy-free**: they evaluate a request, nothing more.
- Make providers testable without a running backend (Ollama, ONNX runtime, …).
- Keep the surface intentionally small: standard library first, no plugin
  framework, no databases, no UI.

## Non-goals

- **Defining or enforcing SOP policy.** Approval rules, thresholds, execution
  choices, and safety gates belong to `agentic-sop`, not to providers.
- Coupling this repository to `agentic-sop` internals (no import of agentic-sop
  packages, no shared policy types).
- Owning workflow state, retries orchestration, or audit trails beyond a single
  evaluate call.
- A generic plugin/discovery framework or dynamic provider loading.
- Databases, web UI, MCP, or agent orchestration.
- Implementing Julia, CLM, or JEV (see the [roadmap](../plans/PLAN.md)).

## Users

- **`agentic-sop` core** — consumes the decision API, owns policy and gates.
- **Platform / adapter engineers** — implement and operate individual providers
  (Nimble first), and later shadow comparisons.
- **Model owners** — expose a decision model behind the contract without touching
  SOP policy.
- **Operators** — run the `sop-decision-adapter` CLI to exercise a provider, and
  configure endpoints via environment.

## Decision-provider abstraction

A provider is a thin, policy-free translator:

```text
DecisionRequest ──▶ Provider ──▶ DecisionResult
```

`DecisionRequest` carries one required `State` plus one or more typed
`Questions`; `DecisionResult` carries one typed `Answer` per question plus
provider-neutral `Usage`. Failures are normalized so callers never branch on
provider-specific strings.

The interface, request/result shapes, and error model are **normative** and
defined once in the [decision contract](../specs/decision-contract.md). This PRD
does not restate them.

Implementations MUST NOT expose Ollama, SystemOne, `noul`, Nimble, ONNX, Julia,
CLM, or JEV types through the contract.

## Initial providers

| Provider          | Backend                                        | Phase |
| ----------------- | ---------------------------------------------- | ----- |
| **Nimble**        | Ollama → `/v1/systemone` → Nimble              | 1.1   |
| **Julia**         | ONNX runtime                                   | 2     |
| **CLM**           | CLM                                            | 3     |
| **JEV / OpenJEV** | OpenJEV                                        | 3     |
| **Shadow**        | wraps a primary + shadow provider (comparison) | 4     |

Nimble is the reference implementation; its `/v1/systemone` mapping, availability
check, and error normalization are specified in the
[Nimble / SystemOne provider](../specs/nimble-systemone.md) spec. There is no
prompt-based `/api/generate` decision path.

## SOP integration boundary

`agentic-sop` owns policy, approval gates, execution decisions, workflow state,
and safety. It consumes the decision API and interprets suggested choices under
its own rules; the API never tells `agentic-sop` what policy to apply. Providers
in this repository MUST NOT define or enforce policy.

The dependency/trust direction and the layering diagram live in the
[architecture overview](../architecture/OVERVIEW.md).

## Success criteria

Phase 0 and Phase 1.1 are successful when:

1. A single stable `Provider` contract exists as a **public** package containing
   no vendor types.
2. `DecisionRequest` supports multiple typed questions and `DecisionResult`
   supports multiple typed answers, all provider-neutral and validated —
   including a required non-empty `State`.
3. The Nimble adapter translates a `DecisionRequest` into a `/v1/systemone`
   request and normalizes the response; no prompt-based `/api/generate` path
   remains.
4. SystemOne-specific concepts (`noul`, request/response shapes) remain internal
   to the Nimble provider.
5. The transport is separated from translation and can be tested without Ollama.
6. Availability verifies the configured model exists without running inference;
   transport errors normalize to working `errors.Is` categories.
7. `go fmt ./...`, `go vet ./...`, and `go test ./...` pass offline;
   `NIMBLE_INTEGRATION_TEST=1 go test ./...` passes against a real backend.
8. The CLI can send a multi-question `DecisionRequest` from a JSON file.
9. Documentation states the SOP boundary and the shadow-mode constraint (shadow
   providers must never affect execution).

## Related documentation

- [Architecture overview](../architecture/OVERVIEW.md)
- [Decision contract](../specs/decision-contract.md) ·
  [Nimble / SystemOne provider](../specs/nimble-systemone.md)
- [Roadmap](../plans/PLAN.md) · [Backlog](../plans/BACKLOG.md)
