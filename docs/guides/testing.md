# Testing

**Type:** Guide

The test suite is **offline by default** — it never requires Ollama, ONNX,
Julia, or the network.

## Offline

```sh
go test ./...
```

Coverage:

- `decision` — question/request/result validation and the normalized error model.
- `internal/providers/nimble` — SystemOne request translation, response
  normalization (missing / unknown / out-of-range answers rejected), provider
  behavior via a fake transport, and the full HTTP path via `net/http/httptest`.
- `internal/providers/julia` — `BuildInputs` mapping, `NormalizeOutputs`
  normalization (missing / incompatible / unknown / out-of-range rejected),
  provider behavior via `RunnerFunc` fakes, and `CommandRunner` round-trips via
  hermetic `sh -c` commands (no ONNX or Julia runtime).
- `tests/fixtures` — shipped fixtures are valid `DecisionRequest`s; the golden
  Nimble response is parsed through the production parser; the Julia sample
  `Outputs` JSON normalizes into a valid `DecisionResult`.

### Golden fixture

`tests/fixtures/nimble/systemone_response.json` is a **real** response captured
from a locally installed Ollama + Nimble instance. It is the first known-good
provider contract fixture and drives the production parser offline.

## Live integration (opt-in)

### Nimble

To validate the contract against a real Ollama + Nimble installation:

```sh
NIMBLE_INTEGRATION_TEST=1 go test ./...
```

The test asserts the **contract** — provider available, every requested answer
present, answer types match the questions, selected choices are allowed, and
probabilities/confidence are in range. It deliberately does **not** assert fixed
judgments such as `risk == HIGH`, because model decisions may change.

### Julia

The Julia suite is offline by default: `BuildInputs`/`NormalizeOutputs` mapping,
the 2–20 option limit, provider behavior via `RunnerFunc` fakes, and
`CommandRunner` round-trips via hermetic `sh -c` commands. The helper's pure
functions have their own standard-library tests:

```sh
python3 tools/julia/test_infer.py
```

The opt-in live path runs only when `JULIA_INTEGRATION_TEST=1`. It SKIPS (never
fails the suite) when the runtime is not configured, otherwise it drives Go →
helper → ONNX Runtime → Julia-1-ONNX and asserts the contract without asserting
fixed judgments:

```sh
export JULIA_INTEGRATION_TEST=1
export JULIA_MODEL_PATH=/path/to/julia-1.onnx
go test ./...
```

`JULIA_TOKENIZER_PATH` (default `tokenizer.json` beside the model),
`JULIA_MODEL_ID`, `JULIA_MAX_TOKENS`, and `JULIA_PYTHON` are read by the helper.
`JULIA_MODEL_PATH` is forwarded to the inference command as an environment
variable. See [`tools/julia/README.md`](../../tools/julia/README.md).

## Related

- [Nimble / SystemOne provider](../specs/nimble-systemone.md) — the behavior
  under test.
- [Julia (ONNX) provider](../specs/julia-adapter.md) — the Runner contract.
- [Configuration reference](../reference/configuration.md) — the
  `NIMBLE_INTEGRATION_TEST` and `JULIA_INTEGRATION_TEST` variables.
