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
     v
internal/providers/nimble
     |   SystemOne translation (CHOICE / BOOLEAN→noul / SCORE)
     v
POST /v1/systemone
     v
Ollama
     v
Nimble
```

Dependencies point one way only: `agentic-sop` depends on the public `decision`
package, which is implemented by the providers here. This repository does **not**
import `agentic-sop`, and the `decision` package exposes no provider-specific
types (no `SystemOne`, no `noul`, no Ollama shapes).

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
| `decision`                  | The stable, public contract: `Provider`, `DecisionRequest`, `DecisionResult`, question/answer types, normalized errors. | Any Ollama/Nimble/SystemOne/`noul` type; SOP policy. |
| `internal/providers/nimble` | Translate the Nimble decision model (via `/v1/systemone`) to/from the contract.                                         | SOP policy; assumptions that the backend is local.   |
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
  [Nimble / SystemOne provider](../specs/nimble-systemone.md)
- [Requirements](../requirements/PRD.md) · [Roadmap](../plans/PLAN.md) ·
  [Backlog](../plans/BACKLOG.md)
