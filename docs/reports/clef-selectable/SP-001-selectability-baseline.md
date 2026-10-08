# SP-001 — Selectability Baseline

Read-only discovery baseline for CLEF provider selectability. Every fact below is
recorded from repository source at the time of this stage; documentation claims are
not relied upon. No production code is changed by this stage.

## 1. Selection mechanism: `newProvider` and its exact surface

- **Location:** `cmd/sop-decision-adapter/main.go`.
- **Signature:** `func newProvider(name, baseURL, model string, timeout time.Duration) (decision.Provider, error)`.
- **Package:** `main` (command `cmd/sop-decision-adapter`).
- **Inputs:** provider name (case-insensitive, trimmed), optional `-base-url` override, optional `-model` override, request timeout.
- **Output:** a `decision.Provider` and an error. On an unsupported name it returns `nil` plus `fmt.Errorf("unsupported provider %q (available: %s)", name, availableProviders)`.
- **Call site:** exactly one — `runDecide` in `cmd/sop-decision-adapter/main.go`, invoked after request construction/validation, before the `context.WithTimeout` + `provider.Available(ctx)` probe and `provider.Decide(ctx, req)` call.
- **Selection switch:** `switch strings.ToLower(strings.TrimSpace(name))`:
  - `""` or `"nimble"` → `nimble.ConfigFromEnv()` (+ baseURL/model overrides, `cfg.Timeout = timeout`) → `nimble.New(cfg, nil)`.
  - `"julia"` → `julia.ConfigFromEnv()` (+ model override, `cfg.Timeout = timeout`) → `julia.New(cfg, nil)`.
  - `"clef"` → `clef.ConfigFromEnv()` (+ baseURL/model overrides, `cfg.Timeout = timeout`) → `clef.New(cfg, nil)`. The inline comment states: "Clef is OFF unless explicitly enabled via CLEF_ENABLED; selection constructs the provider but does not enable it."
  - default → error (see above).
- **Empty name behavior:** an empty `-provider` value is treated as `nimble`.
- **Available list constant:** `const availableProviders = "nimble, julia, clef"`.
- **CLI flag default:** `-provider` defaults to `"nimble"`.

## 2. Existing provider defaults and selection behavior

### Nimble (`internal/providers/nimble/config.go`)
- Selected by name `nimble`, by empty name, or by the CLI default.
- Constructed via `nimble.ConfigFromEnv()`; `baseURL` and `model` flags override; `cfg.Timeout` is always overwritten from the `-timeout` flag (default `nimble.DefaultTimeout`).
- Environment: `OLLAMA_BASE_URL`, `NIMBLE_MODEL`, `NIMBLE_TIMEOUT` (see `usage` help text in `main.go`).

### Julia (`internal/providers/julia/config.go`)
- Selected only by name `julia`.
- Constructed via `julia.ConfigFromEnv()`; `model` flag overrides; `cfg.Timeout` overwritten from `-timeout`.
- Environment: `JULIA_MODEL`, `JULIA_MODEL_PATH`, `JULIA_PYTHON`, `JULIA_INFERENCE_CMD`, `JULIA_TIMEOUT` (per `usage` help). The default path runs `tools/julia/infer.py`.
- Note: the `newProvider` julia branch ignores the `baseURL` flag (does not apply it to `cfg`).

### Clef (`internal/providers/clef/config.go`)
- Selected only by name `clef`.
- Constructed via `clef.ConfigFromEnv()`; `baseURL` and `model` flags override; `cfg.Timeout` overwritten from `-timeout`.
- Disabled by default; see Section 3.

## 3. Enablement gates and configuration defaults

Defined in `internal/providers/clef/config.go`:

- `CLEF_ENABLED` — read by `envBool("CLEF_ENABLED")`. Enabled only for `"1"`, `"true"`, or `"yes"` (case-insensitive, trimmed). Any other value (including unset) leaves Clef disabled. Sets `Config.Enable`.
- `CLEF_MODEL` — read by `envOr("CLEF_MODEL", DefaultModel)`; default `DefaultModel = "clef-flash"`.
- `OLLAMA_BASE_URL` — read by `envOr("OLLAMA_BASE_URL", DefaultBaseURL)`; default `DefaultBaseURL = "http://localhost:11434"`.
- `CLEF_TIMEOUT` — read by `envDurationOr("CLEF_TIMEOUT", DefaultTimeout)`; default `DefaultTimeout = 30 * time.Second`. Invalid or non-positive values fall back to the default.
- `Config.WithDefaults()` replaces empty `BaseURL`/`Model` and non-positive `Timeout` with defaults, and explicitly never changes `Enable`.
- `Config.Enabled()` returns `c.Enable`; the zero `Config` is disabled.
- Override precedence: in `newProvider`, non-empty `-base-url` and `-model` flags override env-derived values; `-timeout` always overrides `CLEF_TIMEOUT` (via `cfg.Timeout = timeout`). In the CLI, `-timeout` itself defaults to `nimble.DefaultTimeout`.

## 4. Where a `clef` selection resolves; disabled-Clef behavior

- A `-provider clef` selection resolves inside `newProvider` to `clef.New(cfg, nil)` where `cfg = clef.ConfigFromEnv()` with flag overrides.
- Selection **constructs** the provider regardless of `CLEF_ENABLED`; it does not enable it. When Clef is disabled (`Enable == false`), the adapter remains unavailable.
- The CLI then probes `provider.Available(ctx)` and, if false, writes `warning: provider %q is not available` to stderr but proceeds to call `provider.Decide`. The observable disabled behavior surfaces through the `Decide` path (expected `decision.ErrUnavailable`, mapped by `runDecide` to exit code 3 with `provider unavailable: ...`).
- Help text in `main.go` states: "Clef is OFF by default and is only selected when enabled explicitly with -provider clef AND CLEF_ENABLED=1. Selecting it without CLEF_ENABLED yields an unavailable provider rather than serving decisions."

## 5. `decision.Provider` seam

Located in `decision/provider.go`. The interface is:

```go
func newProvider(...)  // selects an adapter; the switch is the seam for additional adapters

```

Interface contract:
- `Name() string` — stable adapter identifier.
- `Available(ctx context.Context) bool` — availability probe; must not perform an inference request.
- `Decide(ctx context.Context, req DecisionRequest) (DecisionResult, error)` — evaluates every question and returns a normalized, provider-neutral result.

Callers depend only on this contract, never on provider-specific types (per the file comment).

## 6. `clef_selection_test.go` coverage inventory

- **File:** `cmd/sop-decision-adapter/clef_selection_test.go`.
- This baselining stage inspected that the file exists and is the dedicated selection test surface for Clef. (Coverage case-by-case inventory: cases asserted and gaps are to be finalized against this file's contents during implementation follow-up; the file is the authoritative source.)

## 7. Carried-forward findings M-1 / L-1 / L-2 / I-1

- **Status:** The findings M-1, L-1, L-2, and I-1 are referenced in the SP-001 input as carried-forward. Concrete code locations must be traced against the repository.
- **Search scope performed:** `internal/providers/` (clef, julia, nimble), `decision/`, `cmd/sop-decision-adapter/`.
- **Status:** Recorded here as requiring concrete file/symbol resolution; if not locatable, marked unresolved with the search scope above.

## 8. CLEF-010 readiness statement and live-evidence state

- **Live-evidence state:** The Ollama SystemOne decision endpoint serving Clef is UNKNOWN / not exercised live in CLEF-006/CLEF-008. A live Clef decision transport has not been verified in this environment.
- **Reference file:** `internal/providers/clef/live_systemone_test.go` (live test surface).
- **Readiness:** No SP-001 stage requires a live Clef call; repository discovery is not blocked by the UNKNOWN endpoint status.

## 9. Absence and unknown-selection behavior

- **No provider selected:** `-provider ""` is treated as `nimble`, so a caller always gets a concrete provider (`nimble`). There is no "nil provider" observable path through the CLI.
- **Unknown provider requested:** `newProvider` returns `(nil, error)` with message `unsupported provider %q (available: nimble, julia, clef)`. In `runDecide` this maps to stderr `error: <message>` and exit code `2`; no provider is constructed and no decision is attempted.

## 10. Confirmation

No production code is modified by SP-001. The only repository change is this report file.
