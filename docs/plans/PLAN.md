# PLAN — sop-decision-adapters

A phased plan. Each phase is additive and keeps the `Provider` contract stable.
Nothing in a later phase may leak its vendor types into the public `decision`
package.

Guiding constraints for every phase:

- Providers evaluate; they do not define or enforce SOP policy.
- Prefer the standard library; avoid a plugin framework and unnecessary
  abstractions.
- Every provider must be testable without its backend running.

---

## Phase 0 — Foundation

**Goal:** one stable decision API, a minimal CLI, fixtures, and docs.

Deliverables:

- `decision` (later promoted to a public package):
  - `Provider` interface (`Name`, `Available`, `Decide`).
  - `DecisionRequest` + `Validate`.
  - `DecisionResult` + `Validate`.
  - Normalized `ProviderError` with vendor-neutral `ErrorKind`.
- `cmd/sop-decision-adapter`: a minimal CLI with a `decide` subcommand.
- `tests/fixtures`: `risk`, `approval_required`, `execution_model` shapes.
- Docs, `.env.example`, `.gitignore`, `go.mod`.
- Unit tests: request/result validation, error model, fixture validity.

Exit criteria: `go fmt ./...`, `go vet ./...`, `go test ./...` all pass with no
backend running.

**State: COMPLETE.**

---

## Phase 1 — Nimble

**Goal:** evaluate decisions with the Nimble model served by an Ollama-compatible
backend and return a normalized `DecisionResult`.

Final path (Phase 1.1):

```
decision/ ──▶ internal/providers/nimble ──▶ POST /v1/systemone ──▶ Ollama ──▶ Nimble
```

### Phase 1.1 — Native Nimble / SystemOne hardening

**State: COMPLETE.**

What shipped:

- **Public contract.** `internal/decision` promoted to the importable
  `decision` package (`github.com/imhttran/sop-decision-adapters/decision`) with
  `provider.go`, `request.go`, `result.go`, `errors.go`. No provider-specific
  types appear in it.
- **Typed questions.** `CHOICE`, `BOOLEAN`, `SCORE` with per-type required
  fields, unknown-type and duplicate-ID rejection. The public API contains no
  `noul`.
- **Multi-question requests / multi-answer results.** One `Decide` invocation
  evaluates several questions against one state; `DecisionResult.Answers`
  preserves each answer's semantic type and provider-neutral `Usage`.
- **Native SystemOne transport.** `internal/providers/nimble` uses
  `POST /v1/systemone`; the prompt-based `/api/generate` decision path,
  prompt-building, and Markdown-fence extraction are removed. Transport handles
  HTTP, JSON encode/decode, timeouts, status handling, and context cancellation
  only.
- **Translation.** `decision.Question` → `SystemOneQuestion` (CHOICE→`choice`,
  BOOLEAN→`noul`, SCORE→`score`), and normalized responses back to
  `decision.Answer`.
- **BOOLEAN ↔ `noul` boundary.** The `noul` representation exists only inside
  `internal/providers/nimble`; its probability is provider output, not SOP
  policy.
- **Availability.** `Available` distinguishes Ollama-unreachable,
  request-timeout, and model-absent using the model listing endpoint; no
  inference request is made to check availability.
- **Error normalization.** connection refused / timeout / DNS / context deadline
  → `ErrUnavailable`; undecodable body → `ErrMalformedResponse`; other backend
  errors → `ErrProviderFailure`, all matched with `errors.Is`.
- **Golden fixture.** `tests/fixtures/nimble/systemone_response.json` is a real
  captured response driven through the production parser offline.
- **Live integration test.** `NIMBLE_INTEGRATION_TEST=1 go test ./...` validates
  the contract against a real backend; the default run stays offline.
- **CLI.** `decide -provider nimble -file tests/fixtures/risk-evaluation.json`
  exercises a multi-question request.

Exit criteria (met): `go fmt ./...`, `go vet ./...`, `go test ./...` pass fully
offline; `NIMBLE_INTEGRATION_TEST=1 go test ./...` passes against a locally
installed Nimble; a manual CLI request returns all requested answers.

---

## Phase 2 — Julia

**State: PLANNED.**

**Goal:** adapt Julia (ONNX) to the Phase 1.1 decision contract **without
changing that contract** unless a genuine provider-neutral deficiency is found.

Deliverables:

- `internal/providers/julia` translating `DecisionRequest` into the ONNX model's
  input tensor(s) and back into a `DecisionResult`.
- Model loading/config (model path, input/output names, label mapping) via
  environment.
- Testability: an in-process fake ONNX/runner interface so tests do not require
  the ONNX runtime.

Exit criteria: the same fixtures produce valid `DecisionResult`s; no Julia/ONNX
and no SystemOne-specific types leak into the public `decision` package. Nimble
is the reference implementation against which Julia is tested.

---

## Phase 3 — CLM + JEV

**State: PLANNED.**

**Goal:** add the CLM and JEV/OpenJEV providers behind the same contract.

Deliverables:

- `internal/providers/clm` and `internal/providers/jev`.
- Shared translation helpers where genuinely common — but no premature shared
  abstraction beyond the contract.
- Provider selection in the CLI for all implemented providers.

Exit criteria: multiple providers satisfy identical request/result tests;
`agentic-sop` can switch providers without changing its policy code.

---

## Phase 4 — Shadow evaluation

**State: PLANNED.**

**Goal:** run a shadow provider alongside the primary one and report
disagreements — **without ever affecting execution**.

```
DecisionRequest
      │
      ├──────────────┐
      ▼              ▼
   primary        shadow
      │              │
      └──── comparison ┘
             │
             ▼
     disagreement report
```

Deliverables:

- A shadow wrapper that calls a **primary** provider (whose result is used) and
  a **shadow** provider (whose result is recorded only), and emits a
  disagreement report (choice mismatch, confidence delta, missing choices).
- Reports are observational artifacts (logs/metrics), never an execution input.

Hard constraint: **shadow providers MUST NOT affect execution.** The primary
result is returned regardless of the shadow result; shadow errors are captured,
never propagated as failures of the primary decision.

---

## Deferred / explicitly out of scope for now

Tracked, actionable deferred items live in [`BACKLOG.md`](BACKLOG.md). In
summary:

- Julia, CLM, and JEV implementations (Phases 2–3).
- Databases, web UI, MCP, agent orchestration.
- Any `agentic-sop` dependency.
- A generic plugin/discovery framework.
- A provider registry/selector for future providers.
