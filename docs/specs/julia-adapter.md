# Specification: Julia (ONNX) decision adapter

**Status:** normative for `internal/providers/julia`.

This spec defines the Julia decision provider. It mirrors
[nimble-systemone.md](nimble-systemone.md). Julia adapts the frozen
`decision` contract onto a Julia/ONNX decision model; the adapter only
translates and normalizes and never defines or enforces SOP policy.

## Request path

```
decision.Provider -> internal/providers/julia -> Runner -> ONNX inference -> Julia
```

The concrete ONNX execution is isolated behind the `Runner` seam so the adapter
is testable without any ONNX runtime, and an in-process ONNX binding can replace
the external command runner without touching the adapter, the translation layer,
the CLI, or the public contract.

## Runner contract

The package defines:

```go
type Runner interface {
    Name() string
    Available(ctx context.Context) bool
    Run(ctx context.Context, in Inputs) (Outputs, error)
}
```

- `Available` reports whether inference can currently be served. It MUST NOT
  perform an inference request.
- `Run` evaluates `Inputs` and returns raw `Outputs`.

Runner implementations MUST classify transport-like failures so the adapter can
map them onto a `decision.ErrorKind`:

- unreachable or unconfigured runner (for example an unset/failed/timed-out
  inference command) → the adapter returns `decision.ErrUnavailable`.
- output the runner could not decode → the adapter returns
  `decision.ErrMalformedResponse`.
- any other failure → the adapter returns `decision.ErrProviderFailure`.

### Inputs (JSON on the inference command's stdin)

```go
type Inputs struct {
    State     string          `json:"state"`
    Questions []QuestionInput `json:"questions"`
}

type QuestionInput struct {
    ID     string                `json:"id"`
    Type   decision.QuestionType `json:"type"`
    Text   string                `json:"text"`
    Labels []string              `json:"labels,omitempty"`
}
```

`Text` uses the same fallback chain as Nimble: `Instructions` → `Criteria` →
`"Decide <id>."`. `Labels` carries the question's allowed `Choices` (empty for
non-choice questions).

### Outputs (JSON on the inference command's stdout)

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

No Nimble, SystemOne, Ollama, or `noul` terminology appears in these types.

### CommandRunner

`CommandRunner{Command []string, ModelPath string, Timeout time.Duration}` is the
stdlib `os/exec` implementation:

- marshals `Inputs` to the command's stdin as JSON;
- decodes `Outputs` from stdout as JSON;
- bounds the call with a context timeout (`JULIA_TIMEOUT`, default 30s);
- captures stderr for diagnostics;
- forwards `JULIA_MODEL_PATH` to the child environment when set;
- `Available` is true iff a command is configured.

## Configuration

| Environment variable | Meaning | Default |
| --- | --- | --- |
| `JULIA_MODEL` | model identifier used as the reported model when the runner does not report one | `julia` |
| `JULIA_MODEL_PATH` | path to the ONNX model (forwarded as `JULIA_MODEL_PATH`) | empty |
| `JULIA_INFERENCE_CMD` | external inference command, whitespace-split into argv (empty ⇒ unavailable) | empty |
| `JULIA_TIMEOUT` | inference timeout (Go duration syntax) | `30s` |

`JULIA_INFERENCE_CMD` is whitespace-split; shell quoting rules are not applied,
so an argument containing spaces is unsupported.

## Normalization rules

`NormalizeOutputs(out, req)` requires an answer for every requested question. Per
question type:

- **CHOICE** — `Choice` MUST be non-empty and one of the question's allowed
  choices. Every probability KEY MUST be an allowed choice. Every probability
  value MUST be in `[0, 1]`. Probabilities MUST NOT be required to sum to 1.0.
- **BOOLEAN** — `Probability` MUST be present and in `[0, 1]`.
- **SCORE** — `Score` MUST be present.

`Confidence`, when present, MUST be in `[0, 1]`.

Every rejection wraps `decision.ErrMalformedResponse`, including the answer
type mismatch case (a CHOICE question whose answer has no `Choice` but has a
`Probability`, and any unknown/incompatible answer shape). A request that fails
`req.Validate()` wraps `decision.ErrInvalidRequest`.

The normalized result sets `Provider` to the exported constant
`ProviderName = "julia"` and starts from `Outputs.Model`; the result is then
validated via `result.Validate()`. After validation, the runner-reported model
keeps precedence: when `Outputs.Model` is empty, the configured `JULIA_MODEL` is
reported instead.

## Error mapping

`Decide` validates the request, builds `Inputs`, runs the `Runner`, and
normalizes `Outputs`, wrapping failures in `*decision.ProviderError` attributed
to `"julia"`:

| Condition | ErrorKind / sentinel |
| --- | --- |
| request fails `Validate` | `KindInvalidRequest` / `ErrInvalidRequest` |
| runner unreachable/unconfigured | `KindUnavailable` / `ErrUnavailable` |
| undecodable or non-normalizable output | `KindMalformedResponse` / `ErrMalformedResponse` |
| any other runner failure | `KindProviderFailure` / `ErrProviderFailure` |

## Testing

The default suite is fully offline: it exercises the adapter through `RunnerFunc`
fakes and hermetic `sh -c` commands and requires no ONNX, Julia, Ollama, or
network. A live test is opt-in and only runs when `JULIA_INTEGRATION_TEST=1` and
`JULIA_INFERENCE_CMD` are both set; it skips otherwise.
