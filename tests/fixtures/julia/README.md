# Julia (ONNX) fixtures

Illustrative fixtures for the `internal/providers/julia` adapter. They document
the translation contract between the provider-neutral `decision` package and the
Julia/ONNX model. **No live runtime ships here**: the default test suite is fully
offline and never executes Julia or ONNX.

## Runner contract

The Julia adapter isolates concrete ONNX execution behind the `Runner` seam. A
`CommandRunner` marshals `Inputs` to the inference command's stdin as JSON and
decodes `Outputs` from its stdout as JSON.

### Inputs (stdin)

```json
{
  "state": "deploying a schema migration during business hours",
  "questions": [
    {"id": "risk", "type": "choice", "text": "Assess deployment risk.", "labels": ["LOW", "MEDIUM", "HIGH"]},
    {"id": "execution_model", "type": "choice", "text": "Preferred execution model", "labels": ["SMALL", "MEDIUM", "LARGE"]},
    {"id": "approval_required", "type": "boolean", "text": "Is human approval required?"}
  ]
}
```

### Outputs (stdout)

See [risk-evaluation.outputs.json](risk-evaluation.outputs.json).

## Normalization rules

`NormalizeOutputs` requires an answer for every requested question. Per type:

- `choice` — `choice` must be non-empty and one of the question's allowed
  labels; every probability key must be an allowed label; every probability value
  must be in `[0, 1]`. Probabilities are not required to sum to 1.0.
- `boolean` — `probability` must be present and in `[0, 1]`.
- `score` — `score` must be present.
- `confidence`, when present, must be in `[0, 1]`.

Every rejection wraps `decision.ErrMalformedResponse`. A malformed request wraps
`decision.ErrInvalidRequest`. The result advertises `provider` = `"julia"` and
copies the model identifier verbatim.

This fixture is documentation only; it is not executed by the default suite. The
opt-in live test is guarded by `JULIA_INTEGRATION_TEST=1` and a configured
`JULIA_INFERENCE_CMD`.
