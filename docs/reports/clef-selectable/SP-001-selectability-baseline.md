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

## 5. `decision.Provider` seam and the selection switch

Two distinct symbols are involved; do not conflate them.

**Provider-neutral interface** — `decision/provider.go`:

```go
type Provider interface {
	Name() string
	Available(ctx context.Context) bool
	Decide(ctx context.Context, req DecisionRequest) (DecisionResult, error)
}
```

Interface contract:
- `Name() string` — stable adapter identifier.
- `Available(ctx context.Context) bool` — availability probe; must not perform an inference request.
- `Decide(ctx context.Context, req DecisionRequest) (DecisionResult, error)` — evaluates every question and returns a normalized, provider-neutral result.

Callers depend only on this contract, never on provider-specific types (per the file comment).

**Selection seam** — `func newProvider(name, baseURL, model string, timeout time.Duration) (decision.Provider, error)` in package `main` (`cmd/sop-decision-adapter/main.go`, lines 139–173). Its `switch` is the seam for additional adapters. It returns the `decision.Provider` interface, so no provider-specific type crosses the boundary. `newProvider` is **not** in `decision/`.

## 6. `clef_selection_test.go` coverage inventory

- **File:** `cmd/sop-decision-adapter/clef_selection_test.go` (6 test functions).
- **Existing coverage:**
  - `TestNewProviderClefOffByDefault` — `newProvider("clef", …)` constructs a Clef provider without `CLEF_ENABLED`; `Available()` is false; `Decide()` returns `ErrUnavailable`.
  - `TestNewProviderClefMixedCase` — name resolution trims/lowercases (`"  Clef "` → Clef).
  - `TestNewProviderClefExplicitlyEnabled` — with `CLEF_ENABLED=1`, selection constructs an enabled provider; `Decide` does not return `ErrDisabled`.
  - `TestNewProviderUnknownFailsClosed` — unknown name returns `(nil, error)` containing `"unsupported provider"`.
  - `TestAvailableProvidersIncludesClef` — the advertised list includes `clef`, `nimble`, and `julia`.
  - `TestNewProviderNimbleAndJuliaUnchanged` — `nimble` and `julia` still resolve to their providers.
- **Coverage gaps relevant to selectability (candidates for SP-006):**
  - Absent/empty `-provider` (`""`) → `nimble` default is not asserted.
  - Explicit disabled-Clef non-selection by absence/config default is not asserted beyond `TestNewProviderClefOffByDefault`.
  - Rollback (deselect + disable) has no dedicated test.
  - The `CLEF_ENABLED` value matrix (`1`/`true`/`yes` vs other) at the selection seam is not asserted.

## 7. Carried-forward findings M-1 / L-1 / L-2 / I-1

Each finding is resolved to concrete repository evidence. All four are open,
non-blocking hardening items; SP-003 decides their disposition.

- **M-1 (MEDIUM — BaseURL not parsed/validated at construction; malformed URL fails closed at request creation).**
  - Evidence: `internal/providers/clef/transport.go` `newHTTPTransport` (lines 43–51) stores `baseURL` after only `strings.TrimSpace`/`strings.TrimRight`, with no `url.Parse`. `Config.WithDefaults` (`internal/providers/clef/config.go` lines 48–59) likewise never validates `BaseURL`.
  - Behavior: a malformed URL surfaces at request construction, `http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/v1/systemone", …)` (`transport.go` lines 63–66), as a plain wrapped error. `classifyTransportError` (`clef.go` lines 152–163) maps it to `KindProviderFailure` — fail-closed, no successful evidence.
- **L-1 (LOW — HTTP is allowed for the local backend).**
  - Evidence: `internal/providers/clef/config.go` line 18 sets `DefaultBaseURL = "http://localhost:11434"`; `newHTTPTransport` (`transport.go` lines 43–51) accepts any scheme with no TLS enforcement.
- **L-2 (LOW — successful response bodies have no explicit hard byte cap).**
  - Evidence: the success decode path in `transport.go` — `SystemOne` uses `json.NewDecoder(resp.Body).Decode(&out)` (lines 80–85) and `ListModels` uses the same (lines 116–119), with no size limit. Only non-200 error bodies are bounded, via `io.LimitReader(resp.Body, 4096)` (lines 76 and 112).
- **I-1 (INFO — invalid `CLEF_TIMEOUT` falls back to the default timeout).**
  - Evidence: `internal/providers/clef/config.go` `envDurationOr` (lines 86–96) returns the fallback when `time.ParseDuration` fails or returns a non-positive duration.

## 8. CLEF-010 readiness statement and live-evidence state

- **Live-evidence state:** The Ollama SystemOne decision endpoint serving Clef is UNKNOWN / not exercised live in CLEF-006/CLEF-008. A live Clef decision transport has not been verified in this environment.
- **Reference file:** `internal/providers/clef/live_systemone_test.go` (live test surface).
- **Readiness:** No SP-001 stage requires a live Clef call; repository discovery is not blocked by the UNKNOWN endpoint status.

## 9. Absence and unknown-selection behavior

- **No provider selected:** `-provider ""` is treated as `nimble`, so a caller always gets a concrete provider (`nimble`). There is no "nil provider" observable path through the CLI.
- **Unknown provider requested:** `newProvider` returns `(nil, error)` with message `unsupported provider %q (available: nimble, julia, clef)`. In `runDecide` this maps to stderr `error: <message>` and exit code `2`; no provider is constructed and no decision is attempted.

## 10. Confirmation

No production code is modified by SP-001. The only repository change is this report file.
