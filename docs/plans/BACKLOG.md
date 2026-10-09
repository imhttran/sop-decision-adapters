# BACKLOG — sop-decision-adapters

Deferred work. Nothing here is required for the current shipped phase.

**Two separate phase axes — do not conflate them.** This file numbers the
*provider roadmap* (which adapters ship, in what order). The Julia routing work
numbers its own *experiment sequence*, and those numbers appear in
`tools/julia/*.py` and `docs/reports/julia-*.md`. Julia experiment phases 3 and 4
are complete; provider-roadmap phases 3 and 4 are not. Where this file says
"phase", it always means the provider roadmap.

## Completed

- [x] **Live Nimble response verification.** A real response was captured from a
      locally installed Ollama + Nimble instance over `POST /v1/systemone` and
      committed as `tests/fixtures/nimble/systemone_response.json`. The production
      parser is tested against it (provider-roadmap phase 1.1).
- [x] **Native SystemOne decision transport.** Replaced the prompt-based
      `/api/generate` path with `POST /v1/systemone`; prompt-building and
      Markdown-fence extraction removed.
- [x] **Public decision contract.** `internal/decision/` promoted to the
      importable `decision/` package
      (`github.com/imhttran/sop-decision-adapters/decision`) with multi-question
      requests and typed CHOICE/BOOLEAN/SCORE answers.
- [x] **Golden fixtures.** `tests/fixtures/nimble/systemone_response.json` plus a
      multi-question CLI example (`tests/fixtures/risk-evaluation.json`).
- [x] **Opt-in live integration tests.** Four independent opt-in gates validate
      the contract against a real backend; all are off by default and the
      default suite stays offline:
      - `NIMBLE_INTEGRATION_TEST=1` — Nimble over `POST /v1/systemone`
      - `JULIA_INTEGRATION_TEST=1` — Julia via the `CommandRunner` helper
      - `JULIA_HTTP_INTEGRATION_TEST=1` — Julia over the `http_bridge.py` service
      - `CLEF_LIVE_TEST=1` — Clef over `POST /v1/systemone` (also requires
        `CLEF_ENABLED=1`; skips when the backend or model is absent)

      None of these appears in `.env.example`, which documents provider
      configuration only.
- [x] **Julia provider adapter (provider-roadmap phase 2).** `internal/providers/julia` adapts
      `DecisionRequest` to a Julia runtime request and normalizes raw `Outputs`
      back into a `decision.DecisionResult`. Concrete ONNX execution is isolated
      behind the `Runner` seam, with a stdlib-only `CommandRunner` that runs the
      repository-owned helper `tools/julia/infer.py`. Ships with offline tests,
      fixtures under `tests/fixtures/julia/`, and CLI selection
      (`decide -provider julia`). The public contract is unchanged.
- [x] **Julia-1-ONNX execution path (provider-roadmap phase 2.1).** Added the repo-owned Python/ONNX
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
      `decision.Question` has no dedicated score-scale field yet, so score
      candidates ride along in the existing `Choices` field. **All three adapters
      do this now**, each with its own limit, so the fix is no longer
      single-adapter: Nimble and Clef enforce SystemOne's 2–26 range in
      `translate.go`, and Julia enforces its own 2–20 option limit in
      `semantics.go`. Each returns `decision.ErrInvalidRequest` when too few
      candidates are supplied. Add a first-class score-scale representation to the
      provider-neutral contract so score candidates no longer have to ride along
      in `Choices`. **Owner:** `decision`, `internal/providers/{nimble,clef,julia}`.
- [ ] **Distinct BOOLEAN false/true descriptions.** The provider-neutral BOOLEAN
      question carries no separate descriptions for its false and true options, so
      the Julia adapter always sends `["false", "true"]` as noul options and
      cannot pass caller-supplied false/true criteria. Preserving them requires a
      change to the public contract and is deferred. **Owner:** `decision`,
      `internal/providers/julia`.
- [ ] **De-duplicate the Clef/Nimble SystemOne transport and wire types.**
      `internal/providers/clef/{transport,systemone}.go` are near-byte-for-byte
      copies of `internal/providers/nimble/{transport,systemone}.go` (same HTTP
      client, error classification, and `/v1/systemone` wire shapes under
      different casing). `translate.go` differs legitimately (Clef wraps
      `ErrUnsupported`) and is out of scope for this item. Extract the shared
      transport + wire types into one internal package (e.g.
      `internal/providers/systemone`) that both adapters wrap. Deferred because
      both packages' tests reach the unexported types directly (white-box,
      same-package), so the extraction touches most of
      `internal/providers/{clef,nimble}/*_test.go` (~3000 lines) for a
      maintenance-only (not correctness) gain — the concrete risk this
      duplication caused (a validation fix landing in one copy and not the
      other; see `decision.Answer.Validate`'s NaN/Inf guard) is already closed
      at the shared `decision` package choke point, so further drift of this
      kind is no longer possible even with the duplication left in place.
      **Owner:** `internal/providers/clef`, `internal/providers/nimble`.
- [ ] **CLM, JEV/OpenJEV providers (provider-roadmap phase 3)**.
- [ ] **Shadow evaluation (provider-roadmap phase 4)** — run a second provider for observation
      only; never affect execution. The adapter-side harness already exists
      (`internal/shadow`, delivered by CLEF-007) and is deliberately unwired: no
      production package imports it. What remains is the wiring, not the
      harness. Do not mistake its empty import graph for dead code.
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

## Removed

Work that shipped and was then deleted. Recorded here so a later reader does not
mistake a deleted package for one that never existed, or re-derive it from a
report that still describes it as present.

- **`internal/benchmark`** — the CLEF-008 benchmark harness (`benchmark.go`,
  `cells.go`, `corpus.go`, `corpus.json`, plus tests; ~1,100 lines). Deleted as
  dead code: no package outside it ever imported it, it had no CLI subcommand or
  make target to run it, `DefaultCells()` and `EmbeddedCorpus()` had no callers
  at all, and its own test asserted its disuse
  (`TestBenchmarkNotImportedByProductionPath`). It also reimplemented
  `internal/shadow`'s metric layer rather than reusing it. The measurements it
  produced survive in `docs/reports/clef-provider/CLEF-008-benchmark.md`, which
  still cites the deleted paths as present deliverables and needs correcting;
  `CLEF-010-readiness-decision.md` cites them as readiness evidence and needs
  re-verifying. Rebuild from `internal/shadow` if a benchmark is wanted again.

## Explicitly out of scope

MCP, database, web UI, and a generic plugin framework are not planned for the
adapter layer. Standard library first.

- **Provider registry** — dynamic provider discovery/selection. The `newProvider`
  switch in `cmd/sop-decision-adapter/main.go` is the documented seam for adding
  an adapter and is three cases long; nothing is asking for runtime discovery.
  Moved here from Deferred: speculative flexibility, not deferred work. Revisit
  only when a caller genuinely cannot name its provider at compile time.
