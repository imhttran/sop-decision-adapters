# Architecture Overview

**Type:** Architecture

How `sop-decision-adapters` is structured, and where the boundary with
`agentic-sop` lies. Normative behavior lives in [`../specs/`](../specs); this
document covers structure only.

## Layering

```text
agentic-sop
     |   policy · approval gates · execution decisions · workflow state · safety
     v
decision/                         (PUBLIC, importable by agentic-sop)
     |   DecisionRequest ──▶ Provider ──▶ DecisionResult
     +-------------------------------+
     v                               v
internal/providers/nimble        internal/providers/julia
     |   SystemOne translation        |   provider-neutral Inputs/Outputs
     v                               v
POST /v1/systemone               Runner (seam)
     v                               v
Ollama                           ONNX inference
     v                               v
Nimble                           Julia
```

Both providers implement the same public `decision.Provider` contract. In Julia
the concrete ONNX execution is isolated behind the `Runner` interface
(`Name`/`Available`/`Run`): the shipped `CommandRunner` runs an operator-provided
inference command, and a future in-process ONNX binding can replace it without
touching the adapter, translation, CLI, or contract.

Dependencies point one way only: `agentic-sop` depends on the public `decision`
package, which is implemented by the providers here. This repository does **not**
import `agentic-sop`, and the `decision` package exposes no provider-specific
types (no `SystemOne`, no `noul`, no Ollama/ONNX/Julia shapes).

## SOP integration boundary

- `agentic-sop` owns policy, approval gates, execution decisions, workflow state,
  and safety.
- It consumes the decision API: it asks a question and gets a normalized answer.
- The API never tells `agentic-sop` what policy to apply; `agentic-sop` decides
  how to interpret a suggested choice under its own rules.
- Providers here MUST NOT define, interpret, or enforce policy.

## Package responsibilities

| Package                     | Responsibility                                                                                                          | Must not contain                                     |
| --------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| `decision`                  | The stable, public contract: `Provider`, `DecisionRequest`, `DecisionResult`, question/answer types, normalized errors. | Any Ollama/Nimble/SystemOne/`noul`/ONNX/Julia type; SOP policy. |
| `internal/providers/nimble` | Translate the Nimble decision model (via `/v1/systemone`) to/from the contract.                                         | SOP policy; assumptions that the backend is local.   |
| `internal/providers/julia`  | Translate the Julia/ONNX decision model (behind the `Runner` seam) to/from the contract.                                | SOP policy; assumptions about the concrete runtime.  |
| `cmd/sop-decision-adapter`  | Minimal CLI to exercise providers.                                                                                      | Business logic; a CLI framework.                     |

## Nimble provider structure

`internal/providers/nimble` separates concerns so the provider is testable
without a running Ollama instance:

- **Transport** (`transport.go`) — `Transport` interface and `HTTPTransport` over
  `net/http`: HTTP, JSON encode/decode, timeouts, status handling, and context
  cancellation only.
- **SystemOne wire types** (`systemone.go`) — unexported request/response shapes,
  confined to this package.
- **Translation** (`translate.go`) — `decision.Question` ⇄ SystemOne shapes. Pure
  functions; no I/O.
- **Provider** (`nimble.go`) — wires transport and translation, validates input,
  classifies errors; implements `decision.Provider`.
- **Config** (`config.go`) — environment configuration.

Because the transport is an interface, the provider is tested with a fake (unit)
and with `net/http/httptest` (HTTP path) — never requiring a running Ollama
instance. The backend is configurable (`OLLAMA_BASE_URL`), so remote deployments
work.

The exact mapping, availability behavior, and error normalization are specified
in the [Nimble / SystemOne provider](../specs/nimble-systemone.md) spec; testing
is described in the [testing guide](../guides/testing.md); configuration in the
[configuration reference](../reference/configuration.md).

## Julia provider structure

`internal/providers/julia` mirrors Nimble's separation of concerns so the
provider is testable without any ONNX runtime:

- **Runner** (`runner.go`) — the `Runner` interface plus provider-neutral
  `Inputs`/`Outputs` wire types and the stdlib `CommandRunner`. Isolates concrete
  ONNX execution; exports no vendor vocabulary.
- **Translation** (`translate.go`) — `BuildInputs` and `NormalizeOutputs`. Pure
  functions; no I/O.
- **Provider** (`julia.go`) — wires runner and translation, validates input,
  classifies errors; implements `decision.Provider`.
- **Config** (`config.go`) — environment configuration (`JULIA_*`).

Because the runtime is behind the `Runner` interface, the provider is tested with
`RunnerFunc` fakes and hermetic `sh -c` commands — never requiring ONNX or Julia.
The exact contract is specified in the
[Julia (ONNX) provider](../specs/julia-adapter.md) spec.

## Shadow mode (Phase 4)

Shadow evaluation runs a second provider for observation only.

```text
DecisionRequest
      |
      +------+
      |      |
      v      v
   primary  shadow
      |      |
      +-- comparison
             |
             v
     disagreement report
```

Rules:

- The **primary** provider's `DecisionResult` is the only result returned.
- The **shadow** provider's result is recorded (choice mismatch, confidence
  delta, missing choices) in a disagreement report.
- A shadow error or timeout is captured, never surfaced as a failure of the
  primary decision.
- **Shadow providers MUST NOT affect execution.** They are observational.

## Non-negotiables

1. Providers never define or enforce SOP policy.
2. No provider-specific type crosses into the public `decision` package.
3. No dependency on `agentic-sop` internals.
4. Shadow providers never affect execution.
5. Standard library first; no plugin framework, databases, or UI.

## Related

- [Decision contract](../specs/decision-contract.md) ·
  [Nimble / SystemOne provider](../specs/nimble-systemone.md) ·
  [Julia (ONNX) provider](../specs/julia-adapter.md)
- [Requirements](../requirements/PRD.md) · [Roadmap](../plans/PLAN.md) ·
  [Backlog](../plans/BACKLOG.md)
