# Nimble / SystemOne Provider

**Type:** Normative specification
**Status:** Implemented (Phase 1.1)

## Purpose

Define how the Nimble adapter (`internal/providers/nimble`) implements the
[decision contract](decision-contract.md) against Ollama's native
`POST /v1/systemone` decision API. This document is the single authority for the
Nimble ↔ SystemOne mapping, availability, and error normalization.

## Related Specifications

- [Decision contract](decision-contract.md) — the provider-neutral contract this
  implements.
- [Architecture overview](../architecture/OVERVIEW.md) — package structure and
  responsibilities.
- [Configuration reference](../reference/configuration.md) — `OLLAMA_BASE_URL`,
  `NIMBLE_MODEL`, `NIMBLE_TIMEOUT`.

## Normative Language

MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY are normative.

## Transport

The adapter MUST use Ollama's native `POST /v1/systemone` decision API. There
MUST NOT be a prompt-based `/api/generate` decision path or Markdown-fence
extraction.

```text
decision.DecisionRequest
      │  BuildSystemOneRequest  (structured translation, no prompting)
      ▼
POST /v1/systemone  ──▶  Ollama  ──▶  Nimble
      │  NormalizeSystemOneResponse (validate answers against questions)
      ▼
decision.DecisionResult
```

The transport handles HTTP, JSON encode/decode, timeouts, status handling, and
context cancellation only. It MUST NOT contain SOP policy.

## Question → SystemOne translation

| `decision.Question` | SystemOne request                                     | SystemOne response         | `decision.Answer` |
| ------------------- | ----------------------------------------------------- | -------------------------- | ----------------- |
| `CHOICE`            | `type: "choice"` (criteria as an options map)         | `type: "choice"`           | `AnswerChoice`    |
| `BOOLEAN`           | `type: "noul"` (synthesized non-empty instructions)   | `type: "noul"` + `noul`    | `AnswerBoolean`   |
| `SCORE`             | `type: "score"` (criteria as an array of candidates)  | `type: "score"` + `score`  | `AnswerScore`     |

### Instructions

The adapter MUST always send non-empty `instructions` (SystemOne rejects an empty
value with HTTP 400). It prefers the caller's `Instructions`, then `Criteria`,
then a synthesized fallback derived from the question ID.

### CHOICE criteria

SystemOne REQUIRES a CHOICE question's `criteria` to be an object mapping each
allowed option key to a description or null; a plain string is rejected. The
adapter emits one nil-valued key per allowed choice, e.g. `{"LOW":null,"HIGH":null}`.

### SCORE candidate requirement

SystemOne REQUIRES a SCORE question's `criteria` to be an array of 2–26 candidate
descriptions; a string or a 0/1-element array is rejected with HTTP 400. Because
the provider-neutral contract has no first-class score-scale field yet, the
adapter carries candidates in the question's `Choices` and MUST return
`decision.ErrInvalidRequest` when fewer than two candidates are supplied. A
first-class score-scale representation is deferred (see
[backlog](../plans/BACKLOG.md)).

### BOOLEAN ↔ `noul` boundary

SystemOne represents a yes/no decision as `type: "noul"` with a `noul`
probability. The public API exposes BOOLEAN with a provider-neutral probability.
The translation MUST happen **entirely inside** `internal/providers/nimble`:

```text
decision.BOOLEAN  →  SystemOne "noul"  →  Nimble  →  noul probability  →  decision.AnswerBoolean
```

The `noul` probability is **provider output**. It MUST NOT be treated as an SOP
approval policy: whether a probability is sufficient to trigger an approval gate
is decided by `agentic-sop`. The adapter MUST contain no thresholds or gating.

## Response normalization

The adapter MUST reject, as `decision.ErrMalformedResponse`:

- a missing answer for any requested question;
- an answer whose type is unknown or incompatible with the question type;
- a CHOICE answer that selects a choice outside the allowed choices, or carries a
  probability key outside the allowed choices;
- a probability or confidence outside `[0, 1]`;
- a BOOLEAN answer without a `noul` probability;
- a SCORE answer without a value;
- an undecodable response body.

Probability **sums** MUST NOT be constrained; only key membership and the
`[0, 1]` range are enforced.

## Availability

`Available(ctx)` MUST distinguish the following, without issuing an inference
request (using the model listing endpoint `GET /api/tags`):

- Ollama unreachable or request timeout → `false`;
- Ollama reachable but the configured model absent → `false`;
- Ollama reachable and the configured model present → `true`.

## Transport error normalization

| Underlying condition                                              | Normalized kind     |
| ----------------------------------------------------------------- | ------------------- |
| connection refused/reset, host/network unreachable, DNS failure   | `unavailable`       |
| timeout, context deadline, context cancellation                   | `unavailable`       |
| undecodable response body                                         | `malformed_response`|
| backend non-2xx that is not a connectivity problem                | `provider_failure`  |

## Provider boundary

SystemOne wire types (`systemone.go`) MUST remain unexported and MUST NOT leak
into the public `decision` package. No `noul`, `SystemOne`, or Ollama shape may
appear in the public contract.
