# CLEF — Clef Provider Integration Architecture Analysis

**Type:** Analysis (non-normative; plan input)
**Status:** DRAFT — analysis complete at the stated revision; no implementation.
**Scope:** `sop-decision-adapters` only. This report makes **no production change**
and does not modify `agentic-sop`.
**Authority:** `docs/plans/PLAN-Clef-Provider-Integration.md` (to be created by this
effort); the cross-module contract is `agentic-sop`
`docs/reports/decision-integration/SEAM-002-decision-contract.md`.

This report is the design input for the CLEF plan. It records the adapter's
current architecture, the proven `agentic-sop` handoff, Clef's *actually
observed* capabilities, the `oMLX` compatibility finding, the transport
alternatives, the selected transport, and the request/result and failure
mappings. Where a fact is not established from evidence it is marked
**UNKNOWN / REQUIRES EXPERIMENT** and is *not* assumed.

---

## 0. Evidence baseline

| Fact | Value | Source |
|---|---|---|
| `sop-decision-adapters` branch | `main` | `git status`/`git log` |
| `sop-decision-adapters` HEAD | `a419ce59b86e6e90709556b187deb7fe694eeb30` ("Phase 2.1: concrete Julia-1-ONNX execution path") | `git log -1` |
| Working tree | clean (no uncommitted changes) | `git status --short` |
| Go toolchain | `go1.27.1 darwin/arm64` | `go version` |
| Offline gates | `go build ./...`, `go vet ./...`, `go test ./...` all pass | run at HEAD |
| `agentic-sop` HEAD | `11539138225a84a3bd476600cb1d1d8acabffc21` ("feat(decision): provider-neutral decision integration seam (SEAM-001..008)") | `git log -1` |
| `agentic-sop` seam verdict | **READY_WITH_NOTES** (0 CRITICAL, 0 HIGH, 0 MEDIUM, 0 blocking residuals; L1–L6 non-blocking notes) | `SEAM-008-readiness-decision.md` |
| Local Ollama | `0.40.0` reachable at `http://localhost:11434` | `ollama --version`, `/api/version` |
| `omlx` | `0.7.0`, installed at `/opt/homebrew/bin/omlx` | `omlx --version` |
| Clef (Ollama) | `clef-flash:latest`, architecture `clef`, 9.5B, `mxfp8`, capabilities `decision`, `vision` | `ollama show clef-flash` |
| Clef (HF) | `mlx-community/clef-4bit`, `mlx-community/clef-8bit` referenced; weights **not** materialized locally | HF cache `refs/main` only |

---

## 1. Current adapter architecture (`sop-decision-adapters`)

### 1.1 Current-state diagram

```text
agentic-sop (separate repo, separate Go module)
     |   policy · approval · lifecycle · authorization · human approval
     |   (does NOT import this module)
     v
 decision/                                   <-- PUBLIC, provider-neutral Go contract
     |   Provider{Name, Available, Decide}
     |   DecisionRequest{State, Questions[]} ──▶ DecisionResult{Provider, Model, Answers, Usage}
     |   Question{ID, Type∈{choice,boolean,score}, Instructions, Criteria, Choices}
     |   Answer{Type, Choice, Probabilities, Confidence, Probability, Score}
     |   ProviderError{Provider, Kind∈{invalid_request,unavailable,malformed_response,provider_failure}}
     +---------------------+----------------------+
     |                                            |
     v                                            v
 internal/providers/nimble                 internal/providers/julia
     |  Transport iface (SystemOne,ListModels)   |  Runner iface (Name,Available,Run)
     |  systemone.go (unexported wire)           |  runner.go (Inputs/Outputs, CommandRunner)
     |  translate.go / config.go                 |  translate.go / semantics.go / config.go
     v                                            v
 POST {OLLAMA_BASE_URL}/v1/systemone        tools/julia/infer.py ──▶ ONNX Runtime
     v                                            v
 Ollama ──▶ Nimble (decision)               Julia-1-ONNX (supplied-options)

 cmd/sop-decision-adapter  decide -provider {nimble|julia} [-file|-state]
                           (debugging surface; local multi-question contract)

 tests/  offline suite + opt-in live tests (NIMBLE_INTEGRATION_TEST, JULIA_INTEGRATION_TEST)
```

### 1.2 Inventory (from source, not from old docs)

- **Provider interface** — `decision/provider.go`: `Name() string`,
  `Available(ctx) bool` (cheap, no inference), `Decide(ctx, DecisionRequest)
  (DecisionResult, error)`. Multi-question: one call evaluates several questions
  against one `State`.
- **Request/result types** — `decision/request.go` (`QuestionType` = `choice`,
  `boolean`, `score`; `Question`; `DecisionRequest`) and `decision/result.go`
  (`AnswerType`; `Answer` with `Choice`, `Probabilities`, `Confidence`,
  `Probability`, `Score`; `Usage`; `DecisionResult`). All provider-neutral; no
  vendor vocabulary.
- **Error model** — `decision/errors.go`: `ProviderError{Provider, Kind, Err}`,
  `ErrorKind` = `invalid_request | unavailable | malformed_response |
  provider_failure`, with sentinels `ErrInvalidRequest`, `ErrUnavailable`,
  `ErrMalformedResponse`, `ErrProviderFailure`. There is **no `status` enum** and
  **no `contract_version`** in this local contract.
- **Existing adapters** — `internal/providers/nimble` (HTTP `POST
  /v1/systemone`; `Transport` interface + unexported SystemOne wire types;
  deterministic translation; availability via `GET /api/tags`; error
  classification) and `internal/providers/julia` (ONNX behind a `Runner`
  interface; `CommandRunner` runs `tools/julia/infer.py` JSON-in/JSON-out;
  qtype/option semantics; error classification). Both implement
  `decision.Provider`.
- **Registration/factory** — there is **no registry**. The CLI's
  `newProvider(name, …)` switch in `cmd/sop-decision-adapter/main.go` is the
  single selection seam (`""`/`"nimble"`, `"julia"`); unknown names error.
- **Configuration** — per-provider `…ConfigFromEnv()`: `OLLAMA_BASE_URL`,
  `NIMBLE_MODEL`, `NIMBLE_TIMEOUT`; `JULIA_*`. Precedence: CLI flags → env →
  defaults.
- **Capability representation** — none beyond `QuestionType`
  (`choice`/`boolean`/`score`) and per-adapter native capability rules
  (e.g. Julia's 2–20 option limit). No explicit capability advertisement/handshake.
- **Timeout/cancellation** — `context.Context` on `Available`/`Decide`; each
  transport applies `cfg.Timeout`; cancellation surfaces as `ErrUnavailable`
  (Nimble) / `errRunnerUnavailable` (Julia).
- **Tests** — `decision/*_test.go` (validation, error model);
  `internal/providers/nimble/{translate,transport,nimble,golden,config}_test.go`
  (fake transport + `net/http/httptest` + real golden fixture);
  `internal/providers/julia/{translate,runner,julia,config}_test.go`
  (`RunnerFunc` fakes + hermetic `sh -c`); `tests/fixtures_test.go` +
  `tests/integration_{nimble,julia}_test.go` (opt-in, env-gated, offline-safe).
  The default suite requires no Ollama/ONNX/Julia/network.
- **Transports** — HTTP (`net/http`) for Nimble; process/stdin-stdout
  (`os/exec`) for Julia's `CommandRunner`.
- **Relationship to `agentic-sop`** — *intended* one-way dependency: "agentic-sop
  depends on the public `decision` package". **In practice** (verified in
  `agentic-sop`) the live seam does **not** import this module: it reaches an
  external adapter over a **provider-neutral JSON process boundary**
  (`internal/decision/command`). See §2.

### 1.3 The contract divergence (finding, not a blocker)

There are **two different provider-neutral contracts** in play, and they are not
the same shape:

| Aspect | Local `decision.Provider` (this repo) | SOP-proven SEAM-002 DTO (`agentic-sop`) |
|---|---|---|
| Granularity | one request, **many** questions | one request, **one** bounded decision |
| Request | `State`, `Questions[]` | `kind`, `question`, `signals`, `choices`, `deadline_ms` |
| Result | `Answers map[id]Answer` + `Usage` | `status ∈ {OK,UNSUPPORTED,ERROR,INDETERMINATE}`, `choice`, `confidence`, `diagnostics`, `error` |
| Version | none | `contract_version` (required, `=1`) |
| Failure | `ErrorKind` (4 kinds) | status enum **+** typed transport error |

**Consequence for Clef:** the adapter must bridge these. It reuses the local
`decision.Provider` contract to *implement* a provider (consistent with
Nimble/Julia), and adds a small, **provider-neutral adapter-side wire layer** that
speaks the SEAM-002 DTO to `agentic-sop`. Neither contract changes. The local
result cannot express `status`, so the wire layer derives status from the local
result plus adapter-side capability detection — it does not require a change to
`decision/`.

**This is not a STOP condition.** The generic contract that `agentic-sop` actually
exposes (SEAM-002) fully represents Clef's outcomes (see §8). The local Go
interface is an internal implementation detail, not the cross-module contract.

---

## 2. Proven `agentic-sop` handoff

The seam is committed at `agentic-sop` HEAD
`11539138225a84a3bd476600cb1d1d8acabffc21` and was judged **READY_WITH_NOTES**
(0 CRITICAL / 0 HIGH / 0 MEDIUM / 0 blocking). L1–L6 are non-blocking operational
and documentation notes and are **not** justification to redesign the seam.

### 2.1 How SOP reaches an external adapter (verified)

`internal/decision/command.Adapter` (SOP-owned):

- is constructed with a stable, provider-neutral `name` and an executable `argv`;
- runs the process with `exec.CommandContext` (separate executable + argv, **no
  shell**, `WaitDelay` drain bound, 1 MiB stdout cap);
- writes the SEAM-002 **request DTO** to the process **stdin**;
- reads the SEAM-002 **result DTO** from the process **stdout**;
- validates with `req.ValidateResult(res)` (**SOP-owned, fail-closed**) before any
  policy can consume it;
- treats a non-zero exit, empty stdout, oversized stdout, malformed JSON, or a
  validation failure as a **typed failure** — never a success.

SOP selection wiring (`internal/cli/decision_provider.go`): capability is **OFF by
default**; when enabled, provider name `"command"` → `command.New("command",
cfg.Decision.Command)`; `""`/`"deterministic"` → in-process deterministic; any
other name fails closed. There is **no provider-name branch in governance code**.

### 2.2 The SEAM-002 DTO (authoritative)

Request (`SOP → provider`):

| Field | Type | Req | Rule |
|---|---|---|---|
| `contract_version` | int | yes | unrecognized → failure (fail closed) |
| `kind` | string | yes | opaque, provider-neutral capability label |
| `question` | string | yes | bounded free-text subject (signal, not sole safety basis) |
| `signals` | map<string,number> | no | absent key = **unknown**, never zero-confidence approval |
| `choices` | array<string> | no | when present, a returned `choice` MUST be a member |
| `request_id` | string | no | tracing only; carries no authority |
| `deadline_ms` | int | no | advisory; SOP owns enforcement via `ctx` |

Result (`provider → SOP`):

| Field | Type | Req | Rule |
|---|---|---|---|
| `contract_version` | int | yes | must match request |
| `status` | string | yes | one of `OK,UNSUPPORTED,ERROR,INDETERMINATE` |
| `kind` | string | no | echo of served kind (validated against request) |
| `choice` | string | cond. | required when `status=OK`; domain choice (`LOW`/`MEDIUM`/`HIGH`/`HUMAN`) |
| `confidence` | number | no | probability in `[0,1]`; **absent ⇒ indeterminate** |
| `diagnostics` | map<string,string> | no | advisory only; SOP policy MUST NOT read it |
| `error` | string | cond. | required when `status ∈ {UNSUPPORTED, ERROR}` |

### 2.3 Ownership split (governing for this plan)

| Concern | Owner |
|---|---|
| Provider-neutral boundary, validation, evidence interpretation, governance, policy, lifecycle, authorization, human approval | **`agentic-sop`** |
| Provider implementations, Clef implementation, oMLX/provider transport, endpoint handling, request/response translation, provider capability handling, provider-specific tests | **`sop-decision-adapters`** |

Governing invariant: **Providers evaluate. SOP governs.** Provider output is
**evidence**, never authorization. SOP's `ValidateResult` treats an unknown
version/status/choice, an out-of-range/NaN/Inf confidence, a missing confidence,
an unsupported capability, a provider-reported error, and an indeterminate result
as non-approvals (fail closed); a validated result is interpreted only through
`autonomy.Decide`.

### 2.4 What the adapter MUST NOT carry

Per SEAM-002 §5/§15, the result must not express `CONTINUE`, `BLOCK`, `APPROVE`,
`REJECT`, `COMMIT`, `MERGE`, a lifecycle transition, approval, commit/merge
authority, policy-outcome authority, or an execution-model/class selection. A
provider may emit a **domain choice** (e.g. `HUMAN` meaning "this looks like it
needs a human"); SOP maps validated evidence to its own action vocabulary itself.

---

## 3. Clef capability analysis

Two Clef runtimes were investigated separately: the **Ollama-served
`clef-flash`** model (which the live probe below reached) and the **oMLX/MLX
`mlx-community/clef-4bit`** target named by the task (§4).

### 3.1 Verified live probe — `clef-flash` over Ollama `/v1/systemone`

`ollama show clef-flash` → architecture `clef`, 9.5B, context 262144, embedding
4096, quantization `mxfp8`, capabilities **`decision`**, `vision`.

`POST http://localhost:11434/v1/systemone` with `model: "clef-flash"` returned:

```json
{"model":"clef-flash",
 "answers":{"risk":{"type":"choice","choice":"HIGH",
   "probabilities":{"LOW":0.0298,"MEDIUM":0.0641,"HIGH":0.9061},
   "confidence":0.6630}},
 "usage":{"input_tokens":145,"output_tokens":0}}
```

A second probe with a `score` and a `noul` question returned:

```json
{"model":"clef-flash",
 "answers":{
   "confidence_score":{"type":"score","score":4.6530,
     "legend":{"0":"0", …,"10":"10"},
     "probabilities":{"0":0.0339,…,"10":0.0378},
     "confidence":0.0780},
   "approval_required":{"type":"noul","noul":0.8176}},
 "usage":{"input_tokens":377,"output_tokens":0}}
```

**Conclusion:** Clef (as served by Ollama) exposes the *same* decision API and
semantics as the reference Nimble model, with **choice + per-choice probabilities
+ confidence**, **score**, and **boolean/noul** — a near-perfect fit to the
existing provider-neutral *evidence* contract.

### 3.2 Capability classification

| Capability | Classification | Evidence |
|---|---|---|
| Choice (single selected option) | **SUPPORTED** | live probe: `type:"choice"`, `choice:"HIGH"` |
| Per-choice probabilities | **SUPPORTED** | live probe: `probabilities:{LOW,MEDIUM,HIGH}` |
| Confidence in `[0,1]` | **SUPPORTED** | live probe: `confidence:0.6630`, `0.0780` |
| Score | **SUPPORTED** (native) | live probe: `type:"score"`, weighted `score`, `legend` |
| Boolean / noul | **SUPPORTED** (native) | live probe: `type:"noul"`, `noul:0.8176` |
| Token usage | **SUPPORTED** | `usage.input_tokens`/`output_tokens` |
| Request format | **SUPPORTED** | SystemOne request shape accepted for choice/score/noul |
| Response format | **SUPPORTED** | SystemOne response shape (mirrors Nimble) |
| Null / indeterminate handling | **UNKNOWN / REQUIRES EXPERIMENT** | not exercised; behavior on ties/refusal/empty not observed |
| Confidence calibration | **UNKNOWN / REQUIRES EXPERIMENT** | no ground truth yet (CLEF-008) |
| Decision semantics via `oMLX` (4-bit MLX) | **UNSUPPORTED** as a native endpoint; **UNKNOWN** as a translation | §4 |
| Determinism/repeatability | **UNKNOWN / REQUIRES EXPERIMENT** | not measured (CLEF-006/008) |

### 3.3 The oMLX-served `mlx-community/clef-4bit` target

- `omlx` describes itself as a **"Production-ready LLM server for Apple Silicon"**
  and `omlx serve` as a **"multi-model OpenAI-compatible server"**.
- The HF cache contains only `refs/main` for `mlx-community/clef-4bit` (and
  `-8bit`); the weights are **not materialized** locally, so the 4-bit target was
  not actually run.
- It is **not proven** that `mlx-community/clef-4bit` and `clef-flash` are the
  same base model at different quantizations; they are both "clef", but the
  equivalence is **UNKNOWN**.

---

## 4. oMLX compatibility analysis

### 4.1 Observed interface (static, from the installed 0.7.0 app bundle)

`omlx serve` exposes an OpenAI-compatible surface. Discovered routes include:

```text
/v1/chat/completions   /v1/completions   /v1/embeddings   /v1/rerank
/v1/responses          /v1/messages      /v1/models       /v1/models/status
/v1/audio/*            /v1/mcp           /v1/web
(+ a large /api/* control plane)
```

- **No `/v1/systemone`** route exists (the only `systemone` string match in the
  bundle is inside an unrelated bundled binary). oMLX therefore does **not**
  expose Ollama's decision-specific SystemOne API.
- oMLX **does** appear to support `logprobs` / `top_logprobs` (chat/completions),
  which is the only plausible route to recover bounded-choice probabilities.

### 4.2 Conclusion

**oMLX provides generic OpenAI-compatible generation/chat semantics, not Clef's
decision-specific semantics.** Per the task's instruction, the two consequences
are stated explicitly:

1. **Semantic fidelity is not automatic.** OpenAI chat-completion output is
   *not* equivalent to Clef's decision semantics. Treating it as equivalent would
   be an invented compatibility and is rejected.
2. **A translation is *plausibly* possible but unproven.** A safe adapter could,
   in principle, constrain generation to the allowed choices and read
   `top_logprobs` to derive `choice` + probabilities + confidence, then enforce
   the same `[0,1]` bounds and allowed-choice membership. Whether this yields
   *faithful* decision semantics for Clef is **UNKNOWN / REQUIRES EXPERIMENT**
   and must be demonstrated before it is selected.

Therefore: **`mlx-community/clef-4bit` + oMLX is an evaluation target, not a
verified decision transport.** The verified decision transport is Ollama's
SystemOne endpoint serving Clef (§3.1, §6).

---

## 5. Transport alternatives

Evaluated against the SEAM-002 evidence contract. "Verified" means demonstrated
live; "UNKNOWN" means not demonstrated and therefore not assumed.

| Option | Description | Semantic fidelity | Contract fit | Apple Silicon | Timeout/cancel | Error handling | Deterministic parsing | Confidence fidelity | Ops complexity | Testability | Security | Performance |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| **A** | oMLX native decision endpoint | n/a — **no such endpoint** (oMLX is OpenAI-only) | **none** | yes | yes | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| **B** | oMLX OpenAI-compat + explicit, tested Clef translation (grammar/logprobs) | **UNKNOWN** (must be proven) | plausible | yes | yes (`ctx`) | adapter-owned | **risky** (sampling, tokenization) | **UNKNOWN** | medium | medium | no shell; adapter-owned | unknown |
| **C** | Ollama `/v1/systemone` (Clef decision API) | **VERIFIED — native** (choice+prob+confidence) | **excellent** | yes | yes (`ctx`, `http.Client`) | mirrors proven Nimble mapping | **yes** (structured JSON) | **verified present** | **low** | **high** | no shell | local, fast |
| **D** | Direct executable/process adapter (e.g. `mlx_lm` CLI) | UNKNOWN | plausible | yes | yes (`os/exec`+`ctx`) | adapter-owned | fragile (stdout scraping) | UNKNOWN | high | medium | exec (no shell) | local |
| **E** | Another Clef-native interface already in the repo | n/a — none exists | none | — | — | — | — | — | — | — | — | — |

Notable evidence for the two serious options:

- **Option C** reuses the *exact* transport and normalization pattern already
  proven by `internal/providers/nimble` against `POST /v1/systemone`, and the live
  probe returned the identical response shape (`answers[id]{type,choice,
  probabilities,confidence,noul,score}`, `usage`). Semantic loss is **zero** for
  the SEAM-002 choice/confidence contract.
- **Option B** would require an invented, tested request/response translation
  whose fidelity is unproven, with deterministic parsing risk (sampling,
  tokenization of choice labels) — a real risk of silently inventing compatibility.

No option is selected for convenience; C is selected because it is the only one
whose semantic fidelity to Clef's decision behavior is **demonstrated**.

---

## 6. Selected transport

**Primary: Option C — Ollama's native `POST /v1/systemone` decision API serving
Clef (verified with `clef-flash`).**

Rationale (evidence-first):

1. It is the **only** transport that exposes Clef's *decision* semantics
   (choice + per-choice probabilities + confidence) directly — verified live —
   so it fits the provider-neutral evidence contract with **zero semantic loss**.
2. It reuses a **proven** transport/normalization pattern (`nimble` against the
   same endpoint) rather than inventing one.
3. Parsing is deterministic (structured JSON), and confidence/probability
   fidelity is verified present.
4. It keeps all `oMLX`/`Ollama`/`SystemOne` details adapter-side, as required.

**Secondary (experiment, not default): Option B — oMLX OpenAI-compatible endpoint
with an explicitly defined, tested translation.** CLEF-002 determines whether a
fidelity-preserving translation (constrained generation + `top_logprobs`) is
achievable. It is added, if at all, as an *alternate transport* behind the same
provider seam; it never changes the contract or policy. If CLEF-002 shows the
translation cannot be made faithful, oMLX is recorded as not viable for decision
evaluation and is not implemented.

**Design shape:**

```text
decision.Provider (local, unchanged)
        ▲
        │  internal/providers/clef
        │    clef.go       Provider{cfg, transport}
        │    translate.go  DecisionRequest ⇄ Clef/SystemOne (choice-oriented)
        │    transport.go  Transport interface + SystemOneHTTPTransport (private wire)
        │    config.go     CLEF_* env
        │    capability.go supported kinds/operations + UNSUPPORTED/INDETERMINATE rules
        │
        │  internal/evidence  (provider-neutral SEAM-002 wire layer)
        │    request  → decision.DecisionRequest   (single bounded choice)
        │    decision.DecisionResult + error → SEAM-002 result DTO (status/choice/confidence)
        │
        └─ cmd/sop-decision-adapter serve -provider clef   (stdin request DTO → stdout result DTO)
```

The `serve` subcommand is **provider-neutral** (it accepts any local provider),
which is how architecture guard 11 ("another provider can implement the same
generic contract without SOP policy changes") is demonstrated: `agentic-sop`
selects `command` + argv `[…, serve, -provider, clef]`; swapping to `nimble` is a
configuration change, not a policy change.

---

## 7. Request / result mapping (field-by-field ownership)

### 7.1 Request path

`SOP request (SEAM-002 DTO) → adapter request (local) → Clef operation → Clef
result → provider-neutral evidence (SEAM-002 DTO) → SOP validation`

| SEAM-002 request field | Adapter mapping (owner: adapters) | Clef/SystemOne |
|---|---|---|
| `contract_version` | validated `== 1` adapter-side; not forwarded | — |
| `kind` | becomes the single question `ID`; echoed back in the result `kind` | carried in `instructions`/diagnostics |
| `question` | becomes `State` **and** the question `Criteria` | `state`, `criteria` |
| `signals` | deterministically rendered (sorted keys) and appended to `State`; also copied to `Metadata` (audit only) | `state` text |
| `choices` | becomes the question `Choices`; absent ⇒ contract default vocabulary `LOW/MEDIUM/HIGH/HUMAN` | `criteria` options map + `choices` |
| `request_id` | `Metadata["request_id"]` (audit only; carries no authority) | — |
| `deadline_ms` | advisory; the caller's `ctx` is the upper bound | — |

### 7.2 Result path

| Clef/SystemOne answer | Local `decision` | SEAM-002 result (owner: adapters) |
|---|---|---|
| `type:"choice"` + `choice` + `probabilities` + `confidence` | `AnswerChoice{Choice, Probabilities, Confidence}` | `status=OK`, `kind` echoed, `choice`, `confidence`; `diagnostics` = advisory reason/model |
| `type:"choice"` with **no** `confidence` | `AnswerChoice{Confidence:nil}` | `status=OK`, `choice`, **`confidence` omitted** ⇒ SOP treats as *indeterminate* |
| choice **outside** allowed set | reject → `ErrMalformedResponse` | **process exits non-zero** (fail closed; never an invalid choice) |
| confidence out of `[0,1]` / `NaN` / `Inf` | reject → `ErrMalformedResponse` | **process exits non-zero** |
| Clef cannot decide | adapter rule | `status=INDETERMINATE` (+ `error`), exit 0 |
| requested `kind`/operation not served | adapter capability check | `status=UNSUPPORTED` (+ `error`), exit 0 |
| Clef reports a structured error it observed | `ProviderError` | `status=ERROR` (+ `error`), exit 0 |

**Forbidden content (never emitted):** `CONTINUE`, `BLOCK`, `APPROVE`, `REJECT`,
`COMMIT`, `MERGE`, lifecycle transitions, approval, authorization, or an
execution-model class. `diagnostics` is advisory. Only a domain `choice` from the
request's allowed set is returned, and even a `HUMAN` choice is ordinary domain
data, not an approval.

**Note on `score`/`boolean`:** Clef supports them natively (§3.2), and the local
contract models them, but the **SEAM-002 wire capability is a bounded choice**
(only `choice`/`confidence` cross the SOP boundary). They are used for local
CLI/testing fidelity, not sent as SOP evidence.

---

## 8. Failure mapping (Clef-specific → generic provider contract)

Two deliberate channels (SEAM-002 §7): **in-result** failures the provider
observed (`status`), and **transport/process** failures that yield no usable
result (typed error at the SOP seam, i.e. a non-zero exit).

| Clef-specific failure | Local `ErrorKind` | SEAM-002 representation |
|---|---|---|
| runtime unavailable (oMLX/Ollama not running) | `unavailable` | process non-zero exit → typed transport error |
| model unavailable (Clef not installed) | `unavailable` | process non-zero exit |
| endpoint unavailable (route missing / 4xx-5xx) | `unavailable` / `provider_failure` | process non-zero exit |
| timeout (caller `ctx` / configured) | `unavailable` | process non-zero exit |
| cancellation | `unavailable` | process non-zero exit |
| malformed response (bad JSON / unknown answer type) | `malformed_response` | process non-zero exit |
| invalid choice (outside allowed) | `malformed_response` | process non-zero exit |
| invalid confidence (out of range / NaN / Inf) | `malformed_response` | process non-zero exit |
| unsupported operation / kind | adapter capability (no inference) | `status=UNSUPPORTED` + `error` (exit 0) |
| indeterminate result | adapter rule (nil confidence) | `status=INDETERMINATE` **or** `OK` with confidence omitted |
| transport failure (connection reset, EOF) | `unavailable` / `provider_failure` | process non-zero exit |
| model load failure | `unavailable` | process non-zero exit |

**Contract-sufficiency check (Phase 6 gate):** every real Clef failure above is
representable by the **SOP** contract (SEAM-002 status enum + typed transport
error) without inventing new SOP policy semantics. The only local limitation is
that `decision.DecisionResult` has no `status`; that is handled **adapter-side** by
the wire layer + capability rule, and does **not** require changing either
contract. **No STOP condition is triggered.**

---

## 9. Test strategy

Three levels (no real model required for ordinary gates).

**Level 1 — Unit (offline, no model).** Using a fake `Transport`:
- SEAM-002 request → local request → SystemOne request translation;
- SystemOne response → local answer → SEAM-002 result translation;
- confidence/probability bounds (`[0,1]`, NaN/Inf rejection);
- allowed-choice membership (reject an out-of-set choice);
- unsupported capability → `UNSUPPORTED`;
- indeterminate → `INDETERMINATE`/missing-confidence;
- malformed responses → fail closed;
- error translation → correct `ErrorKind` and process exit behavior.

**Level 2 — Provider contract (offline).** Prove Clef produces **exactly** the
provider-neutral SEAM-002 evidence already proven with SOP's fake external
provider (`agentic-sop/testdata/fake-decision-provider`):
- no SOP internal types required (adapter imports no `agentic-sop` package);
- no governance authority exposed (no `CONTINUE`/`APPROVE`/`COMMIT`/`MERGE`);
- no provider-specific policy required;
- the DTO validates under the SEAM-002 rules (version/status/choice/confidence);
- a second local provider (e.g. `nimble`) satisfies the same wire unchanged.

**Level 3 — Real local runtime (opt-in, `CLEF_INTEGRATION_TEST=1`).** Against
`clef-flash` over Ollama `/v1/systemone`. Measure request success rate, latency
(p50/p95), choice validity, confidence availability, calibration where measurable,
timeout behavior, malformed/indeterminate behavior, and repeatability. Real-model
tests are **not** part of the default unit gate and SKIP when the runtime is absent.

---

## 10. Shadow / evaluation strategy

Clef must operate in **shadow/evaluation**, never as an authoritative or default
provider. Note (from SOP L6): the SOP seam provides OFF and enabled (=live); it has
**no runtime shadow mode**. Therefore shadow evaluation is performed **adapter-side**
as an evaluation harness, not by wiring Clef live into SOP.

Plan:
- Compare Clef decisions against (a) deterministic expected cases, (b) existing
  SOP outcomes, and optionally (c) Nimble (same SystemOne API, comparable).
- Capture: agreement rate; disagreement cases; invalid-choice rate; indeterminate
  rate; latency p50/p95; timeout/failure rate; confidence distribution;
  confidence-vs-correctness where ground truth exists.
- **Provider disagreement is evidence for evaluation only. It must not
  automatically alter SOP policy.** The harness returns the primary result only and
  records the shadow result observationally.

---

## 11. Benchmark strategy (decision-provider suitability)

A focused matrix — **not** general LLM benchmarking:

| Provider / runtime | Quantization | Transport | Purpose |
|---|---|---|---|
| Clef (`clef-flash`) / Ollama | `mxfp8` | SystemOne `/v1/systemone` | **primary** — verified native decision path |
| Clef 4-bit / oMLX | `4-bit` (MLX) | OpenAI-compat + translation (if CLEF-002 proves it) | alternate transport fidelity |
| Clef 8-bit (`mlx-community/clef-8bit`) | 8-bit | MLX/oMLX or Ollama | quantization fidelity baseline |
| Nimble / Ollama | `Q8_0` | SystemOne `/v1/systemone` | comparable reference provider |

Metrics: choice validity, confidence availability, confidence calibration (where
ground truth exists), agreement with expected/SOP outcomes, latency p50/p95,
timeout/failure rate. Optional: Clef 8-bit as a quality/fidelity baseline if
evidence suggests quantization materially affects decision quality.

---

## 12. Adoption gates

```text
IMPLEMENTED → CONTRACT_VERIFIED → LOCAL_RUNTIME_VERIFIED → SHADOW_VERIFIED
            → BENCHMARKED → READY_FOR_SELECTABLE_REVIEW
```

There is **no automatic `DEFAULT` state**. Making Clef selectable, live, or
default requires separate evidence **and human approval**. Clef is OFF/non-default
until then.

---

## 13. Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R1 | oMLX's OpenAI-compatible output is mistaken for Clef decision semantics | High | Select Option C; treat Option B as REQUIRES EXPERIMENT with explicit fidelity tests (CLEF-002) |
| R2 | Confidence from Clef is uncalibrated | Medium | Evidence-only; SOP owns thresholds; measure calibration in CLEF-008; never assert certainty |
| R3 | Sampling/tokenization makes an OpenAI-translator non-deterministic | Medium | Prefer structured SystemOne path; if Option B is used, constrain generation and test determinism |
| R4 | Clef availability/latency differs from Nimble | Low-Med | Availability check without inference; timeout via `ctx`; opt-in live tests |
| R5 | Contract divergence (local vs SEAM-002) is misread as a defect | Medium | Documented in §1.3; adapter-side wire layer; no contract change |
| R6 | Shadow evaluation accidentally influencing policy | High | Harness is observational; returns primary only; SOP has no shadow mode |
| R7 | `clef-flash` vs `mlx-community/clef-4bit` treated as identical | Low | Recorded UNKNOWN; CLEF-002 confirms or rejects equivalence |
| R8 | Adapter leaks governance authority | Critical | Structural: DTO carries none; validated by SOP; guard tests in CLEF-005/009 |

---

## 14. Unresolved questions (for CLEF-002 / later)

1. Does `mlx-community/clef-4bit` materialize and run under oMLX at all in this
   environment (weights are currently a ref only)?
2. Can an oMLX OpenAI-compatible translation reproduce Clef's decision semantics
   faithfully (bounded choices, calibrated-enough probabilities, determinism)?
3. How does Clef behave on ties, refusals, empty choices, and out-of-domain
   questions (null/indeterminate handling)?
4. Is `clef-flash` (Ollama, `mxfp8`) the same base model as
   `mlx-community/clef-4bit` (MLX, 4-bit)? How does quantization affect decisions?
5. Does Clef's `confidence` track correctness well enough to be useful evidence
   (calibration)?
6. Should the SystemOne transport be shared with Nimble, or kept per-provider?
   (Current recommendation: keep it per-provider to avoid touching the proven
   Nimble reference; revisit in the backlog.)
7. What are Clef's real-world latency/timeout characteristics on this host?

---

## 15. Architecture guards and stop-condition review

Guards preserved by the design:

1. No Clef-specific change to `agentic-sop` policy. ✔ (adapter-side only)
2. No Clef-specific lifecycle behavior. ✔
3. No Clef-specific approval behavior. ✔
4. Clef output remains evidence. ✔
5. SOP validates and interprets evidence. ✔ (SOP `ValidateResult` + autonomy)
6. Clef cannot authorize execution. ✔
7. Clef cannot authorize commit/merge. ✔
8. Clef cannot satisfy human approval. ✔
9. Decision-provider selection stays separate from SMALL/MEDIUM/LARGE routing. ✔
10. Clef is OFF/non-default until separately approved. ✔
11. Another provider can implement the same generic contract without SOP policy
    changes. ✔ (provider-neutral `serve` wire layer)
12. oMLX/Ollama/SystemOne details remain adapter-side. ✔

Stop-condition review:

- Clef requires provider-specific SOP policy? **No.**
- Clef requires lifecycle authority? **No.**
- Clef requires approval authority? **No.**
- Existing generic provider contract insufficient? **No** — the SOP contract
  (SEAM-002) fully represents Clef outcomes; the local result's lack of a `status`
  is bridged adapter-side (§1.3, §8).
- oMLX compatibility assumed rather than demonstrated? **No** — recorded as
  UNSUPPORTED-native/UNKNOWN-translation; the verified path is Ollama SystemOne.
- Clef cannot produce the minimum provider-neutral evidence? **No** — it produces
  choice + confidence natively (verified).
- Implementation would require changing execution-model routing? **No.**

**No STOP condition was triggered.**

---

## 16. Verification and scope statement

- This report is **documentation only**; `git diff` for the effort contains new
  files under `docs/reports/clef-provider/` and `docs/plans/` — no production file
  changed.
- `sop-decision-adapters` HEAD at capture: `a419ce5`. `agentic-sop` was **not**
  modified; it was inspected read-only at HEAD `1153913`.
- No Clef provider, transport, or wire layer was implemented; no commit or push was
  performed.
- The plan that consumes this analysis is
  [`../../plans/PLAN-Clef-Provider-Integration.md`](../../plans/PLAN-Clef-Provider-Integration.md).
