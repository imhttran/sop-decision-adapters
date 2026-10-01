# Getting Started

**Type:** Guide

Take a decision through a provider, from a clean checkout to a real Nimble model.

## Prerequisites

- Go 1.22+.
- Optional, for live decisions: an Ollama-compatible backend serving a Nimble
  model. Point `OLLAMA_BASE_URL` at it — it need not be local.

## 1. Build and verify offline

```sh
go fmt ./...
go vet ./...
go test ./...          # fully offline; no Ollama required
go build ./cmd/sop-decision-adapter
```

## 2. Configure the backend

```sh
export OLLAMA_BASE_URL=http://localhost:11434
export NIMBLE_MODEL=nimble
export NIMBLE_TIMEOUT=30s
```

See the [configuration reference](../reference/configuration.md) for details.

## 3. Make a decision

Multi-question request from a JSON file — one state evaluated against several
questions:

```sh
./sop-decision-adapter decide -provider nimble \
  -file tests/fixtures/risk-evaluation.json -json
```

Or a single choice question from flags:

```sh
./sop-decision-adapter decide \
  -decision-id risk -choices LOW,MEDIUM,HIGH \
  -state "deploying a schema migration during business hours"
```

The result is a normalized `DecisionResult` with one answer per question. The
adapter never applies policy: it returns choices and probabilities, and
`agentic-sop` decides what to do with them. See the
[decision contract](../specs/decision-contract.md) for the exact shape.

## Next steps

- [Testing guide](testing.md) — offline tests and the opt-in live test.
- [CLI reference](../reference/cli.md) — every flag and exit code.
- [Decision contract](../specs/decision-contract.md) — the API in detail.
