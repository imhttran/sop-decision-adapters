# sop-decision-adapters

A standalone **decision-model adapter layer** for [`agentic-sop`](https://github.com/imhttran/agentic-sop).

It exposes **one stable decision API** while allowing different decision-model
implementations to sit behind it. Providers only _evaluate_ decisions — they never
define or enforce SOP policy.

## The one hard rule

> **Providers do NOT define or enforce SOP policy.**
> A provider evaluates a `DecisionRequest` and returns a normalized
> `DecisionResult`.

`agentic-sop` remains responsible for policy, approval gates, execution
decisions, workflow state, and safety enforcement. This repository is therefore
deliberately **not** coupled to `agentic-sop` internals, and no provider-specific
type (Ollama, SystemOne, `noul`, ONNX, Julia, CLM, JEV) appears in the shared
contract. Details: [decision contract](docs/specs/decision-contract.md).

## Status

| Phase | Scope                                                 | State       |
| ----- | ----------------------------------------------------- | ----------- |
| 0     | Foundation — contract, CLI, fixtures, docs            | ✅ complete |
| 1.1   | Native Nimble / SystemOne hardening (`/v1/systemone`) | ✅ complete |
| 2     | Julia via ONNX (adapter + Runner seam)                | 🚧 partial  |
| 3     | CLM, JEV/OpenJEV                                      | ⏳ planned  |
| 4     | Shadow evaluation / comparison                        | ⏳ planned  |

Phase 2 ships the Julia adapter, configuration, a configurable `Runner` seam,
offline tests, and CLI selection. The concrete in-process ONNX runtime binding
remains deferred. Roadmap and exit criteria: [`docs/plans/PLAN.md`](docs/plans/PLAN.md).

## Architecture

```text
agentic-sop ──▶ decision/ (public contract) ──▶ internal/providers/nimble ──▶ POST /v1/systemone
                                              └ internal/providers/julia  ──▶ Runner ──▶ ONNX inference
```

The contract lives in the **public** `decision` package, importable by another Go
module:

```go
import "github.com/imhttran/sop-decision-adapters/decision"
```

Layering, responsibilities, and the SOP boundary:
[architecture overview](docs/architecture/OVERVIEW.md).

## Providers

| Provider | Backend                         | Config                          | Spec |
| -------- | ------------------------------- | ------------------------------- | ---- |
| `nimble` | Ollama `/v1/systemone`          | `OLLAMA_BASE_URL`, `NIMBLE_*`   | [spec](docs/specs/nimble-systemone.md) |
| `julia`  | ONNX via a configurable Runner   | `JULIA_*`                       | [spec](docs/specs/julia-adapter.md) |

## Layout

```text
decision/                     PUBLIC provider-neutral contract
internal/providers/nimble/    Nimble adapter (/v1/systemone transport, translation, provider)
internal/providers/julia/     Julia adapter (Runner seam, translation, config, provider)
cmd/sop-decision-adapter/     minimal CLI (decide)
tests/fixtures/               example decision shapes + golden Nimble response + Julia fixtures
docs/                         categorized documentation (index: docs/README.md)
```

## Quick start

Requires Go 1.22+.

```sh
go test ./...                                   # fully offline; no Ollama, ONNX, or Julia required
go build ./cmd/sop-decision-adapter

export OLLAMA_BASE_URL=http://localhost:11434   # remote backends are supported
export NIMBLE_MODEL=nimble

./sop-decision-adapter decide -provider nimble \
  -file tests/fixtures/risk-evaluation.json -json
```

The Julia provider selects the external inference command via
`JULIA_INFERENCE_CMD` (whitespace-split argv) and forwards `JULIA_MODEL_PATH`;
with no command configured it reports unavailable.

```sh
export JULIA_INFERENCE_CMD="my-onnx-runner --model julia.onnx"
./sop-decision-adapter decide -provider julia \
  -file tests/fixtures/risk-evaluation.json -json
```

Full walkthrough: [getting started](docs/guides/getting-started.md). Every flag:
[CLI reference](docs/reference/cli.md). Environment:
[configuration reference](docs/reference/configuration.md).

## Tests

`go test ./...` never requires Ollama, ONNX, Julia, or the network. The opt-in
live test validates the Nimble contract against a real backend:

```sh
NIMBLE_INTEGRATION_TEST=1 go test ./...
```

The Julia suite is fully offline; an opt-in live path runs only when
`JULIA_INTEGRATION_TEST=1` and `JULIA_INFERENCE_CMD` are both set. See the
[testing guide](docs/guides/testing.md).

## Documentation

Start at the [documentation index](docs/README.md). Highlights:

- [Requirements (PRD)](docs/requirements/PRD.md) — problem, goals, users, success criteria.
- [Architecture](docs/architecture/OVERVIEW.md) — layering, responsibilities, boundary.
- [Decision contract](docs/specs/decision-contract.md) — normative public API.
- [Nimble / SystemOne provider](docs/specs/nimble-systemone.md) — normative mapping.
- [Julia (ONNX) provider](docs/specs/julia-adapter.md) — normative Runner/normalization mapping.
- [Roadmap](docs/plans/PLAN.md) · [Backlog](docs/plans/BACKLOG.md).

## License

Apache-2.0. See [`LICENSE`](LICENSE).
