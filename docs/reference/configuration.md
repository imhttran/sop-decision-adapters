# Configuration Reference

**Type:** Reference

The Nimble provider is configured entirely through the environment (see
[`.env.example`](../../.env.example)). The Ollama server is **not** assumed to be
local.

## Environment variables

| Variable                   | Default                  | Meaning                                                        |
| -------------------------- | ------------------------ | -------------------------------------------------------------- |
| `OLLAMA_BASE_URL`          | `http://localhost:11434` | Ollama-compatible backend root (may be remote).                |
| `NIMBLE_MODEL`             | `nimble`                 | Model served by the backend.                                   |
| `NIMBLE_TIMEOUT`           | `30s`                    | Per-request timeout (Go duration syntax).                      |
| `NIMBLE_INTEGRATION_TEST`  | unset                    | When set, enables the opt-in live integration test.            |

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
- [Testing guide](../guides/testing.md) — the `NIMBLE_INTEGRATION_TEST` variable.
