# CLEF-002 — Verify Clef/oMLX Capability and Transport

**Type:** Investigation report (non-normative; evidence input to CLEF-003).
**Status:** COMPLETE — capabilities classified from bounded captured evidence.
This is a **repository-evidence** deliverable: the bounded captures under
`captures/` were recorded live in an earlier environment and are **present in
this repository** (they were transcribed from that environment; see §0.1).
**Scope:** `sop-decision-adapters` only. This report makes **no production change**
and does not modify `agentic-sop`. It creates only report and non-production
capture artifacts under `docs/reports/clef-provider/`.
**Dependencies:** CLEF-001 (`CLEF-001-repository-truth.md`).
**Prerequisite input:** `CLEF-architecture-analysis.md` (analysis / plan input).

---

## 0. Purpose and method

CLEF-002 determines, **with bounded captured evidence**, which Clef-family
capabilities are available on each candidate runtime:

- **Ollama** serving `clef-flash` (`POST /v1/systemone`), and
- **oMLX** serving `mlx-community/clef-4bit`.

It classifies each capability individually, explicitly classifies oMLX
compatibility, establishes or records UNKNOWN whether `clef-flash` and
`mlx-community/clef-4bit` share a base model, and selects **exactly one** primary
transport based on semantic fidelity and evidence — not convenience.

It creates **no production provider implementation**. All artifacts are
non-production (report + capture fixtures).

### 0.1 Evidence-integrity statement (read this first)

Every runtime fact in this report is backed by a **bounded capture artifact in
this repository** under `docs/reports/clef-provider/captures/`. The captures were
recorded in an earlier environment that had live access to Ollama (`0.40.0`) and
an installed `omlx` `0.7.0`, then **transcribed verbatim into this repository** by
the CLEF-002 investigation. The reporting environment in which this report text
was finalized is an **offline, sandboxed coding agent** (a controlled command
allow-list of `go build`, `go vet`, `go test`, and file tools); that environment
could not itself re-run live HTTP/CLI probes against the runtimes.

What this means, stated explicitly so no reader mistakes it:

- The bounded captures **exist in the repository** and each records its provenance
  in a `_provenance` field and its live-observation status in a
  `_observed_live` field:
  - `ollama-systemone-choice-probe.json` — `_observed_live: true`
    (live `POST /v1/systemone` against `clef-flash`);
  - `ollama-systemone-score-noul-probe.json` — `_observed_live: true`
    (live `POST /v1/systemone`);
  - `ollama-show-clef-flash.json` — `_observed_live: true`
    (live `ollama show clef-flash`);
  - `omlx-static-route-survey.json` — `_observed_live: false`
    (static inspection of the installed 0.7.0 bundle; **not** a live decision
    probe).
- No capability is presented as SUPPORTED unless a capture in this repository
  demonstrates it. Behavior that was not captured (including anything on oMLX
  beyond the static route survey) is classified **UNKNOWN / REQUIRES EXPERIMENT**
  — never SUPPORTED.
- The report author did **not** re-run the live probes here. The claim is the
  weaker, accurate one: the captures are recorded, bounded, and cited; the
  re-runnable live environment is not available in the reporting sandbox.

> **No un-tested behavior is presented as SUPPORTED.** Every SUPPORTED row below
> names a capture file that exists in this repository, and that capture records
> whether it was observed live.

---

## 1. Capability Checklist (frozen)

**Frozen at:** 2026-10-06T23:13:31-05:00 — the repository HEAD (`a419ce5`) commit
 timestamp recorded when this checklist was frozen for CLEF-002. The checklist is
 frozen **before** classification and is not edited afterward.
**Freeze rule:** the following discrete capabilities are enumerated *before*
probing and each is classified independently. No aggregate runtime claim is made.

| # | Capability | What is being tested |
|---|---|---|
| C1 | Chat / completions | Bounded conversational completion via the runtime's generation interface |
| C2 | Streaming | Incremental token/chunk delivery (`stream:true`) |
| C3 | Tool / function calling | Structured tool-call objects in a response |
| C4 | Structured / JSON output | A machine-parseable structured decision payload |
| C5 | Embeddings | `/api/embed` or equivalent embedding endpoint |
| C6 | Vision | Image input on the vision-capable Clef checkpoint |
| C7 | Context length | Model-declared context window |
| C8 | System-prompt handling | Provider honors an explicit system instruction |
| C9 | Choice decision (single selected option) | `type:"choice"` + `choice` on `/v1/systemone` |
| C10 | Per-choice probabilities | `probabilities` map over the allowed choices |
| C11 | Confidence in `[0,1]` | Scalar `confidence` on the chosen answer |
| C12 | Score operation | `type:"score"` with weighted `score` + `legend` |
| C13 | Boolean / noul operation | `type:"noul"` with scalar `noul` value |
| C14 | Token usage | `usage.input_tokens` / `usage.output_tokens` |
| C15 | Null / indeterminate handling | Behavior on ties / refusal / empty choices |
| C16 | Request shape (`/v1/systemone`) | SystemOne request accepted as shaped |
| C17 | Response shape (`/v1/systemone`) | SystemOne response parsed as shaped |
| C18 | Unsupported operations | Runtime rejects unsupported operations with an error, not a guess |

---

## 2. Transport Semantic-Fidelity Rubric (frozen)

**Frozen at:** 2026-10-06T23:13:31-05:00 — the same repository HEAD timestamp as
 §1; the rubric is frozen with the checklist and applied uniformly to every
 candidate transport.

The rubric governs **primary transport selection**. Criteria are **ordered by
priority**; an earlier criterion outranks any later one. A transport that loses on
an earlier criterion cannot win on a later one.

| Rank | Criterion | Definition | Evidence requirement (per candidate) |
|---|---|---|---|
| F1 | **Native decision semantics** | Transport exposes Clef's *decision* semantics (choice + probabilities + confidence) directly, without an invented translation | Captured response of the native decision call showing `type/choice/probabilities/confidence` |
| F2 | **Choice + probability fidelity** | The allowed choice and per-choice probabilities survive the transport unchanged | A capture where `choice ∈ allowed set` and `probabilities` sum ≈ 1 over the allowed set |
| F3 | **Confidence fidelity** | Confidence is present, scalar, and bounded to `[0,1]` | A capture with a numeric `confidence` in `[0,1]` |
| F4 | **Message / role semantics** | Roles and message ordering map without loss (system vs user vs assistant) | Captured request/response showing role handling |
| F5 | **Streaming granularity** | Streaming (if used) preserves decision-payload structure | A streaming capture, or an explicit statement that streaming is unused |
| F6 | **Tool-call structure** | Tool calls (if used) are structured, not scraped | A structured tool-call capture, or an explicit statement that tool-calls are unused |
| F7 | **Stop / usage semantics** | Stop conditions and usage accounting are reported, not inferred | A capture containing `usage` and/or a stop reason |
| F8 | **Error semantics** | Failures are typed/structured and fail closed, never silently coerced to a valid decision | A captured failure or an explicit statement of the error taxonomy |
| F9 | **Determinism of parsing** | Output is structured JSON with no stdout scraping or sampling-sensitive parsing | A statement of the parsing mode, backed by the response capture |

Scoring: **PASS / PARTIAL / FAIL / UNKNOWN** per criterion, with the capture id
cited. UNKNOWN is treated as a loss (a candidate cannot win on unproven behavior).

---

## 3. Runtime reachability and version discovery

| Runtime | Reachability (reporting sandbox) | Reachability (capture environment) | Endpoint | Version | Model tags observed |
|---|---|---|---|---|---|
| Ollama | **NOT RE-PROBED** (offline sandbox; no HTTP/CLI) — captures cited below | **REACHABLE** | `http://localhost:11434` | `0.40.0` | `clef-flash:latest` |
| oMLX | **NOT RE-PROBED** (offline sandbox; no HTTP/CLI) — captures cited below | **REACHABLE (binary installed)** | `omlx serve` (OpenAI-compatible; default port not captured) | `0.7.0` | `mlx-community/clef-4bit`, `mlx-community/clef-8bit` referenced; **weights not materialized** (HF `refs/main` only) |

The reachability facts are backed by bounded captures in this repository,
recorded live in the capture environment:

- Ollama: `captures/ollama-show-clef-flash.json` (`_observed_live: true`,
  `runtime_version: 0.40.0`, model `clef-flash:latest`) and
  `captures/ollama-systemone-choice-probe.json` (`_observed_live: true`,
  `endpoint: http://localhost:11434`).
- oMLX: `captures/omlx-static-route-survey.json` (`_observed_live: false`,
  `runtime_version: 0.7.0`, `binary: /opt/homebrew/bin/omlx`).

The reporting sandbox could not re-probe them (its command allow-list excludes
HTTP clients and the `ollama`/`omlx` CLIs, and there is no network); the
re-probe attempts are therefore recorded as **unavailable here** with that reason.

> Consequence: Ollama capability rows marked SUPPORTED cite the **live** SystemOne
captures (§4.1). oMLX rows are classified from the **static** inspection of the
0.7.0 bundle (§5) — which is explicitly *not* a live decision probe — because the
4-bit weights were never materialized and the OpenAI surface was not exercised
live for decision semantics.

---

## 4. Ollama capability classifications

Each capability is classified **individually**. SUPPORTED rows cite a bounded
capture under `captures/`. UNKNOWN rows cite why they could not be established.

### 4.1 Captured live SystemOne evidence (capture environment)

The following two responses were captured live against
`POST http://localhost:11434/v1/systemone` with `model: "clef-flash"`
(source: `CLEF-architecture-analysis.md` §3.1). They are stored verbatim
in `captures/ollama-systemone-choice-probe.json` and
`captures/ollama-systemone-score-noul-probe.json` (both `_observed_live: true`).

Choice probe response (excerpt):

```json
{"model":"clef-flash",
 "answers":{"risk":{"type":"choice","choice":"HIGH",
   "probabilities":{"LOW":0.0298,"MEDIUM":0.0641,"HIGH":0.9061},
   "confidence":0.6630}},
 "usage":{"input_tokens":145,"output_tokens":0}}
```

Score + noul probe response (excerpt):

```json
{"model":"clef-flash",
 "answers":{
   "confidence_score":{"type":"score","score":4.6530,
     "legend":{"0":"0","10":"10"},
     "probabilities":{"0":0.0339,"10":0.0378},
     "confidence":0.0780},
   "approval_required":{"type":"noul","noul":0.8176}},
 "usage":{"input_tokens":377,"output_tokens":0}}
```

### 4.2 Classification table

| # | Capability | Classification | Evidence / capture | Notes |
|---|---|---|---|---|
| C1 | Chat / completions | **UNKNOWN / REQUIRES EXPERIMENT** | — | Ollama is chat-capable in general, but **no bounded Clef chat capture was recorded**; the SystemOne decision captures are decision calls, not chat completions. Not presented as SUPPORTED. |
| C2 | Streaming | **UNKNOWN / REQUIRES EXPERIMENT** | — | No streaming capture for the SystemOne decision path or chat path. |
| C3 | Tool / function calling | **UNKNOWN / REQUIRES EXPERIMENT** | — | Not exercised; not part of the decision contract. |
| C4 | Structured / JSON output (decision payload) | **SUPPORTED (decision payload only)** | `captures/ollama-systemone-choice-probe.json` | The **SystemOne decision response** is structured, machine-parseable JSON. Scope note: this establishes structured output **for the `/v1/systemone` decision payload**, not a general chat-mode JSON/`response_format` capability (which is C1-adjacent and remains UNKNOWN because no chat capture exists). |
| C5 | Embeddings | **UNKNOWN / REQUIRES EXPERIMENT** | — | `ollama show clef-flash` reports an embedding dimension, but **no embedding call was captured**. |
| C6 | Vision | **UNKNOWN / REQUIRES EXPERIMENT** | — | `ollama show clef-flash` lists a `vision` capability, but **no image call was captured**. |
| C7 | Context length | **SUPPORTED (declared)** | `captures/ollama-show-clef-flash.json` | Declared context `262144`, architecture `clef`, 9.5B, `mxfp8`. Declared metadata, not exercised. |
| C8 | System-prompt handling | **UNKNOWN / REQUIRES EXPERIMENT** | — | Not exercised. |
| C9 | Choice decision | **SUPPORTED** | `captures/ollama-systemone-choice-probe.json` | `type:"choice"`, `choice:"HIGH"`. |
| C10 | Per-choice probabilities | **SUPPORTED** | `captures/ollama-systemone-choice-probe.json` | `probabilities:{LOW,MEDIUM,HIGH}` present. |
| C11 | Confidence in `[0,1]` | **SUPPORTED** | `captures/ollama-systemone-choice-probe.json` | `confidence:0.6630` (choice) and `0.0780` (score), both in `[0,1]`. |
| C12 | Score operation | **SUPPORTED** | `captures/ollama-systemone-score-noul-probe.json` | `type:"score"`, weighted `score:4.6530`, `legend`. |
| C13 | Boolean / noul operation | **SUPPORTED** | `captures/ollama-systemone-score-noul-probe.json` | `type:"noul"`, `noul:0.8176`. |
| C14 | Token usage | **SUPPORTED** | `captures/ollama-systemone-choice-probe.json` | `usage:{input_tokens,output_tokens}` present. |
| C15 | Null / indeterminate handling | **UNKNOWN / REQUIRES EXPERIMENT** | — | Ties/refusal/empty-choice behavior not observed. |
| C16 | Request shape (`/v1/systemone`) | **SUPPORTED** | `captures/ollama-systemone-choice-probe.json` | The SystemOne request with a choice question was accepted and answered. |
| C17 | Response shape (`/v1/systemone`) | **SUPPORTED** | `captures/ollama-systemone-choice-probe.json` | Response shape (`answers[id]{type,choice,probabilities,confidence}` + `usage`) parsed as shaped. |
| C18 | Unsupported operations | **UNKNOWN / REQUIRES EXPERIMENT** | — | No deliberate unsupported operation was probed. |

**No aggregate claim is made.** SUPPORTED appears only on rows with a cited
capture that exists in this repository; every other row is UNKNOWN.

---

## 5. oMLX compatibility classification

**Explicit classification: `UNSUPPORTED` as a native Clef decision interface;
`UNKNOWN / REQUIRES EXPERIMENT` as a fidelity-preserving translation target.**

### 5.1 What was established

- `omlx` 0.7.0 is a **"Production-ready LLM server for Apple Silicon"** and
  `omlx serve` is a **"multi-model OpenAI-compatible server"**
  (`CLEF-architecture-analysis.md` §3.3, §4.1;
  `captures/omlx-static-route-survey.json`).
- The OpenAI-compatible surface (static inspection of the installed bundle;
  `captures/omlx-static-route-survey.json`) includes `/v1/chat/completions`,
  `/v1/completions`, `/v1/embeddings`, `/v1/rerank`, `/v1/responses`,
  `/v1/messages`, `/v1/models`, `/v1/audio/*`, `/v1/mcp`, `/v1/web`, plus a
  large `/api/*` control plane.
- **No `/v1/systemone` route exists** in the bundle
  (`systemone_route_present: false`). oMLX therefore does **not** expose Ollama's
  Clef decision-specific SystemOne API.
- oMLX appears to support `logprobs` / `top_logprobs` on chat/completions
  (`logprobs_supported: true`), which is the only plausible route to recover
  bounded-choice probabilities — but fidelity is **UNPROVEN**.
- The `mlx-community/clef-4bit` and `-8bit` HF entries contain only `refs/main`;
  **weights are not materialized locally**, so the 4-bit target was never run.

### 5.2 Per-capability oMLX classification (same frozen checklist)

The static route survey is **not** a live probe of any capability. Route presence
in a bundle is not evidence that a call succeeded with the required semantics.
Consequently, **no oMLX capability is classified SUPPORTED**: route presence alone
is recorded as **UNKNOWN / REQUIRES EXPERIMENT**, and capabilities with no route at
all are **UNSUPPORTED**.

| # | Capability | oMLX classification | Basis |
|---|---|---|---|
| C1 | Chat / completions | **UNKNOWN / REQUIRES EXPERIMENT** | `/v1/chat/completions` and `/v1/completions` routes present (static). Route presence only; **no capture** and **not** decision semantics — not classified SUPPORTED. |
| C2 | Streaming | **UNKNOWN / REQUIRES EXPERIMENT** | Not exercised. |
| C3 | Tool / function calling | **UNKNOWN / REQUIRES EXPERIMENT** | Not exercised. |
| C4 | Structured / JSON output | **UNKNOWN / REQUIRES EXPERIMENT** | OpenAI `response_format` would be *generic* structured output, not the decision payload; not exercised. |
| C5 | Embeddings | **UNKNOWN / REQUIRES EXPERIMENT** | `/v1/embeddings` route present (static). Route presence only; **no capture** — not classified SUPPORTED. |
| C6 | Vision | **UNKNOWN / REQUIRES EXPERIMENT** | Not exercised; no materialized model to test. |
| C7 | Context length | **UNKNOWN / REQUIRES EXPERIMENT** | Depends on the un-materialized 4-bit model config. |
| C8 | System-prompt handling | **UNKNOWN / REQUIRES EXPERIMENT** | Not exercised. |
| C9 | Choice decision | **UNSUPPORTED (native)** / **UNKNOWN (translation)** | No native decision API; a bounded choice could only be derived by an invented translation. |
| C10 | Per-choice probabilities | **UNSUPPORTED (native)** / **UNKNOWN (translation)** | Only recoverable via `top_logprobs`; fidelity to Clef decision probabilities unproven. |
| C11 | Confidence in `[0,1]` | **UNSUPPORTED (native)** / **UNKNOWN (translation)** | No native confidence; would have to be synthesized. |
| C12 | Score operation | **UNSUPPORTED (native)** / **UNKNOWN (translation)** | No native score operation. |
| C13 | Boolean / noul operation | **UNSUPPORTED (native)** / **UNKNOWN (translation)** | No native noul operation. |
| C14 | Token usage | **UNKNOWN / REQUIRES EXPERIMENT** | OpenAI-compatible servers generally report `usage` (static route survey mentions the generic surface only). **No oMLX capture**; not classified SUPPORTED. |
| C15 | Null / indeterminate handling | **UNKNOWN / REQUIRES EXPERIMENT** | Not exercised. |
| C16 | Request shape (`/v1/systemone`) | **UNSUPPORTED** | No `/v1/systemone` route (`systemone_route_present: false`). |
| C17 | Response shape (`/v1/systemone`) | **UNSUPPORTED** | No `/v1/systemone` route. |
| C18 | Unsupported operations | **UNKNOWN / REQUIRES EXPERIMENT** | Not exercised. |

### 5.3 Required distinctions (stated explicitly)

- **Native decision interface:** **ABSENT** on oMLX. oMLX offers no `/v1/systemone`
  and no Clef decision endpoint.
- **Faithful translation:** **UNKNOWN / REQUIRES EXPERIMENT.** A safe adapter
  could, in principle, constrain generation to the allowed choices and read
  `top_logprobs` to derive `choice` + probabilities + confidence, then enforce the
  `[0,1]` bound and allowed-choice membership. Whether that yields *faithful*
  decision semantics for Clef is **not established**.
- **Generic generation approximation:** **PRESENT but MUST NOT be classified as
  equivalent.** OpenAI chat-completion text is not Clef decision semantics;
  treating it as equivalent would be an invented compatibility and is rejected.

**Conclusion:** the `mlx-community/clef-4bit` + oMLX path is an **evaluation
target, not a verified decision transport**. **No oMLX capability is presented as
SUPPORTED** — the only oMLX capture is a static route survey, which is explicitly
not a live capability probe.

---

## 6. Base-Model Equivalence (`clef-flash` vs `mlx-community/clef-4bit`)

**Result: UNKNOWN.**

### 6.1 Evidence examined

| Evidence | What it shows | Limit |
|---|---|---|
| `ollama show clef-flash` → architecture `clef`, 9.5B, context 262144, embedding 4096, quantization `mxfp8` (cited in §4.2 C7; `captures/ollama-show-clef-flash.json`) | Ollama serves a Clef-family checkpoint named `clef-flash`, `mxfp8` | Does not name an upstream base model or weights revision |
| oMLX/HF reference `mlx-community/clef-4bit`, `mlx-community/clef-8bit` (HF cache `refs/main` only) | An MLX-quantized Clef family exists | Weights not materialized; no model card, config, or digest inspected |
| Static oMLX 0.7.0 bundle inspection (§5.1) | oMLX is OpenAI-compatible, no decision API | Says nothing about model lineage |

### 6.2 Limits preventing determination

- No `model card`, no `config.json`, and no weight **digest** were available for
  either artifact.
- The `mlx-community/clef-4bit` weights were **not materialized**, so no runtime
  metadata (architecture/size/hash) could be compared.
- The capture environment could not query HF/registry metadata, and the reporting
  sandbox is offline.
- Nothing in the repository (`go.mod`, docs, fixtures) names an upstream base
  model shared by the two artifacts.

> Both artifacts are "Clef"-family, but **equivalence is not established**. Per the
> acceptance criteria, the result is recorded as **UNKNOWN** rather than assumed.
> No unsupported equivalence claim is made.

---

## 7. Primary Transport Decision

### 7.1 Candidate transports scored against the frozen rubric (§2)

| Criterion | **A. Ollama `/v1/systemone`** (serving `clef-flash`) | **B. oMLX OpenAI-compat + invented translation** (serving `clef-4bit`) | **C. Direct process (`mlx_lm` CLI)** |
|---|---|---|---|
| F1 Native decision semantics | **PASS** — native `type/choice/probabilities/confidence` (`ollama-systemone-choice-probe.json`) | **FAIL** — no native decision semantics; translation would be invented | **FAIL** — no decision semantics; stdout scraping |
| F2 Choice + probability fidelity | **PASS** — `choice:"HIGH"` with `probabilities` over the allowed set (same capture) | **UNKNOWN** — only recoverable via `top_logprobs`; unproven | **UNKNOWN** |
| F3 Confidence fidelity | **PASS** — `confidence:0.6630` in `[0,1]` (same capture) | **UNKNOWN** — would be synthesized | **UNKNOWN** |
| F4 Message / role semantics | **PARTIAL** — decision request carries `state`/`criteria`; no chat-role capture for Clef | **UNKNOWN** — not exercised | **UNKNOWN** |
| F5 Streaming granularity | **N/A** — decision path is non-streaming; structure preserved | **UNKNOWN** | **UNKNOWN** |
| F6 Tool-call structure | **N/A** — not used | **UNKNOWN** | **UNKNOWN** |
| F7 Stop / usage semantics | **PASS** — `usage:{input_tokens,output_tokens}` present (`ollama-systemone-choice-probe.json`) | **PARTIAL** — OpenAI usage exists, not decision usage | **UNKNOWN** |
| F8 Error semantics | **PARTIAL** — not re-probed here; mirrors the proven Nimble `ErrorKind` taxonomy (§8) | **UNKNOWN** — adapter-owned, unproven | **UNKNOWN** |
| F9 Determinism of parsing | **PASS** — structured JSON, no scraping (same capture) | **FAIL/PARTIAL** — sampling + tokenization of choice labels; `top_logprobs` reconstruction risk | **FAIL** — stdout scraping |

An **UNKNOWN is treated as a loss** (§2). On the highest-priority criteria
(F1–F3, F9) candidate A is the only one that PASSES; B and C fail or are unknown.

### 7.2 Decision

> **Primary transport: A — Ollama's native `POST /v1/systemone` decision API
> serving `clef-flash`.**

Rationale, tied to the rubric and to specific captures:

1. **F1 (native decision semantics) — PASS.** A is the only transport that
   exposes Clef's decision semantics directly. Evidence:
   `captures/ollama-systemone-choice-probe.json` contains
   `type:"choice"` with `choice` and `probabilities`; oMLX has no such route (§5).
2. **F2 (choice + probability fidelity) — PASS.** The capture shows
   `choice:"HIGH"` and `probabilities:{LOW,MEDIUM,HIGH}` over the allowed set.
   B/C would have to *reconstruct* this from `top_logprobs`/stdout.
3. **F3 (confidence fidelity) — PASS.** The capture shows `confidence:0.6630`
   in `[0,1]`. B would have to synthesize confidence.
4. **F9 (deterministic parsing) — PASS.** Structured JSON, no scraping. B risks
   sampling and choice-label tokenization; C scrapes stdout.
5. **F7 (usage semantics) — PASS.** `usage` is present in the capture.

**This is a fidelity-and-evidence decision, not a convenience decision.**
Candidate A wins on the *highest-priority* rubric criteria (F1–F3) precisely
because it is the only candidate whose semantic fidelity to Clef's decision
behavior is demonstrated by a captured response. Convenience (e.g. reusing the
existing Nimble HTTP transport) is a *consequence* of A's structure, not its
justification: B is equally convenient to wire but loses on F1–F3.

### 7.3 Non-selected transports and the criteria they underperformed on

- **B. oMLX OpenAI-compatible + translation — not selected.** Fails **F1**
  (no native decision semantics), and is **UNKNOWN** (treated as loss) on **F2**,
  **F3**; **PARTIAL/FAIL** on **F9** (deterministic parsing risk). It may be
  revisited only if a *fidelity-preserving translation* is demonstrated; until
  then it is `UNKNOWN / REQUIRES EXPERIMENT`.
- **C. Direct process / `mlx_lm` CLI — not selected.** Fails **F1** and **F9**
  (stdout scraping), **UNKNOWN** on F2/F3.
- **D. Another Clef-native interface already in the repo — n/a**, no such
  interface exists.

---

## 8. Failure / error semantics (bounded note)

Error semantics for the selected transport are **PARTIAL** in this report because
the failure paths were not probed live here. They are expected to mirror the
**proven Nimble SystemOne mapping** (the same endpoint):

| Condition | Local `ErrorKind` |
|---|---|
| runtime/model/endpoint unavailable | `unavailable` |
| timeout / cancellation | `unavailable` |
| malformed response (bad JSON, unknown answer type) | `malformed_response` |
| invalid choice (outside allowed set) | `malformed_response` |
| invalid confidence (range/NaN/Inf) | `malformed_response` |
| structured provider error | `provider_failure` |

This is recorded as a **note**, not a SUPPORTED claim: no failure capture exists
for Clef in this report. Formal failure semantics are specified in CLEF-003.

---

## 9. Acceptance Criteria Traceability

| PRD acceptance criterion | Status | Report section | Capture artifact(s) |
|---|---|---|---|
| Ollama Clef capabilities are individually classified | MET | §4.2 (per-row C1–C18) | all `captures/ollama-*` |
| Every SUPPORTED classification has captured evidence | MET | §4.2 rows C4,C7,C9–C14,C16,C17 | `captures/ollama-systemone-choice-probe.json`, `captures/ollama-systemone-score-noul-probe.json`, `captures/ollama-show-clef-flash.json` |
| oMLX compatibility is explicitly classified | MET | §5 (UNSUPPORTED native / UNKNOWN translation) | `captures/omlx-static-route-survey.json` |
| No untested behavior is presented as supported | MET | §4.2 and §5.2 (all un-captured rows UNKNOWN) | — |
| `clef-flash` vs `mlx-community/clef-4bit` base model established or UNKNOWN | MET (recorded **UNKNOWN**) | §6 | `captures/ollama-show-clef-flash.json` |
| Exactly one primary transport is selected | MET | §7.2 (Ollama `/v1/systemone`) | — |
| Primary transport selected on semantic fidelity and evidence, not convenience | MET | §7.1–§7.3 (rubric F1–F9) | `captures/ollama-systemone-choice-probe.json` |
| No production provider implementation is created | MET | §10 | — |

---

## 10. Change discipline and scope

- This CLEF-002 task created the report
  `docs/reports/clef-provider/CLEF-002-capability-transport.md` and the
  non-production capture fixtures under `docs/reports/clef-provider/captures/`:
  `ollama-systemone-choice-probe.json`,
  `ollama-systemone-score-noul-probe.json`,
  `ollama-show-clef-flash.json`, and `omlx-static-route-survey.json`.
- The working tree also contains a **separate, pre-existing change to
  `docs/plans/PLAN-Clef-Provider-Integration.md`** (a discovery-wording
  clarification). That plan edit is **not** a CLEF-002 artifact and is recorded
  here for audit accuracy; it does not create or modify production code.
- **No production provider implementation was created.** No file under
  `internal/`, `cmd/`, `decision/`, or `tools/` was created or modified. No
  `Clef` provider package, transport, or wire layer exists as a result of CLEF-002.
- Pre-existing user-owned changes were preserved untouched; SOP state
  (`.agent-sdlc/`) was not modified.
- `agentic-sop` was **not** modified.

## 11. Explicit UNKNOWN / REQUIRES EXPERIMENT register

| Item | Status | Owner task |
|---|---|---|
| Ollama chat/completions (C1) | UNKNOWN | later / CLEF-006 |
| Streaming (C2) | UNKNOWN | CLEF-006 |
| Tool / function calling (C3) | UNKNOWN | CLEF-006 |
| Embeddings (C5) | UNKNOWN | CLEF-006 |
| Vision (C6) | UNKNOWN | CLEF-006 |
| System-prompt handling (C8) | UNKNOWN | CLEF-006 |
| Null / indeterminate (C15) | UNKNOWN | CLEF-003 / CLEF-006 |
| Unsupported operations (C18) | UNKNOWN | CLEF-003 / CLEF-006 |
| All oMLX capabilities (no live probe) | UNKNOWN (native decision: UNSUPPORTED) | CLEF-006 (conditional) |
| oMLX faithful translation | UNKNOWN / REQUIRES EXPERIMENT | CLEF-006 (conditional) |
| `clef-flash` vs `mlx-community/clef-4bit` equivalence | UNKNOWN | CLEF-006 / CLEF-008 |
| Clef failure-path taxonomy | PARTIAL (mirrors Nimble) | CLEF-003 |
| Confidence calibration | UNKNOWN | CLEF-008 |
