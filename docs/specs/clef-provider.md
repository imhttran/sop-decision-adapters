# Clef Provider Specification

## Overview

The Clef provider adapts the Clef decision model, served through an Ollama-compatible
backend via the `/v1/systemone` endpoint (the CLEF-002-selected transport), to the
provider-neutral `decision.Provider` abstraction used by `agentic-sop`.

Clef-specific implementation remains entirely adapter-side; no Clef-specific types
or policy enter the public `decision` package or the provider-neutral wire layer.

## Transport

- Uses the CLEF-002-selected transport: `POST /v1/systemone` to an Ollama-compatible backend.
- Unverified transports (e.g., MLX/oMLX `clef-4bit`) are not implemented and cannot be enabled.
- Transport failures are returned as errors, not successful evidence.

## Configuration

Clef is OFF by default. Explicit operator action is required to activate the provider.

Environment variables:

- `OLLAMA_BASE_URL`: Ollama-compatible backend root URL (default: `http://localhost:11434`).
- `CLEF_MODEL`: Clef decision model name (default: `clef-flash`).
- `CLEF_TIMEOUT`: Request timeout in Go duration syntax (default: `30s`).
- `CLEF_ENABLED`: Explicit enable flag. Must be set to `1`, `true`, or `yes` (case-insensitive) to enable the provider. Any other value leaves Clef disabled.

## Capabilities

The Clef provider supports the `decide` operation for:

- `decision.QuestionChoice`: mapped to SystemOne `choice` question type.
- `decision.QuestionBoolean`: mapped to SystemOne `noul` question type.
- `decision.QuestionScore`: mapped to SystemOne `score` question type.

Unsupported operations (e.g., `stream`, `batch`, `function_call`, `embeddings`, `completion`, `chat`) return `UNSUPPORTED`.

## Behavior

- Choice answers must select from the request's allowed set.
- Confidence and probabilities are validated to be within `[0, 1]`; `NaN`/`Inf` are rejected.
- Indeterminate responses (missing or undefined values) remain indeterminate and are not converted to success-shaped answers.
- The provider reports availability only when the backend is reachable and the configured model is present.
- No `agentic-sop` changes are required; execution-model routing is unchanged.

## Example

Enable Clef explicitly:

```bash
CLEF_ENABLED=1 CLEF_MODEL=clef-flash sop-decision-adapter decide -provider clef -choices APPROVE,REJECT -state "Should we proceed?"
```

With Clef disabled (default), the adapter reports unavailable and does not serve decisions.
