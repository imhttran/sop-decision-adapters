# Decision Contract — `decision` package

**Type:** Normative specification
**Status:** Implemented (Phase 1.1)

## Purpose

Define the stable, provider-neutral decision API that every adapter implements
and that callers — notably `agentic-sop` — depend on. This document is the single
authority for the contract's shape and rules; other documents link here instead
of restating it.

## Related Specifications

- [Nimble / SystemOne provider](nimble-systemone.md) — the reference
  implementation of this contract.
- [Architecture overview](../architecture/OVERVIEW.md) — layering and package
  responsibilities.
- [Requirements](../requirements/PRD.md) — why this contract exists.

## Normative Language

MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY are normative.

## Scope

The contract lives in the **public** `decision` package:

```go
import "github.com/imhttran/sop-decision-adapters/decision"
```

It MUST NOT contain any provider-specific type or term — no `Ollama`,
`SystemOne`, `noul`, `Nimble`, `ONNX`, `Julia`, `CLM`, or `JEV` — and MUST NOT
contain SOP policy.

## Provider interface

```go
type Provider interface {
    Name() string
    Available(ctx context.Context) bool
    Decide(ctx context.Context, req DecisionRequest) (DecisionResult, error)
}
```

- `Name` MUST return a stable identifier (for example `"nimble"`).
- `Available` MUST be cheap, MUST NOT panic, and SHOULD respect `ctx`.
- `Decide` MUST validate the request, evaluate it, and return a normalized
  `DecisionResult` whose answers correspond to the request's questions.

## DecisionRequest

```go
type DecisionRequest struct {
    State     string
    Questions []Question
    Metadata  map[string]string
}
```

- `State` is **required** and MUST be non-empty after trimming surrounding
  whitespace.
- `Questions` MUST contain at least one question, with unique IDs.
- A request that violates these rules MUST fail validation with an error matching
  `errors.Is(err, decision.ErrInvalidRequest)`.

One `Decide` call evaluates **several questions against one state**.

## Question types

```go
type QuestionType string

const (
    QuestionChoice  QuestionType = "choice"
    QuestionBoolean QuestionType = "boolean"
    QuestionScore   QuestionType = "score"
)
```

A `Question` carries `ID`, `Type`, optional `Instructions` and `Criteria`, and
optional `Choices`. Only fields appropriate to the type are required:

- **CHOICE** MUST have non-empty `Criteria` and at least one distinct, non-empty
  choice.
- **BOOLEAN** requires no provider-specific representation (no choices, criteria,
  or transport fields).
- **SCORE** is provider-neutral; `Criteria` MAY describe the scale but is not
  required by the neutral contract.

Unknown question types and duplicate question IDs MUST fail validation.

## DecisionResult

```go
type DecisionResult struct {
    Provider string
    Model    string
    Answers  map[string]Answer
    Usage    Usage
}
```

- `Answers` MUST be keyed by question ID and MUST contain an entry for every
  requested question.
- Each `Answer` MUST preserve its semantic type: CHOICE (choice + probabilities +
  optional confidence), BOOLEAN (probability only, no transport field), SCORE
  (numeric value).
- `Usage` carries provider-neutral token accounting and MAY be zero.

## Error model

Failures are normalized so callers never branch on provider strings:

```text
ProviderError{Provider, Kind, Err}
Kind ∈ { invalid_request, unavailable, malformed_response, provider_failure }
```

- `errors.Is(err, decision.ErrMalformedResponse)` — and the other sentinels
  (`ErrInvalidRequest`, `ErrUnavailable`, `ErrProviderFailure`) — MUST work
  regardless of which provider produced the error.
- `errors.Unwrap` MUST expose the underlying cause for logging.

## Boundary rules

- Providers MUST NOT define, interpret, or enforce SOP policy.
- SOP approval thresholds (for example `probability > 0.8 ⇒ approval`) MUST NOT
  appear in this repository; they belong to `agentic-sop`.
- Shadow providers MUST NOT affect the returned result (see
  [Phase 4](../plans/PLAN.md)).
