# BACKLOG — sop-decision-adapters

Deferred work. Nothing here is required for the current shipped phase.

## Completed

- [x] **Live Nimble response verification.** A real response was captured from a
      locally installed Ollama + Nimble instance over `POST /v1/systemone` and
      committed as `tests/fixtures/nimble/systemone_response.json`. The production
      parser is tested against it (Phase 1.1).
- [x] **Native SystemOne decision transport.** Replaced the prompt-based
      `/api/generate` path with `POST /v1/systemone`; prompt-building and
      Markdown-fence extraction removed.
- [x] **Public decision contract.** `internal/decision/` promoted to the
      importable `decision/` package
      (`github.com/imhttran/sop-decision-adapters/decision`) with multi-question
      requests and typed CHOICE/BOOLEAN/SCORE answers.
- [x] **Golden fixtures.** `tests/fixtures/nimble/systemone_response.json` plus a
      multi-question CLI example (`tests/fixtures/risk-evaluation.json`).
- [x] **Opt-in live integration test.** `NIMBLE_INTEGRATION_TEST=1 go test ./...`
      validates the contract against a real backend.
- [x] **Julia provider adapter (Phase 2).** `internal/providers/julia` adapts
      `DecisionRequest` to a Julia runtime request and normalizes raw `Outputs`
      back into a `decision.DecisionResult`. Concrete ONNX execution is isolated
      behind the `Runner` seam, with a stdlib-only `CommandRunner` that runs the
      repository-owned helper `tools/julia/infer.py`. Ships with offline tests,
      fixtures under `tests/fixtures/julia/`, and CLI selection
      (`decide -provider julia`). The public contract is unchanged.
- [x] **Julia-1-ONNX execution path (Phase 2.1).** Added the repo-owned Python/ONNX
      helper (`tools/julia/infer.py`) with the documented tensor contract
      (`input_ids`, `attention_mask`, `marker_pos`, `marker_mask`, `qtype` →
      `logits`), stable softmax, logits validation, and qtype-correct decoding
      (choice `argmax`, score expected index, boolean `P(true)`). Julia's native
      qtype mapping and 2–20 option limit are enforced in Go
      (`semantics.go`). Offline Go tests plus standard-library helper tests keep
      the default suite dependency-free.

## Deferred — later phases

- [ ] **Concrete in-process ONNX runtime binding for Julia.** The Julia adapter
      ships with a configurable `Runner` seam and the repo-owned `CommandRunner`
      (running `tools/julia/infer.py`). Replace/augment it with an in-process ONNX
      binding that satisfies the same `Runner` interface (`Name`/`Available`/`Run`)
      so live inference no longer requires a helper process. **Owner:**
      `internal/providers/julia`.
- [ ] **Verify the Julia encoder against upstream.** The tensor contract and
      decoding rules are implemented from the documented model contract, but the
      encoding of state/question/options into `marker_pos`/`marker_mask`
      (`Runtime.encode` in `tools/julia/infer.py`) is a best-effort, isolated
      implementation that has **not** been reconciled with the upstream
      Julia-1-ONNX export. Reconcile it before trusting live output, and capture a
      real golden fixture (`tests/fixtures/julia/`) with recorded provenance.
      **Owner:** `tools/julia`, `internal/providers/julia`.
- [ ] **First-class score-scale representation.** SystemOne `score` questions
      require `criteria` to be an array of 2–26 candidate descriptions; a string or
      a 0/1-element array is rejected with HTTP 400. The provider-neutral
      `decision.Question` has no dedicated score-scale field yet, so the Nimble
      adapter maps score candidates from the existing `Choices` field (and returns
      `decision.ErrInvalidRequest` when fewer than two candidates are supplied).
      Add a first-class score-scale representation to the provider-neutral contract
      so score candidates no longer have to ride along in `Choices`.
- [ ] **Distinct BOOLEAN false/true descriptions.** The provider-neutral BOOLEAN
      question carries no separate descriptions for its false and true options, so
      the Julia adapter always sends `["false", "true"]` as noul options and
      cannot pass caller-supplied false/true criteria. Preserving them requires a
      change to the public contract and is deferred. **Owner:** `decision`,
      `internal/providers/julia`.
- [ ] **CLM, JEV/OpenJEV providers (Phase 3)**.
- [ ] **Provider registry** — dynamic provider discovery/selection.
- [ ] **Shadow evaluation (Phase 4)** — run a second provider for observation
      only; never affect execution.
- [ ] **agentic-sop integration** — consume the public `decision` package from
      `agentic-sop` (kept out of the adapter repository).

## Follow-ups — tooling / environment

- [ ] **Stale `SOP_AGENT_*` environment in the SOP session.** The shell exports
      `SOP_AGENT_PROVIDER=command` and
      `SOP_AGENT_COMMAND=sh scripts/sop-deepseek-agent.sh`, but that harness was
      removed in `agentic-sop` TASK-058 and replaced by the capability-aware
      `sop-ollama-agent` (`agentic-sop/scripts/agents/sop-ollama-agent.sh`,
      installed at `~/.local/share/sop/bin/sop-ollama-agent`). A raw `sop prompt`
      therefore fails with
      `sh: scripts/sop-deepseek-agent.sh: No such file or directory` before any work
      runs.
  - **Remediation:** unset `SOP_AGENT_PROVIDER` / `SOP_AGENT_COMMAND` so SOP uses
    this project's configured native harness (`harness: tool`,
    `provider: ollama` in `.agent-sdlc/config.yaml`), or point
    `SOP_AGENT_COMMAND` at `agentic-sop/scripts/agents/sop-ollama-agent.sh`.
  - **Note:** the installed command-agent binary is pinned to a revision that
    predates `SOP_OLLAMA_IMPLEMENT_ITERATIONS`, so raising the IMPLEMENT budget
    requires the native harness path.
  - **Owner:** operator environment / `agentic-sop`.

## Explicitly out of scope

MCP, database, web UI, and a generic plugin framework are not planned for the
adapter layer. Standard library first.
