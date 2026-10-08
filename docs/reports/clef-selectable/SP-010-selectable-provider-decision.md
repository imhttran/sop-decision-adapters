# SP-010 — Final Selectable-Provider Decision

**Status:** final decision (documentation only; no production code changed)
**Scope:** `sop-decision-adapters`, stage SP-010 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline; SP-002 contract; SP-003 hardening; SP-004 runtime-readiness; SP-005 selection; SP-006 selection tests; SP-007 governance/neutrality; SP-008 runtime verification; SP-009 verification/security review)
**Audience:** the human authorizer (governance)

## 0. Decision

**Classification: `SELECTABLE_READY`.**

Clef may become an explicitly selectable provider. The selection surface
(`-provider clef`) already exists, is explicit, opt-in, default-preserving,
fail-closed, neutral, and governance-safe, and no `PRECONDITION_TO_SELECTION`
finding remains open. Enabling Clef, making Clef live, or making Clef default
still requires separate human authorization; this document does **not** authorize
any such step, and no production enablement occurs.

### Recommended next action (exactly one)

```text
Authorize Clef as an explicitly selectable provider.
```

This plan does not enable Clef, make Clef live, or make Clef default. Recommending
this single action neither performs nor authorizes it.

## 1. Selection semantics for a selectable provider

Source of truth for the selection surface: `cmd/sop-decision-adapter/main.go`
(`runDecide`, `newProvider`) and the SP-002 contract
(`docs/reports/clef-selectable/SP-002-selectability-contract.md`), corroborated by
SP-005 (`SP-005-selection-implementation.md`) and SP-006
(`SP-006-selection-tests.md`).

- **Single surface.** Selection is the `-provider` flag of the `decide`
  subcommand, declared `fs.String("provider", "nimble", "decision provider to use")`.
  No other flag, subcommand, environment variable, or config file selects a
  provider (SP-002 S1, S3).
- **Exact tokens.** After `strings.ToLower(strings.TrimSpace(name))`, the resolved
  tokens are `nimble`, `julia`, `clef`, and `""` (empty -> `nimble`); the canonical
  list is `const availableProviders = "nimble, julia, clef"` (SP-002 S2).
- **Explicit opt-in only.** Only the token `clef` selects the Clef provider
  (SP-002 S4). Absent or empty `-provider` resolves to `nimble` and selects no
  Clef (SP-002 S5).
- **Selection != enablement.** `newProvider("clef", ...)` calls
  `clef.ConfigFromEnv()` and `clef.New(cfg, nil)`; `New` never enables Clef and
  `Config.WithDefaults()` preserves `Enable` (SP-002 S7; SP-005 section 1.4).
  Enablement is a separate gate, `CLEF_ENABLED`, read by `envBool` in
  `internal/providers/clef/config.go`.
- **Fail closed.** Any token outside `{"", "nimble", "julia", "clef"}` returns
  `unsupported provider %q (available: nimble, julia, clef)` and exits 2; there is
  no fallthrough to Clef and no silent substitution (SP-002 S6; SP-009 section 3.1).

**Evidence:** SP-002 Clauses S1-S8; SP-005 sections 1-2 (all clauses SATISFIED);
SP-006 requirement-to-test matrix; `cmd/sop-decision-adapter/main.go`; the
selection tests in `cmd/sop-decision-adapter/clef_selection_test.go`.

## 2. Default preservation

- **`nimble` remains the default.** The `-provider` flag default is `"nimble"`, and
  the empty value resolves to `nimble` via the explicit `case "", "nimble":` arm -
  the documented default, not a silent error fallback (SP-002 S9; SP-009 section 3.4).
- **`julia` is unchanged.** It remains non-default and reachable only via an
  explicit `-provider julia` (SP-002 S9; test `TestNewProviderNimbleAndJuliaUnchanged`).
- **`clef` remains non-default and OFF.** It is reachable only by an explicit
  `-provider clef` **and** `CLEF_ENABLED` in `{1, true, yes}` (case-insensitive)
  (SP-002 S9; SP-003 section 1; SP-006 explicit assertions).
- **Defaults do not select.** Configuration defaults (base URL, model, timeout)
  never change which provider is selected, and transport defaults never enable
  Clef (SP-002 S8).

**Evidence:** SP-002 Clauses S8, S9; SP-005 sections 1.1-1.3; SP-006
`TestDefaultProviderUnchangedWhenNothingSelected`,
`TestFlagProviderDefaultIsNimble`,
`TestSelectingClefDoesNotEnableItNorAffectOthers`.

## 3. Architecture boundary

- **Neutral decision contract.** No Clef-specific branch, type, identifier, or
  concept exists in the `decision` package. Direct inspection: a substring search
  for `clef` under `decision/` returns **no matches** (this SP-010 stage re-ran the
  search; result: no matches). `decision/provider.go` declares the neutral
  `Provider` interface (`Name`/`Available`/`Decide`) and `decision/errors.go`
  declares `ProviderError` carrying only a provider name and a neutral `Kind`
  (SP-002 S14, S16).
- **Clef internals confined.** Clef-specific wire shapes (`systemOneRequest`,
  `systemOneQuestion`, `systemOneResponse`, `systemOneAnswer`, `systemOneUsage`)
  are unexported and confined to `internal/providers/clef/systemone.go`;
  `clef.go`'s package comment states these never appear in the public decision
  package (SP-009 section 4.1).
- **No provider-specific policy in `agentic-sop`.** Provider policy (approval
  gates, execution thresholds, safety enforcement) stays outside the decision
  contract; the adapter only translates and normalizes (SP-002 S15; SP-007
  Property 4; SP-009 section 4.2). This stage changes no `agentic-sop` file.
- **Single seam.** All providers are constructed in package `main` and returned as
  `decision.Provider`; no second selection mechanism (registry, plugin,
  config-driven, or env-driven) exists (SP-005 section 4; SP-002 S16).

**Evidence:** SP-002 Clauses S14-S16; SP-005 section 4; SP-007 Property 3 and
section 3 (AST-based neutrality tests in
`internal/providers/clef/governance_test.go`); SP-009 section 4.1; direct
`decision/` search performed in this stage (no matches).

## 4. Governance

Selection does not weaken governance, and this stage changes no governance state.

- **No bypass of approval, lifecycle, or authorization.** The Clef adapter's
  output carries no authoritative verdict (CONTINUE/BLOCK/APPROVE/REJECT/COMMIT/
  MERGE) and performs no lifecycle or approval transition; approval, lifecycle,
  and authorization remain SOP-owned. Selection is a pure name resolution
  returning a `decision.Provider` (SP-007 Property 1; SP-009 section 4.2).
- **No SOP policy modification.** The selection code path writes no policy symbol;
  rollback requires no SOP policy, lifecycle, approval, authorization, or routing
  change (SP-002 S11; SP-007 Property 2).
- **No `agentic-sop` change.** No provider-specific policy is added or present in
  the `agentic-sop` layer; this stage modifies no `agentic-sop` file (SP-002 S15;
  SP-007 Property 4; SP-009 section 4.2).
- **Neutrality guarded by tests.** `internal/providers/clef/governance_test.go`
  provides `TestDecisionPackageIsProviderNeutral`,
  `TestDecisionPackageHasNoProviderSpecificType`, and
  `TestDecisionPackageHasNoClefBranch`, which parse `decision/` sources and reject
  any Clef-specific token, type, or branch (SP-007 section 3; SP-009 section 4.1).

**Evidence:** SP-007 Properties 1-4; SP-002 Clauses S11, S15; SP-009 section 4;
`internal/providers/clef/governance_test.go`.

## 5. Per-finding dispositions (M-1, L-1, L-2, I-1)

SP-003 (`SP-003-hardening-decision.md` section 1) classified each carried-forward
finding and declared **no `PRECONDITION_TO_SELECTION`**. This stage carries those
dispositions forward unchanged.

| Finding | Severity | Title | SP-003 classification | SP-010 disposition |
|---------|----------|-------|-----------------------|--------------------|
| **M-1** | MEDIUM | BaseURL not parsed/validated at construction; malformed URL fails closed at request creation | `ACCEPTABLE_WITH_CONDITION` | Accepted: fail-closed invariant holds (malformed `BaseURL` -> `KindProviderFailure`, never a successful decision, never a silent endpoint substitution). Not a selectability gate. |
| **L-1** | LOW | HTTP allowed for the local backend (no TLS enforcement) | `ACCEPTABLE_WITH_CONDITION` | Accepted under condition: default endpoint is loopback (`http://localhost:11434`); any *remote* deployment must use `https://`. Selectability unaffected because the default path is loopback-only. |
| **L-2** | LOW | Successful response bodies have no explicit hard byte cap | `POST_SELECTION_HARDENING` | Deferred: defense-in-depth resource-exhaustion hardening after Clef is selectable. Bounded in wall-clock time by the client timeout; not a selectability gate. |
| **I-1** | INFO | Invalid `CLEF_TIMEOUT` falls back to the default timeout | `NOT_APPLICABLE` | Not applicable to selectability: the fallback is the safe, bounded default (30s) applied in three places; no fail-open, no unbounded timeout, no selection change. |

**No finding is a precondition to selection; none is waived.** The "minimal change
if reclassified" notes in SP-003 sections 2-4 are explicitly non-authorizing
traceability records and do not authorize any production edit here.

**Evidence:** SP-003 sections 1-6 (classifications and code evidence); SP-004
section 3; SP-005 section 3; SP-008 section 5; SP-009 sections 3.4, 5.

## 6. Runtime decision - implementation vs selection vs runtime-quality readiness

The three readiness dimensions are stated separately, per SP-004 section 1. This
stage reuses that separation without collapsing it: **selectability readiness does
not imply runtime-quality readiness.**

- **Implementation readiness - MET.** The Clef adapter normalizes transport
  failures into the provider-neutral `decision` contract and never turns a failure
  into successful evidence; it is exercised by offline unit tests
  (`internal/providers/clef/*_test.go`) and selection tests
  (`cmd/sop-decision-adapter/clef_selection_test.go`). Evidence: SP-004 section 1.1;
  SP-009 section 4.4.
- **Selection readiness - MET.** The selection surface is explicit, opt-in, and
  default-preserving, and matches the SP-002 contract clause-by-clause with zero
  deltas. Evidence: SP-005 section 2 (S1-S16 all SATISFIED); SP-006; SP-009 section 3.
- **Runtime-quality readiness - NOT ESTABLISHED.** The live Clef
  `POST /v1/systemone` (`clef-flash`) path is **UNKNOWN - NOT_EXERCISED /
  UNAVAILABLE**; the oMLX / `clef-4bit` path is **NOT_EXERCISED** and is not wired
  into the adapter at all. No live Clef-serving Ollama backend was reachable in
  this environment (SP-004 section 2; SP-008 section 4; SP-009 section 3.5). No
  latency, success-rate, choice-validity, or confidence number is reported,
  because none was observed.

**Decision for selectability: live evidence is NOT required (SP-004 section 3;
SP-008 section 1).** Selection readiness and implementation readiness are
independent of runtime quality readiness; a selected, enabled Clef whose backend is
unreachable or whose response is malformed fails closed (`KindUnavailable` /
`KindMalformedResponse`), never serving a decision. **This report does not claim the
Ollama SystemOne decision endpoint serving Clef is validated.** The live path
remains UNKNOWN and is surfaced as a residual risk (section 11).

**Evidence:** SP-004 sections 1-4; SP-008 sections 1, 3, 4; SP-009 sections 3.5, 5.

## 7. Failure behavior

Failure behavior is fail-closed throughout, unchanged by this stage.

- **Disabled Clef never serves.** `Provider.Available` returns `false` when
  `!cfg.Enabled()`; `Provider.Decide` returns
  `ProviderError(KindUnavailable, ErrDisabled)`; `New` never enables Clef
  (SP-002 S7; SP-009 section 3.4).
- **Unclassified/edge failures fail closed.** `classifyTransportError` maps
  connectivity -> `KindUnavailable`, undecodable response -> `KindMalformedResponse`,
  otherwise -> `KindProviderFailure`; a failure is never a successful result
  (SP-009 sections 3.3, 3.4). An indeterminate normalized response is
  `KindMalformedResponse`, never a success.
- **CLI exit codes.** `0` success; `2` invalid request/unsupported provider/parse
  error/unknown command; `3` provider unavailable; `1` other failure (SP-002 S12;
  SP-009 sections 3.1, 3.3).
- **Bounded time, not hanging.** `runDecide` applies
  `context.WithTimeout(context.Background(), *timeout)` to `Available`, `Decide`,
  and the underlying `http.NewRequestWithContext`; an unresponsive endpoint fails
  after the bounded timeout (SP-009 section 3.3).

**Evidence:** SP-002 S12; SP-006 failure-path-to-ErrorKind matrix and
`TestSelectionFailClosedPaths`; SP-008 section 5; SP-009 sections 3.3-3.4.

## 8. Observability

Minimum observable facts (SP-002 S12), realized on the adapter/CLI surface only:

1. **Applied selection token** - the resolved `-provider` token.
2. **Availability result** - the boolean from `Provider.Available(ctx)` (the stderr
   line `warning: provider %q is not available`).
3. **Classified failure kind** - `decision.ErrorKind`
   (`unavailable`, `invalid_request`, `malformed_response`, `provider_failure`)
   via `decision.ProviderError`.
4. **Resulting exit code** - `0`/`1`/`2`/`3` as in `runDecide`.

**Forbidden content (SP-002 S13):** observability must never emit secrets or
credentials, prompts or request payloads, model internals (weights, hidden state,
logits), or Clef-specific wire shapes. `decision.ProviderError` carries only the
provider name, a `Kind`, and an error string. Clef's configuration surface reads
only a base URL, a model, a timeout, and a boolean enable flag - there is no
credential variable that can leak; `printResult` prints neither configuration nor
environment values (SP-007 Property 7; SP-009 section 3.2).

**Evidence:** SP-002 S12, S13; SP-007 Property 7; SP-009 section 3.2;
`decision/errors.go`.

## 9. Rollback

Rollback = **deselect + disable** (SP-002 S10), and it changes no SOP policy
(SP-002 S11).

1. **Deselect:** run with the default/another provider - never pass `-provider clef`
   (or unset the flag, which resolves to `nimble`).
2. **Disable:** clear/omit `CLEF_ENABLED` (any value other than `1`/`true`/`yes`,
   including unset, leaves Clef disabled).

Rollback is immediate and requires **no** change to SOP policy, lifecycle,
approval gates, authorization, or routing, and **no** restart of `agentic-sop`.
It removes only the CLI selection token and the enablement environment value.

**Evidence:** SP-002 Clauses S10, S11; SP-005 section 5; SP-007 Property 6;
SP-008 section 5; rollback-path tests in
`cmd/sop-decision-adapter/clef_selection_test.go`
(`TestDisabledClefNotAccidentallySelected`,
`TestDefaultProviderUnchangedWhenNothingSelected`,
`TestSelectingClefDoesNotEnableItNorAffectOthers`).

## 10. Test boundary

- **In scope (offline, deterministic, model-free).** The selection surface is
  pinned by `cmd/sop-decision-adapter/clef_selection_test.go`: explicit Clef
  selection, default preservation, disabled-Clef accidental-selection guard,
  empty/absent/unknown selection, selection-not-enabling-Clef, and fail-closed
  classification for each failure path (unavailable, timeout, malformed, invalid
  request/choice, unsupported operation, provider failure) (SP-006).
- **Neutrality guards.** `internal/providers/clef/governance_test.go` provides
  AST-based tests asserting the `decision` package has no provider-specific
  identifier, type, or branch (SP-007 section 3; SP-009 section 4.1).
- **In adapter scope.** Offline adapter tests (`clef_test.go`,
  `capability_test.go`, `config_test.go`, `governance_test.go`,
  `translate_test.go`, `transport_test.go`) exercise validation, translation,
  normalization, and fail-closed classification (SP-004 section 1.1).
- **Out of scope (opt-in, live, not run).**
  `internal/providers/clef/live_systemone_test.go` (CLEF-006) is disabled by
  default, opted in via `CLEF_LIVE_TEST` (separate from `CLEF_ENABLED`), and skips
  cleanly when no runtime is reachable; it asserts contract shape only and never a
  fixed model verdict (SP-008 section 3). It was not run in SP-010, consistent with
  SP-008's determination that live evidence is not a selectability gate.
- **Gate evidence.** SP-009 ran `gofmt -l .`, `go vet ./...`, `go test -count=1 ./...`,
  `go test -race -count=1 ./...`, `go build ./...`, and `git diff --check`, all
  exit 0, with no required gate failed (SP-009 section 2). SP-010 re-runs the
  required `go build ./...`, `go test ./...`, and `go vet ./...` (section 14).

**Evidence:** SP-004 section 1.1; SP-006; SP-007 section 3; SP-008 sections 3, 6;
SP-009 section 2.

## 11. Residual risks

1. **Live runtime unverified (highest residual).** The Ollama SystemOne decision
   endpoint serving Clef (`POST /v1/systemone`, `clef-flash`) remains
   **NOT_EXERCISED / UNAVAILABLE**; wire-level behavior of a real Clef-serving
   backend (SystemOne request/response shapes, latency, choice/confidence
   population) is unverified in this environment. **Bounded** because Clef is OFF
   by default and only selected by an explicit `-provider clef` **and**
   `CLEF_ENABLED`, and every unverified runtime path fails closed to
   `KindUnavailable` / `KindMalformedResponse` rather than serving a decision
   (SP-004 section 3; SP-008 section 4; SP-009 section 3.5).
2. **oMLX / `clef-4bit` unwired.** That runtime path is **NOT_EXERCISED** and is not
   wired into the adapter at all; it is not a candidate transport and cannot be
   enabled (SP-008 sections 2, 4; SP-009 section 3.5).
3. **M-1 (MEDIUM).** No early `BaseURL` validation; a malformed URL fails closed at
   request construction (`KindProviderFailure`). Accepted with condition
   (SP-003 section 2).
4. **L-1 (LOW).** HTTP permitted for the backend; acceptable only while the backend
   is the loopback default or a remote host uses `https://` (SP-003 section 3).
5. **L-2 (LOW).** No explicit hard byte cap on success bodies; time-bounded by the
   client timeout; deferred post-selection hardening (SP-003 section 4).
6. **I-1 (INFO).** Invalid `CLEF_TIMEOUT` silently falls back to the safe bounded
   default; not applicable to selectability (SP-003 section 5).

## 12. Confirmations

- **Clef remains OFF.** `CLEF_ENABLED` is unset by default; `Config{}.Enable ==
  false`; `WithDefaults` never enables Clef; this stage changes no production code
  (SP-003 section 7; SP-005 section 8; SP-008 section 8; SP-009 section 7).
- **Clef remains non-default.** `-provider` still defaults to `"nimble"`, and the
  empty value still resolves to `nimble` (SP-002 S9; SP-006).
- **No `agentic-sop` change.** No provider-specific policy enters `agentic-sop`,
  and this stage modifies no `agentic-sop` file (SP-002 S15; SP-007 Property 4;
  SP-009 section 4.2).
- **SMALL/MEDIUM/LARGE routing unchanged.** Selection switches only on the
  provider token and never reads or branches on a routing tier; selection is a
  function of the `-provider` flag string alone (SP-002 S1/S3; SP-007 Property 5).
- **Separate human authorization required.** Making Clef selectable, live, or
  default requires separate human authorization. This plan does not enable Clef,
  make Clef live, or make Clef default, and performs no production enablement.

## 13. Acceptance-criteria disposition

| Criterion | Disposition | Evidence |
|-----------|-------------|----------|
| Exactly one of the three classifications is returned | SATISFIED - `SELECTABLE_READY` (section 0) | section 0 |
| Every required element recorded with evidence | SATISFIED - selection semantics, default-preservation, architecture boundary, governance, per-finding dispositions, runtime decision, failure behavior, observability, rollback, test boundary, residual risks | sections 1-11 |
| Confirmations (OFF / non-default / no `agentic-sop` change / routing unchanged / separate authorization) stated | SATISFIED | section 12 |
| If not `NOT_SELECTABLE_READY`, exactly one next action recommended | SATISFIED - "Authorize Clef as an explicitly selectable provider." (section 0) | section 0 |
| No production enablement occurs | SATISFIED - documentation only | section 12; section 14 |

## 14. No-production-change confirmation

The only repository change introduced by SP-010 is this report file,
`docs/reports/clef-selectable/SP-010-selectable-provider-decision.md`. No Go
source, test file, configuration, policy, lifecycle, approval, authorization, or
routing file is modified. In particular, Clef is not enabled, made live, or made
default. The required validations `go build ./...`, `go test ./...`, and
`go vet ./...` run against the otherwise-unchanged tree.

## 15. Evidence index

| Artifact | Path |
|----------|------|
| SP-001 selectability baseline | `docs/reports/clef-selectable/SP-001-selectability-baseline.md` |
| SP-002 selectability contract | `docs/reports/clef-selectable/SP-002-selectability-contract.md` |
| SP-003 hardening decision | `docs/reports/clef-selectable/SP-003-hardening-decision.md` |
| SP-004 runtime-readiness decision | `docs/reports/clef-selectable/SP-004-runtime-readiness-decision.md` |
| SP-005 selection verification | `docs/reports/clef-selectable/SP-005-selection-implementation.md` |
| SP-006 selection tests | `docs/reports/clef-selectable/SP-006-selection-tests.md` |
| SP-007 governance/neutrality | `docs/reports/clef-selectable/SP-007-governance-neutrality-verification.md` |
| SP-008 runtime verification | `docs/reports/clef-selectable/SP-008-runtime-verification.md` |
| SP-009 verification/security review | `docs/reports/clef-selectable/SP-009-verification-security-review.md` |
| CLI / selection seam | `cmd/sop-decision-adapter/main.go` (`runDecide`, `newProvider`) |
| Selection tests | `cmd/sop-decision-adapter/clef_selection_test.go` |
| Neutral decision contract | `decision/provider.go`, `decision/errors.go` |
| Neutrality guards | `internal/providers/clef/governance_test.go` |
| Clef adapter / config / transport | `internal/providers/clef/clef.go`, `internal/providers/clef/config.go`, `internal/providers/clef/transport.go` |
| Opt-in live check (not run) | `internal/providers/clef/live_systemone_test.go` |
| CLEF-002 transport selection | `docs/reports/clef-provider/CLEF-002-capability-transport.md` |
