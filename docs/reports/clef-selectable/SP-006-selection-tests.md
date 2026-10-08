# SP-006 — Selection and Default-Preservation Tests

## Scope

SP-006 adds focused, deterministic, model-free tests that pin the
**selection** surface of the sop-decision-adapter CLI: explicit Clef
selection, **default preservation**, the disabled-Clef accidental-selection
guard, empty/absent/unknown selection, selection not enabling Clef and not
affecting other providers, and **fail-closed** behavior for each failure path.

No production selection code, provider default, or agentic-sop policy is
modified. This is consistent with SP-002 (selectability contract SATISFIED) and
SP-005 (selection implementation, no-change outcome).

## Supported surface

Selection is asserted through the single supported surface only:

- the in-process `newProvider(name, baseURL, model, timeout)` seam in
  `cmd/sop-decision-adapter/main.go`;
- the `-provider` flag of the `decide` subcommand (default `"nimble"`, as
  documented in `usage()`).

No second selection mechanism is introduced or exercised. Clef **enablement**
is a separate gate controlled by `CLEF_ENABLED` / `clef.Config.Enable`, and
`WithDefaults` never enables it.

## Requirement-to-test matrix

All tests live in `cmd/sop-decision-adapter/clef_selection_test.go`.

| Required behavior | Test | Mechanism (deterministic, model-free) |
| --- | --- | --- |
| Explicit selection of Clef works through the supported surface | `TestNewProviderClefOffByDefault`, `TestNewProviderClefMixedCase`, `TestSelectingClefDoesNotEnableItNorAffectOthers` | `newProvider("clef", ...)` / `newProvider("  Clef ", ...)`; assert `Name() == clef.ProviderName` |
| Default provider unchanged when nothing is selected | `TestDefaultProviderUnchangedWhenNothingSelected`, `TestFlagProviderDefaultIsNimble` | `newProvider("", ...)` and `newProvider("   ", ...)` resolve to nimble even with `CLEF_ENABLED=1`; `usage()` documents `-provider` default `"nimble"` |
| Disabled Clef not accidentally selectable | `TestDisabledClefNotAccidentallySelected` | `t.Setenv("CLEF_ENABLED", ""/"0"/"false"/"no"/"off"/"nonsense")`; assert `Available()==false` and `Decide` matches `clef.ErrDisabled` and `decision.ErrUnavailable` |
| Absence/empty/unknown selection selects nothing (fails closed) | `TestDefaultProviderUnchangedWhenNothingSelected`, `TestNewProviderUnknownFailsClosed` | empty/blank -> nimble; unknown token -> nil provider + `unsupported provider` error |
| Selecting Clef does not enable it and does not affect other providers | `TestSelectingClefDoesNotEnableItNorAffectOthers`, `TestNewProviderNimbleAndJuliaUnchanged`, `TestAvailableProvidersIncludesClef` | constructing Clef with `CLEF_ENABLED` unset leaves it `Available()==false`; nimble/julia selection unchanged; `availableProviders` still lists nimble, julia, clef |
| Fail-closed: unavailable | `TestSelectionFailClosedPaths/unavailable` | `httptest` server closed immediately; connection refused -> `decision.KindUnavailable` |
| Fail-closed: timeout | `TestSelectionFailClosedPaths/timeout` | `httptest` handler sleeping well past the client timeout; client timeout fires -> `decision.KindUnavailable` |
| Fail-closed: malformed | `TestSelectionFailClosedPaths/malformed` | `httptest` server returning `{not json` -> `decision.KindMalformedResponse` |
| Fail-closed: invalid request/choice | `TestSelectionFailClosedPaths/invalid-request`, `/invalid-choice` | request with no questions / choice question with no choices -> `decision.KindInvalidRequest` |
| Fail-closed: unsupported operation | `TestSelectionFailClosedPaths/unsupported-operation` | SCORE question the adapter cannot express -> `clef.ErrUnsupported` + `decision.KindInvalidRequest` |
| Fail-closed: provider failure | `TestSelectionFailClosedPaths/provider-failure` | `httptest` server returning HTTP 500 -> `decision.KindProviderFailure` |

## Failure-path to ErrorKind mapping

| Failure path | Injection point | Expected `decision.ErrorKind` |
| --- | --- | --- |
| Unavailable | immediately-closed `httptest` server (connection refused) via real `httpTransport` | `KindUnavailable` |
| Timeout | `httptest` slow handler + short client timeout | `KindUnavailable` |
| Malformed | `httptest` returning undecodable JSON | `KindMalformedResponse` |
| Invalid request | request failing `DecisionRequest.Validate()` | `KindInvalidRequest` |
| Invalid choice | choice question with no allowed choices | `KindInvalidRequest` |
| Unsupported operation | SCORE question the adapter cannot express | `KindInvalidRequest` (wraps `clef.ErrUnsupported`) |
| Provider failure | `httptest` returning HTTP 500 | `KindProviderFailure` |

All failure paths additionally assert (via `assertDecisionKind`) that the error
is a `*decision.ProviderError` attributed to `clef.ProviderName` and matches the
corresponding `decision` sentinel (`ErrUnavailable`, `ErrMalformedResponse`,
`ErrInvalidRequest`, `ErrProviderFailure`).

## Explicit assertions

- **Default preservation**: `TestDefaultProviderUnchangedWhenNothingSelected`
  and `TestFlagProviderDefaultIsNimble` assert that nothing selected (absent or
  empty) resolves to the default nimble provider — even when `CLEF_ENABLED` is
  set — so the default provider is unchanged by the addition of Clef.
- **No accidental selection**: `TestDisabledClefNotAccidentallySelected`
  asserts that a disabled Clef provider (for every non-enabling `CLEF_ENABLED`
  value) is never `Available` and its `Decide` fails closed with
  `clef.ErrDisabled` / `decision.ErrUnavailable`.

## Determinism and model-free guarantee

Every test uses only `t.Setenv` environment control and/or in-process
`httptest` servers. No test performs a network call to a live Ollama/SystemOne
endpoint, and no test requires a live runtime. The tests are deterministic:
they assert on error kinds and provider names, not on model output. Test
servers are closed cleanly (handlers return on their own) so the suite cannot
hang during `httptest.Server.Close`.

## Affected packages

- `cmd/sop-decision-adapter` (extended test file).
- `internal/providers/clef` (exercised indirectly; unmodified).

## Validation commands

Run against the working tree:

- `go build ./...` — pass (exit 0).
- `go vet ./...` — pass (exit 0).
- `go test ./...` — pass (exit 0); all listed packages `ok`, including
  `cmd/sop-decision-adapter` and `internal/providers/clef`.

## Terminology

This report preserves the PRD terminology: **selection**, **enablement**,
**default preservation**, and **fail-closed**.
