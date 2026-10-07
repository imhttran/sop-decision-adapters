# CLEF-002 — Verify Clef/oMLX Capability and Transport

**Type:** Investigation report (non-normative; evidence input to CLEF-003).
**Status:** COMPLETE — every classification is backed by a live capture taken
on 2026-10-07 (operator-run probe `tools/clef/probe_systemone.py`); the per-criterion
acceptance evidence map is §9.1.
**Scope:** `sop-decision-adapters` only. No production change. `agentic-sop` is not
modified. Artifacts: this report, `captures/live/*.json`, and the probe script.
**Dependencies:** CLEF-001 (`CLEF-001-repository-truth.md`).
**Prerequisite input:** `CLEF-architecture-analysis.md` (analysis / plan input).

---

## 0. Purpose and method

CLEF-002 determines, with bounded live captures, which Clef capabilities are
available on each candidate runtime:

- **Ollama** serving `clef-flash` (`POST /v1/systemone`), and
- **oMLX** serving `mlx-community/clef-4bit`.

It classifies each capability individually, explicitly classifies oMLX
compatibility, establishes whether `clef-flash` and `mlx-community/clef-4bit`
share a base model, and selects exactly one primary transport on semantic
fidelity and evidence.

### 0.1 Evidence provenance and supersession

Every runtime fact below cites a file under `captures/live/`. Each capture holds
the exact request sent, HTTP status, elapsed time, UTC timestamp, the response
body (bounded to 8192 bytes; `_truncated` records whether the bound was hit — no
capture was truncated), and, for answers carrying probabilities, `invariants`
recomputed from the response itself.

Re-run from the repository root (Ollama on `:11434`, oMLX on `:8000`):

```bash
python3 tools/clef/probe_systemone.py
```

**Superseded evidence.** An earlier revision of this report (commit `79569b8`)
cited four captures transcribed from `CLEF-architecture-analysis.md` §3.1 in an
offline environment. Live re-probing found them unfaithful:

| Earlier claim                                                                | Live result                                                                                           | Capture                                |
| ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------- |
| Choice request shape `{"criteria":{"risk":{"type":"choice","choices":{…}}}}` | **HTTP 400** `questions must contain 1–64 fields`                                                     | `reject-transcribed-choice-shape.json` |
| Score request with `legend:{"0":"0","10":"10"}`                              | **HTTP 400** `score criteria must be an array`                                                        | `reject-score-legend.json`             |
| `mlx-community/clef-4bit` weights not materialized                           | Weights present at `~/.omlx/omlx/models/mlx-community/clef-4bit` (3 shards, 15 GB) and listed by oMLX | `omlx-models-status.json`              |
| oMLX route survey from static bundle inspection                              | Replaced by live `GET /v1/models`, `/v1/models/status`, `POST /v1/systemone`                          | `omlx-*.json`                          |

The analysis document recorded only response bodies; the request bodies in the
transcribed captures were reconstructed and never sent. Those four files were
removed and replaced by `captures/live/`. The analysis document itself is not a
CLEF-002 deliverable and is left unchanged; its §3.3/§4 "weights not
materialized" statement is superseded by this report.

### 0.2 Current runtime re-verification (2026-10-07)

The oMLX runtime state was re-verified live after the first CLEF-002 capture set.
`mlx-community/clef-4bit` is now **downloaded and loaded** by oMLX and is
API-visible: `GET /v1/models` lists `clef-4bit` and `clef-8bit`
(`omlx-models.json`), and `GET /v1/models/status` reports `clef-4bit` with
`model_path=/Users/imhttran/.omlx/omlx/models/mlx-community/clef-4bit` and
`loaded: true` (`omlx-models-status.json`).

This supersedes the earlier "not loaded" observation in §3. The historical
statements are retained verbatim for provenance; they are not rewritten. The §5
classification is unaffected: loading the checkpoint changes availability, not the
transport conclusion — the decision semantics still live in the joint head that
oMLX does not execute, and `POST /v1/systemone` on oMLX still returns 404
(`omlx-systemone-route.json`).

Reproduce this re-verification with the operator-run probe:

```bash
python3 tools/clef/probe_systemone.py
```

---

## 1. Capability Checklist (frozen)

**Frozen at:** 2026-10-06T23:13:31-05:00 — the repository HEAD (`a419ce5`) commit
timestamp recorded when this checklist was frozen for CLEF-002. The checklist is
frozen **before** classification and is not edited afterward.
**Freeze rule:** the following discrete capabilities are enumerated _before_
probing and each is classified independently. No aggregate runtime claim is made.

| #   | Capability                               | What is being tested                                                     |
| --- | ---------------------------------------- | ------------------------------------------------------------------------ |
| C1  | Chat / completions                       | Bounded conversational completion via the runtime's generation interface |
| C2  | Streaming                                | Incremental token/chunk delivery (`stream:true`)                         |
| C3  | Tool / function calling                  | Structured tool-call objects in a response                               |
| C4  | Structured / JSON output                 | A machine-parseable structured decision payload                          |
| C5  | Embeddings                               | `/api/embed` or equivalent embedding endpoint                            |
| C6  | Vision                                   | Image input on the vision-capable Clef checkpoint                        |
| C7  | Context length                           | Model-declared context window                                            |
| C8  | System-prompt handling                   | Provider honors an explicit system instruction                           |
| C9  | Choice decision (single selected option) | `type:"choice"` + `choice` on `/v1/systemone`                            |
| C10 | Per-choice probabilities                 | `probabilities` map over the allowed choices                             |
| C11 | Confidence in `[0,1]`                    | Scalar `confidence` on the chosen answer                                 |
| C12 | Score operation                          | `type:"score"` with weighted `score` + `legend`                          |
| C13 | Boolean / noul operation                 | `type:"noul"` with scalar `noul` value                                   |
| C14 | Token usage                              | `usage.input_tokens` / `usage.output_tokens`                             |
| C15 | Null / indeterminate handling            | Behavior on ties / refusal / empty choices                               |
| C16 | Request shape (`/v1/systemone`)          | SystemOne request accepted as shaped                                     |
| C17 | Response shape (`/v1/systemone`)         | SystemOne response parsed as shaped                                      |
| C18 | Unsupported operations                   | Runtime rejects unsupported operations with an error, not a guess        |

---

## 2. Transport Semantic-Fidelity Rubric (frozen)

**Frozen at:** 2026-10-06T23:13:31-05:00 — the same repository HEAD timestamp as
§1; the rubric is frozen with the checklist and applied uniformly to every
candidate transport.

The rubric governs **primary transport selection**. Criteria are **ordered by
priority**; an earlier criterion outranks any later one. A transport that loses on
an earlier criterion cannot win on a later one.

| Rank | Criterion                         | Definition                                                                                                                    | Evidence requirement (per candidate)                                                         |
| ---- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| F1   | **Native decision semantics**     | Transport exposes Clef's _decision_ semantics (choice + probabilities + confidence) directly, without an invented translation | Captured response of the native decision call showing `type/choice/probabilities/confidence` |
| F2   | **Choice + probability fidelity** | The allowed choice and per-choice probabilities survive the transport unchanged                                               | A capture where `choice ∈ allowed set` and `probabilities` sum ≈ 1 over the allowed set      |
| F3   | **Confidence fidelity**           | Confidence is present, scalar, and bounded to `[0,1]`                                                                         | A capture with a numeric `confidence` in `[0,1]`                                             |
| F4   | **Message / role semantics**      | Roles and message ordering map without loss (system vs user vs assistant)                                                     | Captured request/response showing role handling                                              |
| F5   | **Streaming granularity**         | Streaming (if used) preserves decision-payload structure                                                                      | A streaming capture, or an explicit statement that streaming is unused                       |
| F6   | **Tool-call structure**           | Tool calls (if used) are structured, not scraped                                                                              | A structured tool-call capture, or an explicit statement that tool-calls are unused          |
| F7   | **Stop / usage semantics**        | Stop conditions and usage accounting are reported, not inferred                                                               | A capture containing `usage` and/or a stop reason                                            |
| F8   | **Error semantics**               | Failures are typed/structured and fail closed, never silently coerced to a valid decision                                     | A captured failure or an explicit statement of the error taxonomy                            |
| F9   | **Determinism of parsing**        | Output is structured JSON with no stdout scraping or sampling-sensitive parsing                                               | A statement of the parsing mode, backed by the response capture                              |

Scoring: **PASS / PARTIAL / FAIL / UNKNOWN** per criterion, with the capture id
cited. UNKNOWN is treated as a loss (a candidate cannot win on unproven behavior).

---

## 3. Runtime reachability and version discovery

All facts observed live on 2026-10-07 from this machine.

| Runtime | Endpoint                 | Version                    | Models                                                                                                                                                         | Capture                                              |
| ------- | ------------------------ | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| Ollama  | `http://localhost:11434` | `0.40.0`                   | `clef-flash:latest` — capabilities `decision`, `vision`; family `clef`; 9,531,164,032 params; 32 blocks; embedding 4096; context 262144; `mxfp8`; runner `mlx` | `ollama-version.json`, `ollama-show-clef-flash.json` |
| oMLX    | `http://localhost:8000`  | `0.7.0` (`omlx --version`) | `clef-4bit`, `clef-8bit` (engine `vlm`, config `qwen3_5`, not loaded), `Qwen3-4B-4bit` (loaded)                                                                | `omlx-models.json`, `omlx-models-status.json`        |

> **Superseded detail (2026-10-07, see §0.2):** a later live re-verification found
> `clef-4bit` is now downloaded and **loaded** (`omlx-models-status.json`,
> `loaded: true`). The historical "not loaded" observation above is retained
> unchanged for provenance.

---

## 4. Ollama capability classifications

### 4.1 Wire contract (live)

`clef-flash` on Ollama accepts the **same SystemOne request shape the Nimble
provider already uses** (`internal/providers/nimble/systemone.go`,
`systemOneRequest`): `{"model","state","questions":{id:{type,instructions,criteria[,choices]}}}`.

- choice: `criteria` is an object of candidate → description|null (2–26 keys).
- score: `criteria` is an array of 2–26 candidate descriptions.
- noul: `instructions` only.
- `state` must be non-empty; `questions` must have 1–64 entries.

Choice response (`choice-run1.json`):

```json
{"model":"clef-flash",
 "answers":{"risk":{"type":"choice","choice":"HIGH",
   "probabilities":{"LOW":0.02424653587413193,"MEDIUM":0.06590891786973317,"HIGH":0.9098445462561349},
   "confidence":0.676513609057984}},
 "usage":{"input_tokens":150,"output_tokens":0},
 "prompt_eval_cached_count":…}
```

Score response (`score.json`):

```json
{
  "q": {
    "type": "score",
    "score": 0.10430811755852966,
    "legend": { "0": "not safe", "1": "somewhat safe", "2": "very safe" },
    "probabilities": {
      "0": 0.9218425424821404,
      "1": 0.05200679747718923,
      "2": 0.026150660040670213
    },
    "confidence": 0.7050259489055778
  }
}
```

Noul response (`noul.json`): `{"q":{"type":"noul","noul":0.9453993942183476}}` —
**no probabilities and no confidence**.

### 4.2 Answer semantics recomputed from the captures

The probe recomputes these from each response's own probabilities and stores the
result under `invariants` in the capture:

| Property                                                                            | Result                      | Captures                                                |
| ----------------------------------------------------------------------------------- | --------------------------- | ------------------------------------------------------- |
| Probabilities sum to 1 over the allowed candidates                                  | holds (`1.0`)               | `choice-run*.json`, `score.json`, `multi-question.json` |
| `choice` is the argmax of `probabilities`                                           | holds                       | `choice-run*.json`, `multi-question.json`               |
| `confidence` = 1 − H(p) / ln(n) (normalized-entropy certainty, not max-probability) | holds to 1e-6               | `choice-run*.json`, `score.json`, `multi-question.json` |
| `score` = Σ i·p(i) — expected candidate **index** in `[0, n−1]`, not a 0–10 scale   | holds to 1e-6               | `score.json`                                            |
| Identical single-question requests → identical answers                              | bit-identical across 3 runs | `choice-run1..3.json`                                   |

Observed model-specific behavior (inputs to CLEF-003, not classifications):

- **Questions are not independent.** Adding a noul question shifts the choice
  distribution (`HIGH` 0.9098 → 0.9127; `multi-question.json`).
- **Small run-to-run drift.** The multi-question noul value moved by ≤1.5e-3
  between probe runs (0.8962 → 0.8948) while single-question choice stayed
  bit-identical; exact reproducibility is not guaranteed.
- **noul without `instructions` is accepted** (HTTP 200, `noul-no-instructions.json`)
  but answers an unspecified question. The adapter must always send instructions
  (the Nimble provider already synthesizes a fallback).
- Response carries an extra top-level `prompt_eval_cached_count` field.
- `usage.output_tokens` is always `0` (single forward pass, no generation).

### 4.3 Classification table

| #   | Capability                                  | Classification                                         | Evidence                                                                                                                                                                                                                                                                       |
| --- | ------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| C1  | Chat / completions                          | **UNSUPPORTED**                                        | `chat.json`: HTTP 400 `"clef-flash" does not support chat`; `completion` not in declared capabilities                                                                                                                                                                          |
| C2  | Streaming                                   | **UNSUPPORTED (N/A)**                                  | Decision path is a single non-generative forward pass (`output_tokens:0`); chat is rejected, so there is no stream to consume                                                                                                                                                  |
| C3  | Tool / function calling                     | **UNSUPPORTED**                                        | No chat surface (`chat.json`); `tools` not in declared capabilities                                                                                                                                                                                                            |
| C4  | Structured / JSON output (decision payload) | **SUPPORTED**                                          | `choice-run1.json`, `score.json`, `noul.json`                                                                                                                                                                                                                                  |
| C5  | Embeddings                                  | **UNSUPPORTED**                                        | `embed.json`: HTTP 400 on a 2-word input; `embedding` not in declared capabilities                                                                                                                                                                                             |
| C6  | Vision                                      | **UNKNOWN / REQUIRES EXPERIMENT**                      | `vision` declared (`ollama-show-clef-flash.json`); no image probe run. Not needed by the text-only decision contract                                                                                                                                                           |
| C7  | Context length                              | **SUPPORTED (declared)**                               | `clef.context_length: 262144` (`ollama-show-clef-flash.json`); declared, not exercised at length                                                                                                                                                                               |
| C8  | System-prompt handling                      | **UNSUPPORTED (N/A)**                                  | No role model; the decision API takes `state` + per-question `instructions` only                                                                                                                                                                                               |
| C9  | Choice decision                             | **SUPPORTED**                                          | `choice-run1..3.json`                                                                                                                                                                                                                                                          |
| C10 | Per-choice probabilities                    | **SUPPORTED**                                          | `choice-run1..3.json` (sum 1.0 over allowed candidates)                                                                                                                                                                                                                        |
| C11 | Confidence in `[0,1]`                       | **SUPPORTED (choice, score)** / **UNSUPPORTED (noul)** | `choice-run1.json` 0.6765, `score.json` 0.7050; `noul.json` carries none                                                                                                                                                                                                       |
| C12 | Score operation                             | **SUPPORTED**                                          | `score.json` (array criteria only; `legend` input rejected, `reject-score-legend.json`)                                                                                                                                                                                        |
| C13 | Boolean / noul operation                    | **SUPPORTED**                                          | `noul.json`, `multi-question.json`                                                                                                                                                                                                                                             |
| C14 | Token usage                                 | **SUPPORTED**                                          | `usage.input_tokens` / `output_tokens` in every 200 response                                                                                                                                                                                                                   |
| C15 | Null / indeterminate handling               | **UNSUPPORTED (native)**                               | No abstain/null answer type exists; a choice is always returned. 0 or 1 candidate → HTTP 400 (`reject-empty-candidates.json`, `reject-single-candidate.json`); empty state → HTTP 400 (`reject-empty-state.json`). Indeterminacy must be derived from `confidence` in CLEF-003 |
| C16 | Request shape (`/v1/systemone`)             | **SUPPORTED**                                          | `choice-run1.json`, `score.json`, `noul.json`, `multi-question.json`; non-conforming shapes rejected (`reject-transcribed-choice-shape.json`, `reject-score-legend.json`)                                                                                                      |
| C17 | Response shape (`/v1/systemone`)            | **SUPPORTED**                                          | `answers[id]{type, choice                                                                                                                                                                                                                                                      | score | noul, probabilities, legend, confidence}`+`usage` |
| C18 | Unsupported operations                      | **SUPPORTED (fails closed)**                           | Structured `{"error": …}` with HTTP 400/404, never a coerced answer: `reject-unknown-type.json`, `reject-unknown-model.json` (404), `reject-non-decision-model.json`, `reject-malformed-json.json`                                                                             |

---

## 5. oMLX compatibility classification

**Explicit classification: oMLX is `UNSUPPORTED` as a Clef decision transport —
both as a native interface and as a faithful translation target.**

### 5.1 Evidence

- `POST /v1/systemone` on oMLX → **HTTP 404** (`omlx-systemone-route.json`).
- oMLX registers `clef-4bit`/`clef-8bit` with its generic `vlm` engine and
  `config_model_type: qwen3_5` (`omlx-models-status.json`) — i.e. it serves the
  Qwen3.5 backbone.
- The Clef decision probabilities come from a separate **joint schema head**
  (`joint_head.safetensors`, `joint_head_config.json`; `clef-4bit-local-metadata.json`)
  that is executed only by the vendor loader `clef_mlx.py`. The model card
  states: _"Not a chat model. Text generation tools (`mlx_lm.generate`,
  `mlx_vlm.generate`, LM Studio, Ollama) load the backbone but give meaningless
  output. Only `clef_mlx.py` runs the decision head."_
  (`~/.omlx/omlx/models/mlx-community/clef-4bit/README.md`, "Limitations").

### 5.2 Required distinctions

- **Native decision interface (oMLX):** **ABSENT** — live 404.
- **Faithful translation (oMLX):** **UNSUPPORTED.** The decision semantics live in
  the joint head, which oMLX does not execute. Reconstructing choices from
  backbone `top_logprobs` would not be Clef's decision output.
- **Generic generation approximation:** **REJECTED as non-equivalent.** Vendor-
  documented as meaningless for this checkpoint. Not live-probed: loading the
  17 GB checkpoint to confirm a documented negative was judged unnecessary.

### 5.3 Per-capability oMLX classification (same frozen checklist)

| #       | Capability                                                         | oMLX classification                                  | Basis                                                                                                        |
| ------- | ------------------------------------------------------------------ | ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| C1      | Chat / completions                                                 | **UNKNOWN** (route exists) — **not a decision path** | Generic route; vendor: meaningless output for Clef. Not probed                                               |
| C2–C8   | Streaming, tools, JSON, embeddings, vision, context, system prompt | **UNKNOWN / REQUIRES EXPERIMENT**                    | Not probed; irrelevant once C9–C17 are unsupported. C7 declared `max_model_len: 262144` (`omlx-models.json`) |
| C9–C13  | Choice, probabilities, confidence, score, noul                     | **UNSUPPORTED**                                      | No decision route (404); decision head not executed by oMLX                                                  |
| C14     | Token usage                                                        | **UNKNOWN**                                          | Not probed                                                                                                   |
| C15     | Null / indeterminate                                               | **UNSUPPORTED**                                      | No decision surface                                                                                          |
| C16–C17 | `/v1/systemone` request/response shape                             | **UNSUPPORTED**                                      | `omlx-systemone-route.json` (404)                                                                            |
| C18     | Unsupported operations                                             | **SUPPORTED (fails closed)**                         | 404 `{"detail":"Not Found"}` on the decision route                                                           |

### 5.4 Separate MLX candidate surfaced by this investigation

The clef-4bit bundle ships `clef_mlx.py serve`, documented as a local
`POST /v1/systemone` server with the same request/response shape (plus
`usage.latency_ms`). It is **not oMLX**. It was **not probed**: it needs a pinned
`mlx-vlm` 0.7.x install and executes vendor custom code. Classification:
**UNKNOWN / REQUIRES EXPERIMENT** (owner: CLEF-006, conditional). Even if it
works it serves a **different model** from `clef-flash` (§6).

---

## 6. Base-Model Equivalence (`clef-flash` vs `mlx-community/clef-4bit`)

**Result: ESTABLISHED — NOT the same base model.**

| Property     | `clef-flash` (Ollama)          | `mlx-community/clef-4bit`            |
| ------------ | ------------------------------ | ------------------------------------ |
| Parameters   | 9,531,164,032 (9.5B)           | 27B base (model-card variant table)  |
| Layers       | 32 (`clef.block_count`)        | 64 (`text_config.num_hidden_layers`) |
| Hidden size  | 4096 (`clef.embedding_length`) | 5120 (`text_config.hidden_size`)     |
| Base         | Clef "flash" family, 9B        | `base_model: Cloudflare/clef`        |
| Quantization | `mxfp8`                        | 4-bit affine, group 64               |
| Capture      | `ollama-show-clef-flash.json`  | `clef-4bit-local-metadata.json`      |

The clef-4bit model card lists `clef-flash-4bit`/`clef-flash-8bit` (9B) as the
MLX variants of the flash model; neither is present locally. Weight digests were
not compared — the architecture mismatch already rules out equivalence.

---

## 7. Primary Transport Decision

### 7.1 Candidates scored against the frozen rubric (§2)

| Criterion                        | **A. Ollama `/v1/systemone`** (`clef-flash`)                                          | **B. oMLX OpenAI-compat + translation** (`clef-4bit`) | **C. `mlx_lm`/`mlx_vlm` generate** | **D. `clef_mlx.py serve`** (`clef-4bit`) |
| -------------------------------- | ------------------------------------------------------------------------------------- | ----------------------------------------------------- | ---------------------------------- | ---------------------------------------- |
| F1 Native decision semantics     | **PASS** (`choice-run1.json`)                                                         | **FAIL** (404; head not executed)                     | **FAIL** (vendor: meaningless)     | **UNKNOWN** (documented, not probed)     |
| F2 Choice + probability fidelity | **PASS** — argmax choice, sum 1.0                                                     | **FAIL**                                              | **FAIL**                           | **UNKNOWN**                              |
| F3 Confidence fidelity           | **PASS** — `[0,1]`, = 1 − H/ln n (choice, score); absent on noul                      | **FAIL**                                              | **FAIL**                           | **UNKNOWN**                              |
| F4 Message / role semantics      | **PASS (N/A roles)** — `state` + `instructions` carried as sent                       | UNKNOWN                                               | UNKNOWN                            | UNKNOWN                                  |
| F5 Streaming granularity         | **N/A** — non-streaming                                                               | UNKNOWN                                               | UNKNOWN                            | UNKNOWN                                  |
| F6 Tool-call structure           | **N/A** — unused                                                                      | UNKNOWN                                               | UNKNOWN                            | UNKNOWN                                  |
| F7 Stop / usage semantics        | **PASS** — `usage` on every 200                                                       | UNKNOWN                                               | UNKNOWN                            | UNKNOWN                                  |
| F8 Error semantics               | **PASS** — structured 400/404 on 9 failure probes (`reject-*.json`)                   | PARTIAL (404 only)                                    | UNKNOWN                            | UNKNOWN (documented 400/413)             |
| F9 Determinism of parsing        | **PASS** — structured JSON; identical requests bit-identical (small drift noted §4.2) | **FAIL** — logprob reconstruction                     | **FAIL** — text scraping           | UNKNOWN                                  |

UNKNOWN counts as a loss (§2).

### 7.2 Decision

> **Primary transport: A — Ollama's native `POST /v1/systemone` serving `clef-flash`.**

A is the only candidate that passes F1–F3 with live captures. B and C fail F1
outright because the decision head is not executed. D may be native but is
unproven here and serves a different (27B) model. Reuse of the Nimble wire shape
is a consequence of A's live-verified contract, not the reason for choosing it.

---

## 8. Failure / error semantics (live)

| Condition                         | Live behavior                                    | Capture                                                        |
| --------------------------------- | ------------------------------------------------ | -------------------------------------------------------------- |
| Unknown model                     | HTTP 404 `model "…" not found`                   | `reject-unknown-model.json`                                    |
| Model without decision capability | HTTP 400 `… does not support decision`           | `reject-non-decision-model.json`                               |
| Malformed JSON                    | HTTP 400 `unexpected end of JSON input`          | `reject-malformed-json.json`                                   |
| Missing/misnamed `questions`      | HTTP 400 `questions must contain 1–64 fields`    | `reject-transcribed-choice-shape.json`                         |
| Unknown question type             | HTTP 400 `type must be choice, noul, or score`   | `reject-unknown-type.json`                                     |
| <2 choice candidates              | HTTP 400 `criteria must contain 2–26 candidates` | `reject-single-candidate.json`, `reject-empty-candidates.json` |
| Score criteria not an array       | HTTP 400 `score criteria must be an array`       | `reject-score-legend.json`                                     |
| Empty state                       | HTTP 400 `state must not be empty`               | `reject-empty-state.json`                                      |

Every failure is a structured `{"error": string}` body; none returned a coerced
answer. Runtime-down, timeout, and cancellation were not probed; they are
adapter-owned transport errors (Nimble's `transport.go` already classifies them)
and are specified in CLEF-003.

---

## 9. Acceptance Criteria Traceability

| Acceptance criterion                                          | Status                           | Section                                 | Captures                                                       |
| ------------------------------------------------------------- | -------------------------------- | --------------------------------------- | -------------------------------------------------------------- |
| Ollama Clef capabilities are individually classified          | MET                              | §4.3 (C1–C18)                           | `captures/live/*`                                              |
| Every SUPPORTED classification has captured evidence          | MET                              | §4.3, §5.3                              | each SUPPORTED row cites a live capture                        |
| oMLX compatibility is explicitly classified                   | MET                              | §5 (UNSUPPORTED native and translation) | `omlx-*.json`, `clef-4bit-local-metadata.json`                 |
| No untested behavior is presented as supported                | MET                              | §4.3, §5.3 (unprobed rows UNKNOWN)      | —                                                              |
| `clef-flash` vs `clef-4bit` base model established or UNKNOWN | MET (established: **different**) | §6                                      | `ollama-show-clef-flash.json`, `clef-4bit-local-metadata.json` |
| Exactly one primary transport is selected                     | MET                              | §7.2                                    | —                                                              |
| Selected on semantic fidelity and evidence                    | MET                              | §7.1                                    | `choice-run*.json`, `reject-*.json`                            |
| No production provider implementation is created              | MET                              | §10                                     | —                                                              |

Stop condition check: Ollama `clef-flash` produces bounded choice evidence with an
allowed choice and usable confidence (`choice-run1.json`). The stop condition is
**not** triggered.

### 9.1 Acceptance evidence map (verbatim criteria to evidence)

CLEF-002 is an investigation whose deliverable is this report; it produces no
production code. Each compiled acceptance criterion maps to concrete, inspectable
evidence. Criteria 5–6 and 8–9 are continuation fragments of two logical criteria in
the source plan and are quoted here verbatim as compiled.

```text
1. Ollama Clef capabilities are individually classified.
2. Every SUPPORTED classification has captured evidence.
3. oMLX compatibility is explicitly classified.
4. No untested behavior is presented as supported.
5. Whether `clef-flash` and `mlx-community/clef-4bit` represent the same base
6. model is either established or recorded UNKNOWN.
7. Exactly one primary transport is selected.
8. The primary transport is selected based on semantic fidelity and evidence,
9. not convenience.
10. No production provider implementation is created.
```

| #   | Evidence file(s)                                                                                                           | Section    |
| --- | -------------------------------------------------------------------------------------------------------------------------- | ---------- |
| 1   | `captures/live/*.json`                                                                                                     | §4.3       |
| 2   | `captures/live/choice-run1.json`, `choice-run2.json`, `choice-run3.json`, `score.json`, `noul.json`, `multi-question.json` | §4.3       |
| 3   | `captures/live/omlx-systemone-route.json`, `omlx-models.json`, `omlx-models-status.json`                                   | §5         |
| 4   | `captures/live/*.json` (UNKNOWN rows make no SUPPORTED claim)                                                              | §4.3, §5.3 |
| 5   | `captures/live/ollama-show-clef-flash.json`, `clef-4bit-local-metadata.json`                                               | §6         |
| 6   | `captures/live/ollama-show-clef-flash.json`, `clef-4bit-local-metadata.json`                                               | §6         |
| 7   | this report                                                                                                                | §7.2       |
| 8   | `captures/live/choice-run1.json`, `reject-*.json`                                                                          | §7.1       |
| 9   | `captures/live/choice-run1.json`, `reject-*.json`                                                                          | §7.1       |
| 10  | repository tree (`git status --short`)                                                                                     | §10        |

**How to re-verify (already-satisfied path).** The live captures are produced by the
operator-run probe:

```bash
python3 tools/clef/probe_systemone.py
```

The governed IMPLEMENT agent's command policy permits only `go build|test|vet|list|fmt`
and read-only `git`; it therefore **inspects the captured files** above — reading
**every** file it cites — and runs the configured validations (`go build ./...`,
`go vet ./...`, `go test ./...`) rather than re-running the probe itself.

---

## 10. Change discipline and scope

- Added `tools/clef/probe_systemone.py` (stdlib-only, non-production probe) and
  `docs/reports/clef-provider/captures/live/*.json`.
- Removed the four transcribed captures superseded in §0.1.
- Rewrote this report. §1 and §2 are carried over verbatim.
- No file under `internal/`, `cmd/`, or `decision/` was created or modified. No Clef
  provider, transport, registration, or configuration exists.
- Julia artifacts and conclusions untouched. `agentic-sop` not modified.

## 11. Open items register

| Item                                                         | Status                        | Owner task                                      |
| ------------------------------------------------------------ | ----------------------------- | ----------------------------------------------- |
| Vision (C6) on `clef-flash`                                  | UNKNOWN                       | CLEF-006 (only if image input is ever in scope) |
| Long-context behavior near 262144 tokens                     | declared only                 | CLEF-006                                        |
| noul carries no confidence — mapping for BOOLEAN             | mapping decision              | CLEF-003                                        |
| Indeterminacy derived from normalized-entropy confidence     | mapping decision              | CLEF-003                                        |
| Score is expected index `[0, n−1]` — normalization for SCORE | mapping decision              | CLEF-003                                        |
| Cross-question coupling and small run-to-run drift           | observed                      | CLEF-008                                        |
| Runtime-down / timeout / cancellation error mapping          | adapter-owned                 | CLEF-003 / CLEF-005                             |
| `clef_mlx.py serve` on MLX (candidate D)                     | UNKNOWN / REQUIRES EXPERIMENT | CLEF-006 (conditional)                          |
| Confidence calibration                                       | UNKNOWN                       | CLEF-008                                        |
