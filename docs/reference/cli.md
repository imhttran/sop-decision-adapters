# CLI Reference — `sop-decision-adapter`

**Type:** Reference

The CLI is a debugging and testing surface, **not** the primary integration
mechanism for `agentic-sop`.

## Synopsis

```text
sop-decision-adapter <command> [flags]
```

## Commands

| Command               | Description                                                                |
| --------------------- | -------------------------------------------------------------------------- |
| `decide`              | Evaluate a `DecisionRequest` through a provider and print the result.       |
| `help`, `-h`, `--help`| Print usage.                                                               |

## `decide` flags

| Flag           | Default                | Description                                                                 |
| -------------- | ---------------------- | --------------------------------------------------------------------------- |
| `-provider`    | `nimble`               | Provider to use. Only `nimble` is implemented today.                        |
| `-file`        |                        | Path to a JSON `DecisionRequest` (multi-question). Takes precedence over the flag form. |
| `-state`       |                        | Free-form state/context (single-question form).                             |
| `-choices`     |                        | Comma-separated allowed choices (single-question form). Required when `-file` is unset. |
| `-decision-id` | `example`              | Question ID for the single-question form.                                   |
| `-base-url`    | env `OLLAMA_BASE_URL`  | Provider base URL.                                                          |
| `-model`       | env `NIMBLE_MODEL`     | Model name.                                                                 |
| `-timeout`     | `30s`                  | Per-request timeout.                                                        |
| `-json`        | `false`                | Print the result as indented JSON.                                          |

## Exit codes

| Code | Meaning                                                                 |
| ---- | ----------------------------------------------------------------------- |
| `0`  | Success.                                                                 |
| `1`  | Provider failure (`ErrProviderFailure` / `ErrMalformedResponse`).        |
| `2`  | Invalid request, usage error, or unsupported input.                     |
| `3`  | Provider unavailable (`ErrUnavailable`).                                |

If the provider is not available, a warning is written to stderr but the decision
is still attempted.

## Examples

Multi-question request from a file:

```sh
sop-decision-adapter decide -provider nimble \
  -file tests/fixtures/risk-evaluation.json -json
```

Single choice question from flags:

```sh
sop-decision-adapter decide \
  -decision-id risk -choices LOW,MEDIUM,HIGH \
  -state "deploying a schema migration during business hours"
```

## Related

- [Configuration reference](configuration.md) — environment variables.
- [Getting started](../guides/getting-started.md) — a worked example.
