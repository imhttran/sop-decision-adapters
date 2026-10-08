# SP-002 — Selectability Contract

**Status:** contract (documentation only, no production change)
**Scope:** `sop-decision-adapters`, stage SP-002 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline)
**Audience:** SP-005 (implementation), SP-006 (tests), SP-010 (judging), SP-004 (runtime-evidence decision)

## 0. Purpose and non-goals

This document defines, precisely and testably, what "selectable" means for this
repository. Every clause is written so that SP-005/SP-006 can implement and test
it mechanically and SP-010 can judge it.

**Non-goals (explicit):** this stage performs no implementation. It changes no
production code, no SOP policy, no lifecycle, no approval, no authorization, no
routing, and no provider defaults. The selection surface it pins already
exists; the contract only names it. Any clause that cannot be expressed without
a production change must stop this stage rather than expand scope.

## 1. Selection surface and exact tokens

**Clause S1 — Selection is a single CLI flag.**
The only selection surface is the `-provider` flag of the `decide` subcommand of
`sop-decision-adapter` (`cmd/sop-decision-adapter/main.go`), declared as:

```go
providerName = fs.String("provider", "nimble", "decision provider to use")
```

No other flag, subcommand, environment variable, config field, or default
selects a provider.

**Clause S2 — Exact accepted tokens.**
The flag is parsed by `newProvider(name, baseURL, model string, timeout time.Duration)
(cmd/sop-decision-adapter/main.go, lines 139–173)` after `strings.ToLower(strings.TrimSpace(name))`.
The exact, case-insensitive, whitespace-trimmed tokens are:

| Token   | Resolved provider                       |
|---------|-----------------------------------------|
| `nimble`| `internal/providers/nimble`             |
| `julia` | `internal/providers/julia`              |
| `clef`  | `internal/providers/clef`               |
| `""`    | `nimble` (empty resolves to the default)|

`availableProviders = "nimble, julia, clef"` is the canonical token list echoed
on the unsupported-provider error.

**Testability:** token resolution is a pure function of the flag string; a test
can call `newProvider` with each token and assert the returned `decision.Provider.Name()`.

## 2. Selection is CLI-driven only

**Clause S3 — CLI-driven, not config/env/default-driven.**
Selection is driven **exclusively** by the `-provider` flag value.
Configuration and environment values feed **provider construction** (base URL,
model, timeout, and the Clef enable flag) and **provider defaults**; they never
select a provider.

- `OLLAMA_BASE_URL`, `NIMBLE_MODEL`, `JULIA_MODEL`, `CLEF_MODEL`, `CLEF_TIMEOUT`,
  `NIMBLE_TIMEOUT`, `JULIA_TIMEOUT`, `CLEF_ENABLED` are **not** selection surfaces.
- No environment variable sets `-provider`; no config file is read for selection.
- A default (the flag default `"nimble"`, or the empty string) is not "selection"
  of anything but the pre-existing default provider.

**Testability:** with every environment variable set to a Clef-selecting-looking
value, the resolved provider still depends only on the `-provider` argument.

## 3. Opt-in and absence-of-selection semantics

**Clause S4 — Explicit opt-in only.**
Only an explicit `-provider clef` (modulo case/whitespace) selects the Clef
provider. Nothing else selects Clef.

**Clause S5 — Absence/empty selects nothing Clef.**
An absent flag (default) or an empty `-provider` value resolves to `nimble` and
therefore selects **no** Clef. "Absent" and "empty" both leave the default
provider in place; neither is an implicit Clef opt-in.

**Clause S6 — Fail closed on unknown tokens.**
Any token outside `{"", "nimble", "julia", "clef"}` returns the existing
error and exit code 2:

```
unsupported provider %q (available: nimble, julia, clef)
```

Fail-closed behavior is mandatory: there is no fallthrough to Clef and no silent
substitution.

**Testability:** `newProvider("NIMBLE  ", ...).Name() == "nimble"`;
`newProvider("", ...).Name() == "nimble"`; `newProvider("bogus", ...)` returns
a non-nil error mentioning `available: nimble, julia, clef`.

## 4. Selection is independent of enablement and of configuration defaults

**Clause S7 — Selection ≠ enablement.**
Selecting `clef` constructs the provider; it does **not** enable it. Enablement
is the separate gate `CLEF_ENABLED` read by `envBool` in
`internal/providers/clef/config.go`, surfaced through `Config.Enable` /
`Config.Enabled()`.

- `newProvider("clef", ...)` calls `clef.ConfigFromEnv()` and `clef.New(...)`;
  `clef.New` never enables Clef (`cfg.WithDefaults()` preserves `Enable`).
- A disabled Clef is never available: `Provider.Available` returns `false` when
  `!cfg.Enabled()`.
- A disabled Clef does not serve: `Provider.Decide` returns
  `ProviderError(KindUnavailable, ErrDisabled)`.

**Clause S8 — Selection is independent of configuration defaults.**
Changing configuration defaults (base URL, model, timeout) never changes which
provider is selected. Transport defaults must not implicitly enable Clef
(stated in `internal/providers/clef/config.go`).

## 5. Default-preservation for all existing providers

**Clause S9 — No default changes.**
Selection of one provider does not alter any other provider's availability,
defaults, or behavior. Specifically:

| Provider | Default status          | Preservation rule                                                                 |
|----------|-------------------------|-----------------------------------------------------------------------------------|
| `nimble` | CLI default, empty-name resolution | Remains the `-provider` default (`"nimble"`) and the `""` resolution. Unchanged. |
| `julia`  | Non-default, selectable | Selection and behavior unchanged; only reachable via explicit `-provider julia`. |
| `clef`   | Non-default, **OFF**    | Remains non-default and OFF unless both `-provider clef` **and** `CLEF_ENABLED` (1/true/yes) are set. |

**Testability:** resolving any one token returns a provider whose `Name()` is
the requested token; resolving `julia` or `clef` leaves `nimble`'s default role
intact; `clef` availability stays `false` while `CLEF_ENABLED` is unset even
when selected.

## 6. Rollback rule (immediate disable)

**Clause S10 — Rollback = deselect + disable.**
Clef is disabled immediately by doing **both** of:

1. **Deselect:** run with the default/another provider — never pass `-provider clef`
   (or unset the flag, which resolves to `nimble`).
2. **Disable:** clear/omit `CLEF_ENABLED` (any value other than `1`/`true`/`yes`,
   including unset, leaves it disabled).

**Clause S11 — Rollback changes no SOP policy.**
The rollback is immediate and requires **no** change to SOP policy, lifecycle,
approval gates, authorization, or routing, and **no** restart of agentic-sop.
It only removes the CLI selection token and the enablement environment value.

**Testability:** with Clef previously selected and enabled, running with
`-provider nimble` (and `CLEF_ENABLED` unset) yields a non-serving, non-Clef
decision path; no code path requires operator policy edits to complete rollback.

## 7. Minimum observability (secret/model-internal safe)

**Clause S12 — Minimum observable facts.**
A selectable provider must surface, through the adapter/CLI surface only:

1. **Applied selection token** — which `-provider` token was resolved.
2. **Availability result** — the boolean from `Provider.Available(ctx)`.
3. **Classified failure kind** — the `decision.ErrorKind`
   (`unavailable`, `invalid_request`, `malformed_response`, `provider_failure`)
   via `decision.ProviderError`.
4. **Resulting exit code** — `0` success, `2` invalid request / unsupported
   provider, `3` provider unavailable, `1` other failure (as implemented in
   `runDecide`).

The existing stderr warning line
(`warning: provider %q is not available`) and the mapped error lines are the
current realization of this minimum; SP-006/SP-007 may test the minimum at this
level.

**Clause S13 — Forbidden observability content.**
Observability must **never** emit: secrets or credentials, prompts or request
payloads (`State`, `Criteria`, choices content beyond what the provider result
already returns), model internals (weights, hidden state, logits), or
Clef-specific wire shapes. `decision.ProviderError` carries only the provider
name, a `Kind`, and an error string — never secrets or prompts.

**Testability:** log/error surfaces can be asserted to contain provider name,
kind, and exit code only; assertions can scan for forbidden substrings
(credentials, prompt text, `/v1/systemone` wire JSON).

## 8. Neutral-boundary rule (testable contract)

**Clause S14 — Provider-neutral decision package.**
No Clef-specific branch, type, identifier, or concept may enter the `decision`
package. Evidence: `search_files` for `clef` under `decision/` returns **no
matches**; `decision/provider.go` declares the neutral `Provider` interface
(`Name`/`Available`/`Decide`) and `decision.ProviderError` carries only a
provider name and a neutral `Kind`.

**Clause S15 — No provider-specific policy in agentic-sop.**
No provider-specific policy (approval gates, execution thresholds, safety
enforcement) may enter `agentic-sop`. Provider policy stays outside the decision
contract.

**Clause S16 — All selection behind `decision.Provider`.**
Every provider — `nimble`, `julia`, `clef` — is constructed in package `main`
(`cmd/sop-decision-adapter/main.go`) and returned as `decision.Provider`.
Selection logic lives in package `main`, not in `decision`; callers depend only
on the interface.

**Testability:** existing test `TestNoProviderSpecificTermsInPublicAPI`
(`decision/request_test.go`) already guards the neutral boundary; a grep for
`clef` in `decision/**` must return no matches; `newProvider` returns
`decision.Provider` for every token.

## 9. Citations

- SP-001 baseline: `docs/reports/clef-selectable/SP-001-selectability-baseline.md`
  §1–§5 (selection surface, tokens, enablement split, `decision.Provider` seam).
- `cmd/sop-decision-adapter/main.go` — `-provider` flag default `"nimble"`,
  `availableProviders`, `newProvider` switch (lines 139–173), `runDecide` exit
  codes and warning line.
- `internal/providers/clef/config.go` — `Config.Enable`, `Enabled()`,
  `WithDefaults` (never enables), `ConfigFromEnv`, `envBool` (`1`/`true`/`yes`).
- `internal/providers/clef/clef.go` — `Available` false when disabled;
  `Decide` returns `KindUnavailable`/`ErrDisabled`.
- `decision/provider.go` — neutral `Provider` interface.
- `decision/errors.go` — `ProviderError` (provider name + `Kind` only).

## 10. Mechanical acceptance checklist for SP-005/SP-006 and SP-010

| # | Clause | Machine check |
|---|--------|---------------|
| 1 | S1/S2 surface + tokens | flag `-provider` present; `availableProviders == "nimble, julia, clef"`; default `"nimble"` |
| 2 | S3 CLI-driven only | env vars never change resolved provider |
| 3 | S4/S5 opt-in/absence | only `clef` selects Clef; `""`/absent → `nimble` |
| 4 | S6 fail closed | unknown token → error + exit 2 |
| 5 | S7/S8 independence | selected Clef unavailable while `CLEF_ENABLED` unset |
| 6 | S9 default preservation | all three defaults unchanged |
| 7 | S10/S11 rollback | `-provider nimble` + unset `CLEF_ENABLED` disables Clef, no policy change |
| 8 | S12/S13 observability | token/availability/kind/exit code present; secrets/prompts/internals absent |
| 9 | S14–S16 neutral boundary | no `clef` in `decision/**`; all providers behind `decision.Provider` |
| 10 | No implementation | only `docs/reports/clef-selectable/SP-002-selectability-contract.md` added |
