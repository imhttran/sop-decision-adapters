> **Historical — non-normative.** This is the completed Phase 1.1 implementation
> plan, kept as a record. It is not a specification and may not reflect the
> current contract; see the [documentation index](../README.md) for the current
> documents.

# Phase 1.1 — Native Nimble / SystemOne Hardening

## Status

COMPLETE

## Objective

Replace the initial prompt-based Nimble integration with Ollama's native `/v1/systemone` decision API and establish the stable public decision contract that future providers will implement.

Phase 1.1 establishes Nimble as the reference implementation for `sop-decision-adapters`.

At the end of this phase:

- Nimble uses `/v1/systemone`
- no prompt-based `/api/generate` decision path remains
- multiple questions can be evaluated against one state
- the provider-neutral decision contract is publicly importable
- SystemOne-specific details remain internal
- real Nimble responses are represented by golden fixtures
- unit tests remain fully offline
- an opt-in live integration test validates a real Ollama + Nimble installation
- `agentic-sop` can safely build against the public contract in a later integration phase

---

# 1. Scope

Phase 1.1 focuses only on hardening the Nimble provider.

Included:

1. Public decision API
2. Multi-question decision requests
3. Strongly typed question types
4. Multi-answer results
5. Native `/v1/systemone` transport
6. SystemOne request translation
7. SystemOne response normalization
8. BOOLEAN ↔ `noul` translation
9. Provider/model availability checks
10. Normalized transport errors
11. Golden response fixtures
12. Offline unit tests
13. Opt-in live integration testing
14. CLI updates
15. Documentation updates

Not included:

- Julia
- CLM
- JEV/OpenJEV
- shadow evaluation
- provider registry
- agentic-sop integration
- MCP
- database
- web UI
- generic plugin framework

---

# 2. Architecture

Target architecture:

```text
agentic-sop
     │
     │ future Go dependency
     ▼
decision/
     │
     │ provider-neutral contract
     ▼
decision.Provider
     │
     ▼
internal/providers/nimble
     │
     │ SystemOne translation
     ▼
POST /v1/systemone
     │
     ▼
Ollama
     │
     ▼
Nimble
```

Provider-specific concepts must not leak through the public API.

In particular:

```text
SystemOne
noul
Ollama request structures
Ollama response structures
```

remain implementation details of the Nimble provider.

---

# 3. Task 1 — Promote the Decision Contract

## Goal

Allow another Go module, particularly `agentic-sop`, to import the decision API.

Move:

```text
internal/decision/
```

to:

```text
decision/
```

Expected structure:

```text
decision/
├── provider.go
├── request.go
├── result.go
└── errors.go
```

Update all internal imports.

### Acceptance criteria

- `github.com/imhttran/sop-decision-adapters/decision` is importable externally.
- No Nimble/Ollama/SystemOne types appear in the public package.
- Existing normalized errors continue to work with `errors.Is`.
- `go test ./...` passes.

---

# 4. Task 2 — Introduce Provider-Neutral Question Types

## Goal

Support the kinds of decisions needed by SystemOne without exposing SystemOne terminology.

Define provider-neutral question types:

```text
CHOICE
BOOLEAN
SCORE
```

Conceptually:

```go
type QuestionType string

const (
    QuestionChoice  QuestionType = "choice"
    QuestionBoolean QuestionType = "boolean"
    QuestionScore   QuestionType = "score"
)
```

A question should contain enough information to describe:

```text
ID
Type
Instructions
Criteria
Choices
```

Only fields appropriate to the question type should be required.

### Important rule

The public API MUST NOT contain:

```text
noul
```

`noul` is a Nimble/SystemOne transport representation.

### Acceptance criteria

- CHOICE validates allowed choices/criteria.
- BOOLEAN requires no provider-specific representation.
- SCORE has a provider-neutral representation.
- Unknown question types fail validation.
- Duplicate question IDs fail validation.
- Validation has unit tests.

---

# 5. Task 3 — Multi-Question DecisionRequest

Replace the current single-question model:

```text
DecisionID
State
Choices
```

with a request capable of evaluating several questions against one state.

Conceptually:

```go
DecisionRequest {
    State
    Questions
    Metadata
}
```

Example:

```text
State:
  "A schema migration will execute against production
   during business hours."

Questions:

  risk
    type: CHOICE
    LOW
    MEDIUM
    HIGH

  approval_required
    type: BOOLEAN

  execution_model
    type: CHOICE
    SMALL
    MEDIUM
    LARGE
```

### Acceptance criteria

A single `Provider.Decide()` invocation can evaluate all three questions.

The common API remains independent of Nimble.

---

# 6. Task 4 — Multi-Answer DecisionResult

Replace the single-choice result with a result capable of representing all answers returned by a provider.

Conceptually:

```text
DecisionResult
├── Provider
├── Model
├── Answers
└── Usage
```

Each answer must preserve its semantic type.

Example:

```text
risk
  type        CHOICE
  choice      HIGH
  probability 0.9576
  confidence  0.8349

approval_required
  type        BOOLEAN
  probability 0.9974

execution_model
  type        CHOICE
  choice      LARGE
  probability 0.9613
  confidence  0.8447
```

Do not force irrelevant values onto every answer type.

### Acceptance criteria

- CHOICE answers preserve the selected choice.
- CHOICE probabilities are preserved.
- confidence is preserved when supplied.
- BOOLEAN represents the underlying decision without exposing `noul`.
- SCORE results can be represented for future use.
- token usage can be preserved in a provider-neutral form.

---

# 7. Task 5 — Implement Native SystemOne Transport

Remove Nimble decision use of:

```text
POST /api/generate
```

Implement:

```text
POST /v1/systemone
```

Transport responsibilities:

```text
HTTP
JSON encoding
JSON decoding
timeouts
HTTP status handling
context cancellation
```

Transport must NOT contain SOP policy.

Expected package:

```text
internal/providers/nimble/
├── config.go
├── nimble.go
├── transport.go
├── systemone.go
└── translate.go
```

Exact filenames may vary if a cleaner organization emerges.

### Acceptance criteria

- `/api/generate` is not used for Nimble decisions.
- no prompt-generation code remains.
- no Markdown fence extraction remains.
- transport is testable using `httptest`.
- normal tests require no Ollama process.

---

# 8. Task 6 — SystemOne Request Translation

Translate provider-neutral questions into SystemOne questions.

Example:

```text
CHOICE
   ↓
type: choice
criteria: {...}
```

BOOLEAN:

```text
BOOLEAN
   ↓
type: noul
```

SCORE:

```text
SCORE
   ↓
type: score
```

The translation boundary should therefore be:

```text
decision.Question
       │
       ▼
Nimble translator
       │
       ▼
SystemOneQuestion
```

### Acceptance criteria

Translation tests exist for:

- CHOICE
- BOOLEAN
- SCORE
- multiple questions
- malformed questions

---

# 9. Task 7 — Normalize SystemOne Responses

Use the verified real Nimble response contract.

Verified response:

```json
{
  "model": "nimble",
  "answers": {
    "risk": {
      "type": "choice",
      "choice": "HIGH",
      "probabilities": {
        "LOW": 0.0013367338788281202,
        "MEDIUM": 0.041031786153483775,
        "HIGH": 0.9576314799676882
      },
      "confidence": 0.8349416441131589
    },
    "approval_required": {
      "type": "noul",
      "noul": 0.9974712183588139
    },
    "execution_model": {
      "type": "choice",
      "choice": "LARGE",
      "probabilities": {
        "SMALL": 0.0016481575652357388,
        "MEDIUM": 0.03704375624257817,
        "LARGE": 0.9613080861921861
      },
      "confidence": 0.8447327143351214
    }
  },
  "usage": {
    "input_tokens": 983,
    "output_tokens": 4
  }
}
```

Normalize this without losing meaningful information.

### Acceptance criteria

Reject:

- missing requested answers
- unknown answer types
- choices outside allowed choices
- invalid probability ranges
- invalid confidence ranges
- malformed JSON
- incompatible response types

---

# 10. Task 8 — BOOLEAN / Noul Boundary

SystemOne exposes:

```text
type: noul
noul: 0.997...
```

The public decision API exposes:

```text
BOOLEAN
```

The translation must happen entirely inside the Nimble provider.

Conceptually:

```text
decision.BOOLEAN
       │
       ▼
SystemOne "noul"
       │
       ▼
Nimble
       │
       ▼
noul probability
       │
       ▼
decision.BooleanAnswer
```

Document the semantic interpretation clearly.

Do not treat the probability as an SOP approval policy.

For example:

```text
0.997
```

is provider output.

Whether `0.997` is sufficient to trigger an approval gate remains the responsibility of `agentic-sop`.

---

# 11. Task 9 — Improve Availability

Current behavior checks only:

```text
Ollama reachable?
```

Phase 1.1 must distinguish:

```text
Ollama unavailable
Nimble model unavailable
Nimble available
```

Use Ollama's model information endpoint or model listing to verify that the configured model exists.

### Acceptance criteria

`Available()` returns false when:

- Ollama cannot be reached
- request times out
- configured Nimble model is absent

Tests cover each case.

Do not make an inference request merely to determine availability.

---

# 12. Task 10 — Normalize Transport Errors

Connection failures currently risk becoming generic provider failures.

Normalize at least:

```text
connection refused
timeout
DNS/network unreachable
context deadline
```

into:

```text
ErrUnavailable
```

Keep malformed responses separate:

```text
ErrMalformedResponse
```

Backend errors that do not represent availability problems can remain:

```text
ErrProviderFailure
```

### Acceptance criteria

`errors.Is()` works correctly for each normalized category.

---

# 13. Task 11 — Golden Fixture

Create:

```text
tests/fixtures/nimble/
└── systemone_response.json
```

Populate it with the verified real Nimble response captured during Phase 1.1.

Use it to test the production response parser.

This becomes the first known-good provider contract fixture.

### Acceptance criteria

The fixture:

- comes from a real Nimble instance
- parses successfully
- produces all three normalized answers
- verifies token usage
- does not require Ollama during tests

---

# 14. Task 12 — Live Integration Test

Add an opt-in integration test.

Default:

```bash
go test ./...
```

must remain completely offline.

Live test:

```bash
NIMBLE_INTEGRATION_TEST=1 go test ./...
```

The integration test should:

1. connect to `OLLAMA_BASE_URL`
2. verify configured Nimble model availability
3. send the SOP example state
4. ask:
   - risk
   - approval_required
   - execution_model
5. validate all requested answers exist
6. validate result types
7. validate probability ranges
8. validate confidence ranges where applicable

Do NOT assert:

```text
risk == HIGH
execution_model == LARGE
```

Model decisions may change.

Test the contract, not the exact judgment.

---

# 15. Task 13 — Update CLI

Keep:

```text
sop-decision-adapter decide
```

but allow a single invocation to exercise a multi-question request.

Prefer input from a JSON request file for complex requests rather than creating dozens of CLI flags.

For example:

```bash
sop-decision-adapter decide \
  -provider nimble \
  -file tests/fixtures/risk-evaluation.json
```

The CLI remains a debugging/testing surface.

It is not the primary integration mechanism for `agentic-sop`.

---

# 16. Task 14 — Documentation

Update:

```text
README.md
docs/PRD.md
docs/PLAN.md
docs/ARCHITECTURE.md
docs/BACKLOG.md
```

Document:

- native `/v1/systemone`
- multi-question requests
- public decision API
- BOOLEAN ↔ `noul`
- golden fixtures
- integration testing
- provider responsibility boundary

Mark the previous live Nimble response verification backlog item complete.

---

# 17. Tests

Required offline test coverage:

```text
decision/
  request validation
  question validation
  result validation
  error classification

nimble/
  CHOICE translation
  BOOLEAN translation
  SCORE translation
  multi-question translation
  response normalization
  malformed response
  unknown answer
  missing answer
  invalid probability
  invalid confidence
  backend unavailable
  model unavailable
  HTTP failure
  timeout

fixtures/
  golden SystemOne response
```

All tests must run without Ollama unless explicitly opted into integration testing.

---

# 18. Validation

Before Phase 1.1 can be marked complete:

```bash
go fmt ./...
go vet ./...
go test ./...
```

must succeed.

Then, with the locally installed Nimble:

```bash
NIMBLE_INTEGRATION_TEST=1 go test ./...
```

must also succeed.

Finally perform one manual CLI request against Nimble.

---

# 19. Definition of Done

Phase 1.1 is complete when:

- [x] `decision` is a public package
- [x] decision API supports multiple questions
- [x] CHOICE is provider-neutral
- [x] BOOLEAN is provider-neutral
- [x] SCORE answers are provider-neutral
- [x] Nimble uses `/v1/systemone`
- [x] `/api/generate` decision path is removed
- [x] prompt-building code is removed
- [x] SystemOne types remain internal
- [x] BOOLEAN ↔ `noul` translation works
- [x] availability verifies Nimble exists
- [x] transport errors are normalized
- [x] golden real-response fixture exists
- [x] offline tests pass
- [x] live Nimble integration test passes
- [x] CLI works against real Nimble
- [x] documentation reflects the actual architecture
- [x] no SOP policy has moved into the adapter

Deferred (see `docs/BACKLOG.md`): a first-class provider-neutral score-scale
field so SCORE candidates no longer have to ride along in `Choices`.

---

# 20. Phase Boundary

Do not begin Julia during Phase 1.1.

The completion boundary is:

```text
                  VERIFIED
                     │
                     ▼
              decision contract
                     │
                     ▼
              Nimble provider
                     │
                     ▼
              /v1/systemone
                     │
                     ▼
               real Ollama
```

Only after this contract is stable should Phase 2 begin:

```text
                 decision.Provider
                    /       \
                   /         \
              Nimble         Julia
                 │             │
            SystemOne        ONNX
```

Nimble becomes the reference implementation against which future adapters are tested.

# Expected Next Phase

**Phase 2 — Julia Provider**

Phase 2 should adapt Julia to the established Phase 1.1 decision contract without changing that contract unless a genuine provider-neutral deficiency is discovered.
