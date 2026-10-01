# Testing

**Type:** Guide

The test suite is **offline by default** — it never requires Ollama to be
running.

## Offline

```sh
go test ./...
```

Coverage:

- `decision` — question/request/result validation and the normalized error model.
- `internal/providers/nimble` — SystemOne request translation, response
  normalization (missing / unknown / out-of-range answers rejected), provider
  behavior via a fake transport, and the full HTTP path via `net/http/httptest`.
- `tests/fixtures` — shipped fixtures are valid `DecisionRequest`s, and the
  golden Nimble response is parsed through the production parser.

### Golden fixture

`tests/fixtures/nimble/systemone_response.json` is a **real** response captured
from a locally installed Ollama + Nimble instance. It is the first known-good
provider contract fixture and drives the production parser offline.

## Live integration (opt-in)

To validate the contract against a real Ollama + Nimble installation:

```sh
NIMBLE_INTEGRATION_TEST=1 go test ./...
```

The test asserts the **contract** — provider available, every requested answer
present, answer types match the questions, selected choices are allowed, and
probabilities/confidence are in range. It deliberately does **not** assert fixed
judgments such as `risk == HIGH`, because model decisions may change.

## Related

- [Nimble / SystemOne provider](../specs/nimble-systemone.md) — the behavior
  under test.
- [Configuration reference](../reference/configuration.md) — the
  `NIMBLE_INTEGRATION_TEST` variable.
