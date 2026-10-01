# Configuration Reference

**Type:** Reference

Providers are configured entirely through the environment (see
[`.env.example`](../../.env.example)).

## Nimble

The Ollama server is **not** assumed to be local.

| Variable                   | Default                  | Meaning                                                        |
| -------------------------- | ------------------------ | -------------------------------------------------------------- |
| `OLLAMA_BASE_URL`          | `http://localhost:11434` | Ollama-compatible backend root (may be remote).                |
| `NIMBLE_MODEL`             | `nimble`                 | Model served by the backend.                                   |
| `NIMBLE_TIMEOUT`           | `30s`                    | Per-request timeout (Go duration syntax).                      |
| `NIMBLE_INTEGRATION_TEST`  | unset                    | When set, enables the opt-in live integration test.            |

## Julia

| Variable               | Default | Meaning                                                                        |
| ---------------------- | ------- | ------------------------------------------------------------------------------ |
| `JULIA_MODEL`          | `julia` | Model identifier used as the reported `DecisionResult.Model` when the runner does not report one. |
| `JULIA_MODEL_PATH`     | unset   | Path to the ONNX model. Forwarded to the inference command as `JULIA_MODEL_PATH`. |
| `JULIA_INFERENCE_CMD`  | unset   | External inference command, whitespace-split into argv. Empty ⇒ provider unavailable. |
| `JULIA_TIMEOUT`        | `30s`   | Inference timeout (Go duration syntax).                                         |
| `JULIA_INTEGRATION_TEST` | unset | When set (and `JULIA_INFERENCE_CMD` is set), enables the opt-in live test.      |

The model a runner reports in `Outputs.Model` takes precedence; `JULIA_MODEL` is
reported only when the runner reports no model.

`JULIA_INFERENCE_CMD` is whitespace-split; shell quoting rules are not applied,
so an argument containing spaces is unsupported. The configured command must read
provider-neutral `Inputs` JSON on stdin and write `Outputs` JSON on stdout; see
the [Julia provider spec](../specs/julia-adapter.md).

The `decide` flags `-base-url`, `-model`, and `-timeout` override the
environment for a single invocation.

## Precedence and fallback

For a single call the precedence is: `-base-url` / `-model` / `-timeout` flags →
environment → built-in defaults. Empty or malformed values fall back to the
default; a non-positive timeout falls back to `30s`.

## Related

- [CLI reference](cli.md) — flag details.
- [Nimble / SystemOne provider](../specs/nimble-systemone.md) — availability and
  error normalization.
- [Julia (ONNX) provider](../specs/julia-adapter.md) — Runner contract and
  normalization rules.
- [Testing guide](../guides/testing.md) — the `NIMBLE_INTEGRATION_TEST` and
  `JULIA_INTEGRATION_TEST` variables.
