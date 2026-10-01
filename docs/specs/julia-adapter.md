# Specification: Julia (ONNX) decision adapter

**Status:** normative for `internal/providers/julia`.

This spec defines the Julia decision provider. It mirrors
[nimble-systemone.md](nimble-systemone.md). Julia adapts the frozen `decision`
contract onto the Julia-1-ONNX supplied-options model; the adapter only
translates and normalizes and never defines or enforces SOP policy.

## Request path

```text
decision.DecisionRequest
        ↓
Julia adapter (BuildInputs)
        ↓
Julia runtime request (JSON on the helper's stdin)
        ↓
repo-owned helper (tools/julia/infer.py)
        ↓
tokenizer → ONNX tensors → Julia-1-ONNX → logits → decode
        ↓
Outputs (JSON on the helper's stdout)
        ↓
Julia adapter (NormalizeOutputs)
        ↓
decision.DecisionResult
```

The concrete ONNX execution is isolated behind the `Runner` interface: the
shipped `CommandRunner` invokes the repository-owned helper (or an operator
override), and a future in-process ONNX binding can replace it without touching
the adapter, translation, CLI, or public contract.

## Julia native semantics (qtype)

Julia is a **supplied-options** model. Each native decision carries an ordered
option list plus a question type (`qtype`). The mapping is:

| `decision.QuestionType` | Julia qtype | Options                                   | Decoding                                            |
| ----------------------- | ----------- | ----------------------------------------- | --------------------------------------------------- |
| `QuestionChoice`        | `0` choice  | the question's `Choices`                  | `argmax` → selected label; per-option probabilities |
| `QuestionScore`         | `1` score   | the question's `Choices` (ordered levels) | **expected zero-based option index** `Σ i·pᵢ`       |
| `QuestionBoolean`       | `2` noul    | fixed `["false","true"]`                  | `P(true)` = probability of index 1                  |

These are Julia-specific and live only in `internal/providers/julia` and the
helper. Nothing here (no `qtype`, no `noul`, no `marker_pos`) appears in the
public `decision` package.

### Option limits

- CHOICE: **2–20** ordered options.
- SCORE: **2–20** ordered options.
- BOOLEAN: exactly two internally generated options.

The provider-neutral `decision.Question` has no first-class score-scale field, so
SCORE levels ride along in `Choices` (existing design debt; see
[`../plans/BACKLOG.md`](../plans/BACKLOG.md)). A CHOICE or SCORE question with
fewer than two or more than twenty options has no valid Julia decision and is
rejected with `decision.ErrInvalidRequest` by `BuildInputs` (and therefore by
`Decide`) before any inference.

### Confidence

Julia's documented runtime exposes no distinct confidence value, so the adapter
does **not** fabricate one: `decision.Answer.Confidence` is left unset. The
maximum softmax probability is **not** equated with confidence.

### BOOLEAN criteria (limitation)

The provider-neutral BOOLEAN question carries no distinct descriptions for its
false/true options, so the adapter cannot pass separate false/true criteria to
Julia: noul options are always `["false", "true"]`. Preserving distinct
descriptions would require a change to the public contract and is deliberately
deferred (see [`../plans/BACKLOG.md`](../plans/BACKLOG.md)).

## Runner contract

```go
type Runner interface {
    Name() string
    Available(ctx context.Context) bool
    Run(ctx context.Context, in Inputs) (Outputs, error)
}
```

- `Available` reports whether inference can currently be served. It MUST NOT
  perform an inference request.
- `Run` evaluates a Julia runtime request (`Inputs`) and returns raw `Outputs`.

`Inputs`/`Outputs` are **Julia adapter/runtime structures**, not the
provider-neutral contract. `Inputs` carries Julia's native vocabulary:

```go
type Inputs struct {
    State     string          `json:"state"`
    Questions []QuestionInput `json:"questions"`
}

type QuestionInput struct {
    ID      string                `json:"id"`
    Type    decision.QuestionType `json:"type"`    // diagnostics only
    QType   int                   `json:"qtype"`   // 0 choice, 1 score, 2 noul
    Text    string                `json:"text"`
    Options []string              `json:"options,omitempty"`
}
```

`Text` uses the same fallback chain as Nimble: `Instructions` → `Criteria` →
`"Decide <id>."`.

```go
type Outputs struct {
    Model   string                    `json:"model,omitempty"`
    Answers map[string]QuestionOutput `json:"answers"`
}

type QuestionOutput struct {
    Choice        string             `json:"choice,omitempty"`
    Probabilities map[string]float64 `json:"probabilities,omitempty"`
    Probability   *float64           `json:"probability,omitempty"`
    Score         *float64           `json:"score,omitempty"`
    Confidence    *float64           `json:"confidence,omitempty"`
}
```

No Nimble, SystemOne, Ollama, or ONNX vocabulary appears in the public contract.

### CommandRunner and the repository helper

`CommandRunner{Command, ModelPath, HelperPath, Timeout}` is the stdlib `os/exec`
implementation:

- marshals `Inputs` to the command's stdin as JSON;
- decodes `Outputs` from stdout as JSON;
- bounds the call with a context timeout (`JULIA_TIMEOUT`, default 30s);
- captures stderr for diagnostics;
- forwards `JULIA_MODEL_PATH` to the child environment;
- classifies failures (see [Error mapping](#error-mapping)).

The **default** command is the repository-owned helper
`tools/julia/infer.py`, resolved relative to the working directory or one of its
ancestors and run with `JULIA_PYTHON`. `JULIA_INFERENCE_CMD` overrides it.

`Available` is a cheap, side-effect-free prerequisite check: the command program
must resolve on `PATH`; for the built-in helper the helper script **and** a real
`JULIA_MODEL_PATH` must exist. For an operator override (`JULIA_INFERENCE_CMD`)
only the program is checked.

## ONNX graph contract

The helper executes Julia-1-ONNX with these tensors:

```text
input_ids       int64  [batch, tokens]
attention_mask  int64  [batch, tokens]
marker_pos      int64  [batch, options]
marker_mask     bool   [batch, options]
qtype           int64  [batch]
        ↓
logits          float  [batch, options]
```

The helper validates raw logits (expected length, finiteness — no `NaN`/`±Inf`)
before applying a numerically stable softmax (`exp(x − max)`). It then decodes by
qtype (choice `argmax`; score expected index; boolean `P(true)`) and emits
`Outputs`. Confidence is omitted.

### Encoder verification status

The tensor contract, qtype mapping, option limits, and decoding rules are
implemented from the documented Julia-1-ONNX contract. The **encoding** of
state/question/options into `marker_pos`/`marker_mask` is a best-effort,
isolated implementation (`Runtime.encode` + `find_marker_positions` in
`tools/julia/infer.py`) that **has not been re-verified against the upstream
Julia-1-ONNX export** in this environment. It is the single place to reconcile
with upstream before trusting live output.

## Configuration

| Environment variable  | Meaning                                                                              | Default   |
| --------------------- | ------------------------------------------------------------------------------------ | --------- |
| `JULIA_MODEL`         | model identifier reported when the runner reports none                               | `julia`   |
| `JULIA_MODEL_PATH`    | path to the ONNX model (required by the helper)                                      | unset     |
| `JULIA_PYTHON`        | interpreter for the default helper                                                   | `python3` |
| `JULIA_INFERENCE_CMD` | external inference command overriding the default helper, whitespace-split into argv | unset     |
| `JULIA_TIMEOUT`       | inference timeout (Go duration syntax)                                               | `30s`     |

The helper additionally reads `JULIA_TOKENIZER_PATH` (default `tokenizer.json`
beside the model), `JULIA_MODEL_ID` (default `julia`), and `JULIA_MAX_TOKENS`
(default `8192`), which are inherited from the parent environment.

`JULIA_INFERENCE_CMD` is whitespace-split; shell quoting rules are not applied,
so an argument containing spaces is unsupported.

## Error mapping

`Decide` validates the request, builds `Inputs` (rejecting Julia option-count
violations), runs the `Runner`, and normalizes `Outputs`, wrapping failures in
`*decision.ProviderError` attributed to `"julia"`:

| Condition                                                                                                              | ErrorKind / sentinel                             |
| ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| request fails `Validate`, or a CHOICE/SCORE question has an invalid option count                                       | `KindInvalidRequest` / `ErrInvalidRequest`       |
| helper/model/runtime unavailable (unset or missing model, missing dependencies, unclassifiable command failure)        | `KindUnavailable` / `ErrUnavailable`             |
| undecodable or non-normalizable output (invalid JSON, bad logits shape, non-finite probabilities, incompatible answer) | `KindMalformedResponse` / `ErrMalformedResponse` |
| unexpected inference failure (classified `failure` by the helper)                                                      | `KindProviderFailure` / `ErrProviderFailure`     |

The helper reports failures on stdout as `{"error":{"kind":…,"message":…}}`, which
`CommandRunner` maps onto the matching runner sentinel so `errors.Is` still holds.

## Normalization rules

`NormalizeOutputs(out, req)` requires an answer for every requested question. Per
question type:

- **CHOICE** — `Choice` MUST be non-empty and one of the question's allowed
  choices. Every probability KEY MUST be an allowed choice. Every probability
  value MUST be in `[0, 1]`. Probabilities MUST NOT be required to sum to 1.0.
- **BOOLEAN** — `Probability` MUST be present and in `[0, 1]`.
- **SCORE** — `Score` MUST be present.

`Confidence`, when present, MUST be in `[0, 1]`.

Every rejection wraps `decision.ErrMalformedResponse`, including the answer type
mismatch case. A request that fails `req.Validate()` wraps
`decision.ErrInvalidRequest`.

The normalized result sets `Provider` to `ProviderName = "julia"` and starts from
`Outputs.Model`; the runner-reported model keeps precedence, and the configured
`JULIA_MODEL` is used only when the runner reports none.

## Testing

The default suite is fully offline: it exercises the adapter through `RunnerFunc`
fakes and hermetic `sh -c` commands and requires no ONNX, Julia, Ollama, Python,
or network. The helper's pure functions are unit-tested with the standard
library alone (`python3 tools/julia/test_infer.py`).

The live test is opt-in and only runs when `JULIA_INTEGRATION_TEST=1`; it SKIPS
(never fails the suite) when the runtime is not configured, otherwise it drives
Go → helper → ONNX Runtime → Julia-1-ONNX and asserts the contract, not exact
judgments.
