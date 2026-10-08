# SP-005 — Selection Verification and Minimal Implementation

**Status:** verification outcome (documentation added; no production code change)
**Scope:** `sop-decision-adapters`, stage SP-005 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline; SP-002 contract; SP-003 hardening decision; SP-004 runtime-readiness decision)
**Deliverables:** this report; `cmd/sop-decision-adapter/main.go` (only if SP-002 required a change)

## 0. Outcome summary

The existing `-provider clef` selection surface in `cmd/sop-decision-adapter/main.go`
**already satisfies the SP-002 selectability contract exactly**, and SP-003 declared
**no `PRECONDITION_TO_SELECTION`** findings. Therefore SP-005 makes **no selection
change**: no second provider-selection mechanism is added, no Clef-specific branch or
type enters `decision`, no provider-specific policy enters `agentic-sop`, and every
existing provider default is preserved. The only repository change from this stage is
this report.

This is a verification-first, no-change outcome, recorded per the task instruction:
"Where repository evidence shows the existing surface already satisfies SP-002, record
that verification outcome and make no selection change."

## 1. Selection-surface inventory (SP-005-D1)

All paths below are relative to the repository root; all observations were made by
direct inspection of the baseline tree.

### 1.1 Flag parsing and default — `cmd/sop-decision-adapter/main.go`

`runDecide` declares the selection flag (Clause S1):

```go
providerName = fs.String("provider", "nimble", "decision provider to use")
```

- **Default:** `"nimble"` — unchanged from baseline.
- **No other flag, subcommand, environment variable, config field, or default selects
  a provider.** Only `-provider` feeds `newProvider`.

### 1.2 Resolution seam — `newProvider` in `cmd/sop-decision-adapter/main.go`

```go
const availableProviders = "nimble, julia, clef"

func newProvider(name, baseURL, model string, timeout time.Duration) (decision.Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "nimble":
		...
	case "julia":
		...
	case "clef":
		// Clef is OFF unless explicitly enabled via CLEF_ENABLED; selection
		// constructs the provider but does not enable it.
		...
	default:
		return nil, fmt.Errorf("unsupported provider %q (available: %s)", name, availableProviders)
	}
}
```

- Name is normalized with `strings.ToLower(strings.TrimSpace(name))`.
- Resolved tokens: `"nimble"`, `"julia"`, `"clef"`, and `""` (empty → `nimble`).
- Unknown tokens fail closed with an "unsupported provider" error (Clause S6).

### 1.3 Empty/absent value behavior — explicit statement

An **absent** `-provider` flag takes the declared default `"nimble"`. An **empty**
`-provider` value (`""`) is matched by the `case "", "nimble"` arm and resolves to the
**nimble** provider. Neither absence nor emptiness selects Clef. There is no implicit
Clef opt-in (Clause S5 satisfied).

### 1.4 Clef reachability and `CLEF_ENABLED` interaction — explicit statement

Clef is reachable **only** when the normalized token is exactly `"clef"`. Selecting it
calls `clef.ConfigFromEnv()` and `clef.New(cfg, nil)`; this constructs the provider but
**does not enable it**. Enablement remains a separate gate: `CLEF_ENABLED` is read by
`envBool` in `internal/providers/clef/config.go` and surfaced through `Config.Enable` /
`Config.Enabled()`. Selecting Clef without enablement yields a provider whose
`Available(ctx)` is `false` and whose `Decide` returns `decision.ErrUnavailable`
(`KindUnavailable`). Selection ≠ enablement (Clause S7 satisfied).

### 1.5 Existing test pinning the surface — `cmd/sop-decision-adapter/clef_selection_test.go`

Already-present tests corroborate the inventory:

- `TestNewProviderClefOffByDefault` — `CLEF_ENABLED=""` → Clef unavailable, `Decide` → `ErrUnavailable`.
- `TestNewProviderClefMixedCase` — `"  Clef "` resolves to the Clef provider.
- `TestNewProviderClefExplicitlyEnabled` — `CLEF_ENABLED=1` → enabled; `Decide` never returns `clef.ErrDisabled`.
- `TestNewProviderUnknownFailsClosed` — unknown token errors, nil provider.
- `TestAvailableProvidersIncludesClef` — pins `nimble, julia, clef`.
- `TestNewProviderNimbleAndJuliaUnchanged` — nimble and julia selection preserved.

## 2. SP-002 contract comparison (clause-by-clause)

| Clause | Requirement | Observed evidence | Result |
|--------|-------------|-------------------|--------|
| S1 | Single CLI flag `-provider`, default `"nimble"` | `fs.String("provider", "nimble", ...)` in `runDecide`; no other selection surface | SATISFIED |
| S2 | Tokens `nimble`/`julia`/`clef`/`""`; `availableProviders = "nimble, julia, clef"` | `newProvider` switch on `strings.ToLower(strings.TrimSpace(name))`; `availableProviders` const | SATISFIED |
| S3 | CLI-driven only; env/config never select | Only `-provider` reaches `newProvider`; env vars feed construction/defaults only | SATISFIED |
| S4 | Explicit opt-in only for Clef | Only the `"clef"` arm constructs the Clef provider | SATISFIED |
| S5 | Absent/empty selects no Clef | `case "", "nimble"` resolves to nimble | SATISFIED |
| S6 | Fail closed on unknown tokens (exit 2) | `default:` returns error; `runDecide` maps `newProvider` error to exit 2 | SATISFIED |
| S7 | Selection ≠ enablement | `clef.New` never enables; `CLEF_ENABLED` gates `Available`/`Decide` | SATISFIED |
| S8 | Selection independent of config defaults | Transport defaults (base URL/model/timeout) do not change resolved provider | SATISFIED |
| S9 | No existing-provider default changes | nimble remains default and `""` resolution; julia unchanged; clef remains non-default and OFF | SATISFIED |
| S10 | Rollback = deselect + disable | `-provider nimble` (or unset) + unset `CLEF_ENABLED` yields non-Clef, non-serving path | SATISFIED |
| S11 | Rollback changes no SOP policy | Rollback affects only CLI token and env value; no policy/lifecycle/routing change | SATISFIED |
| S12 | Minimum observability (token, availability, kind, exit code) | `runDecide` warning line + `decision.ProviderError` kind mapping + exit codes 0/2/3/1 | SATISFIED |
| S13 | No forbidden observability content | `ProviderError` carries provider name + `Kind` + error string only | SATISFIED |
| S14 | No Clef code in `decision` | Search for `clef` under `decision/` returns no matches (see §4) | SATISFIED |
| S15 | No provider-specific policy in `agentic-sop` | Selection logic lives in package `main`; no provider policy file exists in `agentic-sop` (see §4) | SATISFIED |
| S16 | All providers behind `decision.Provider` | `newProvider` returns `decision.Provider` for every token | SATISFIED |

**Contract comparison result: all SP-002 clauses SATISFIED; zero deltas.**

## 3. SP-003 hardening-precondition list

SP-003 (`docs/reports/clef-selectable/SP-003-hardening-decision.md`) classified all four
carried-forward findings and declared **no `PRECONDITION_TO_SELECTION`**:

| Finding | Classification | Precondition change |
|---------|----------------|---------------------|
| M-1 (BaseURL not validated) | ACCEPTABLE_WITH_CONDITION | none |
| L-1 (HTTP local backend) | ACCEPTABLE_WITH_CONDITION | none |
| L-2 (no success byte cap) | POST_SELECTION_HARDENING | none |
| I-1 (invalid CLEF_TIMEOUT fallback) | NOT_APPLICABLE | none |

**No SP-003 hardening preconditions were declared.** The "minimal change if
reclassified" notes in SP-003 §2–§4 are explicitly non-authorizing traceability
records; SP-005 therefore implements no hardening. The existing fail-closed
invariants (malformed `BaseURL` → `KindProviderFailure`; bounded default timeout)
remain in force and are unchanged.

## 4. Boundary findings (observed, with paths)

- **`decision` package — no Clef-specific code.** A substring search for `clef`
  across `decision/` returns **no matches**. `decision/provider.go` declares the
  neutral `Provider` interface (`Name`/`Available`/`Decide`) and
  `decision/errors.go` declares `ProviderError` with provider name + `Kind` only.
  No Clef branch, type, identifier, or concept exists in `decision` (Clause S14).
- **`agentic-sop` — no provider-specific policy.** Selection logic lives in package
  `main` (`cmd/sop-decision-adapter/main.go`); no provider-specific approval gate,
  execution threshold, or safety-enforcement policy is added or present in the
  `agentic-sop` layer (Clause S15). This stage modifies no `agentic-sop` file.
- **No second selection mechanism.** The sole seam remains `newProvider`; no
  registry, plugin, config-driven, or env-driven selector was introduced.

## 5. Disabled/rollback path verification

- **Disabled path (selection without enablement):** with `CLEF_ENABLED` unset/
  empty, `newProvider("clef", ...)` returns a constructed Clef provider that is
  **not enabled**: `Available(ctx)` is `false` and `Decide` returns
  `decision.ErrUnavailable` (`KindUnavailable`). Verified by the existing test
  `TestNewProviderClefOffByDefault`.
- **Accidental-selection guard:** an absent default resolves to `nimble`, and an
  empty value resolves to `nimble`; only an explicit `"clef"` token reaches the
  Clef arm. Disabled Clef therefore cannot be selected accidentally.
- **Rollback path (SP-002 Clause S10):** run with `-provider nimble` (or unset the
  flag) **and** `CLEF_ENABLED` unset → the run takes a non-Clef provider and Clef
  stays disabled. No SOP policy, lifecycle, approval, authorization, or routing
  change is required, and no `agentic-sop` restart is needed (Clause S11).

## 6. Commands run and results

Executed against the working tree with the deliverable in place:

```
go build ./...   # exit 0
go vet ./...     # exit 0
go test ./...    # exit 0
```

`go test ./...` output (all packages `ok`):

```
ok  github.com/imhttran/sop-decision-adapters/cmd/sop-decision-adapter
ok  github.com/imhttran/sop-decision-adapters/decision
ok  github.com/imhttran/sop-decision-adapters/internal/benchmark
ok  github.com/imhttran/sop-decision-adapters/internal/providers/clef
ok  github.com/imhttran/sop-decision-adapters/internal/providers/julia
ok  github.com/imhttran/sop-decision-adapters/internal/providers/nimble
ok  github.com/imhttran/sop-decision-adapters/internal/shadow
ok  github.com/imhttran/sop-decision-adapters/tests
```

The affected packages are `cmd/sop-decision-adapter`, `decision`, and
`internal/providers/clef`. All three required commands pass.

## 7. Acceptance-criteria disposition

| Criterion | Disposition | Evidence |
|-----------|-------------|----------|
| Selection surface matches SP-002 exactly | SATISFIED | §2 per-clause table (S1–S16 all SATISFIED) |
| Default and existing-provider behavior unchanged | SATISFIED | §1.1–§1.3; `TestNewProviderNimbleAndJuliaUnchanged` |
| Disabled Clef cannot be selected accidentally | SATISFIED | §5; `TestNewProviderClefOffByDefault` |
| Enablement remains separate from selection | SATISFIED | §1.4; `clef.ConfigFromEnv` / `Enabled()` gating |
| No Clef-specific code in `decision` | SATISFIED | §4; no-match search under `decision/` |
| Build/test/vet pass | SATISFIED | §6; `go build ./...`, `go vet ./...`, `go test ./...` all exit 0 |
| No `agentic-sop` change | SATISFIED | §4; this stage changes only this report |

## 8. No-production-change confirmation

SP-005 makes no change to `cmd/sop-decision-adapter/main.go`: SP-002 required no
selection change because the existing surface already matches the contract, and
SP-003 declared no `PRECONDITION_TO_SELECTION` hardening. The only repository change
introduced by this stage is this report file,
`docs/reports/clef-selectable/SP-005-selection-implementation.md`. No Go source,
configuration, policy, lifecycle, approval, authorization, or routing file is modified.
