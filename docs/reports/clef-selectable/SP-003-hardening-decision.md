# SP-003 — Security-Hardening Decision

**Status:** decision (documentation only; no production code changed)
**Scope:** `sop-decision-adapters`, stage SP-003 of the Clef Selectable-Provider Review
**Baseline:** `2b0909b` (SP-001 selectability baseline; SP-002 selectability contract)
**Audience:** SP-004 (runtime-evidence decision), SP-005 (implementation), SP-006 (tests), SP-010 (judging)

## 0. Purpose

For each carried-forward finding **M-1, L-1, L-2, I-1**, this stage records the
concrete code evidence, the impact under selectability, and assigns **exactly
one** classification from the PRD's four categories:

```text
PRECONDITION_TO_SELECTION
ACCEPTABLE_WITH_CONDITION
POST_SELECTION_HARDENING
NOT_APPLICABLE
```

No finding is silently waived: each of the four appears below with an explicit
classification, rationale, and code evidence. For any finding classified
`PRECONDITION_TO_SELECTION`, the minimal in-scope change and its scope are named.

**No production code is changed by this task.** The only repository change is
this report file (`docs/reports/clef-selectable/SP-003-hardening-decision.md`).

### Classification definitions used here

- **PRECONDITION_TO_SELECTION** — the finding is a defect or risk that must be
  remediated *before* the Clef provider can be called selectable/serving.
- **ACCEPTABLE_WITH_CONDITION** — the behavior is acceptable for selectability
  only while a stated condition holds; the condition is recorded, not waived.
- **POST_SELECTION_HARDENING** — the finding is a real hardening item that may
  be addressed after Clef is selectable; it does not gate selectability.
- **NOT_APPLICABLE** — the finding does not apply to the selectability decision
  (for example, it concerns a default/fallback that is not part of the selection
  surface or has no security-relevant effect).

### Method and evidence limits

Evidence below is drawn from direct source inspection at baseline `2b0909b`:
`cmd/sop-decision-adapter/main.go` (`newProvider` switch, `runDecide`),
`internal/providers/clef/config.go`, `internal/providers/clef/transport.go`,
`internal/providers/clef/clef.go`, and `decision/errors.go`. Static evidence is
sufficient for all four classifications. **Evidence limit (recorded, not
assumed away):** the Ollama SystemOne decision endpoint serving Clef is UNKNOWN
and was not exercised live in CLEF-006/CLEF-008; no finding below depends on the
live wire behavior of a running backend, so the UNKNOWN endpoint does not block
these classifications. Live confirmation is deferred to the SP-004 runtime
capability decision.

## 1. Summary of dispositions

| Finding | Severity | Title | Classification | Precondition change |
|---------|----------|-------|----------------|---------------------|
| M-1 | MEDIUM | BaseURL not parsed/validated at construction | **ACCEPTABLE_WITH_CONDITION** | none |
| L-1 | LOW | HTTP allowed for the local backend (no TLS enforcement) | **ACCEPTABLE_WITH_CONDITION** | none |
| L-2 | LOW | Successful response bodies have no explicit hard byte cap | **POST_SELECTION_HARDENING** | none |
| I-1 | INFO | Invalid `CLEF_TIMEOUT` falls back to the default timeout | **NOT_APPLICABLE** | none |

No finding is classified `PRECONDITION_TO_SELECTION`; therefore no minimal
in-scope change is required to be named by this stage. The "Minimal change if
reclassified" notes below record, for traceability, what the change would be if
a later stage (SP-004) reverses a disposition — they are **not** preconditions
and do **not** authorize any production edit under SP-003.

---

## 2. M-1 — BaseURL not parsed/validated at construction (MEDIUM)

**Code evidence.**

- `internal/providers/clef/transport.go`, `newHTTPTransport` (lines 43–51):
  `baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/")` — the value is
  only trimmed; there is **no** `url.Parse` and no scheme/host validation.
- `internal/providers/clef/config.go`, `Config.WithDefaults` (lines 48–59):
  replaces an empty `BaseURL` with `DefaultBaseURL` but never validates a
  non-empty one.
- `internal/providers/clef/transport.go`, `SystemOne` (lines 63–66):
  `http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/v1/systemone", …)`
  is where a malformed URL first surfaces, as a plain wrapped error
  (`fmt.Errorf("clef: build request: %w", err)`).
- `internal/providers/clef/clef.go`, `classifyTransportError` (lines 152–163):
  a build-request error is not `isMalformedResponse` and not `isUnavailable`, so
  it maps to `decision.KindProviderFailure` (confirmed against
  `decision/errors.go`).

**Impact under selectability.**

Reachability: `ConfigFromEnv` reads `OLLAMA_BASE_URL` (`config.go`,
`envOr("OLLAMA_BASE_URL", DefaultBaseURL)`), and `newProvider`'s `clef` branch
(`cmd/sop-decision-adapter/main.go`) applies the `-base-url` override. A malformed
value therefore reaches `newHTTPTransport` from either env or flag. The failure
is **fail-closed**: a malformed URL never yields a successful `DecisionResult`;
it is normalized to `KindProviderFailure` and surfaces through `runDecide` as a
non-zero exit. The selection surface (the `-provider` token) is unaffected: the
provider is still selected, still constructed, and a disabled Clef remains
unavailable regardless. The defect is an ergonomics/fail-closed-quality issue,
not a silent-success or security-bypass path.

**Classification: `ACCEPTABLE_WITH_CONDITION`.**

**Condition.** Acceptable for selectability only while the fail-closed
invariant holds: any malformed `BaseURL` must never produce a successful decision
and must never be silently substituted with a different reachable endpoint. This
is satisfied today because the value is used verbatim (post-trim) and a bad value
produces `KindProviderFailure`. Selectability does not depend on early URL
validation, because a mis-set base URL is an operational misconfiguration that
fails closed, not a selectability blocker.

**Minimal change if reclassified as a precondition (NOT required; recorded for
traceability only).** Add a `url.Parse` (with a scheme+host check) for the
non-empty `BaseURL` in `Config.WithDefaults` or in `newHTTPTransport`, returning a
construction-time error. Scope: `internal/providers/clef/config.go` and/or
`internal/providers/clef/transport.go` only; one helper plus one call site; no
interface or `decision` package change. SP-003 does not perform this change.

---

## 3. L-1 — HTTP allowed for the local backend (LOW)

**Code evidence.**

- `internal/providers/clef/config.go` line 18:
  `DefaultBaseURL = "http://localhost:11434"` — the default is plaintext HTTP.
- `internal/providers/clef/transport.go`, `newHTTPTransport` (lines 43–51) and
  `SystemOne`/`ListModels`: the transport passes `baseURL` to
  `http.NewRequestWithContext(..., http.MethodPost/GET, t.baseURL+...)` and
  enforces **no** scheme check — any scheme, including `http://`, is accepted.
- No TLS-enforcement code exists anywhere in the package (searching
  `internal/providers/clef` yields no `https`-only gate or `tls.Config` pin).

**Impact under selectability.**

A Clef selection resolves and can reach the endpoint over plaintext HTTP. The
default endpoint is the loopback host `localhost:11434`, so by default traffic
does not leave the machine and there is no network eavesdropper on that path.
The finding becomes security-relevant only when an operator points
`OLLAMA_BASE_URL`/`-base-url` at a **remote** host over `http://`. Under the
selectability contract, transport configuration does not select providers
(SP-002 Clause S3/S8), and Clef remains OFF unless explicitly enabled, so this
does not change *which* provider is selected and does not gate selectability.

**Classification: `ACCEPTABLE_WITH_CONDITION`.**

**Condition.** Acceptable for selectability only while the Clef backend is the
declared **local** endpoint (`localhost`, loopback) — matching the default — or
any remote deployment uses `https://`. A deployment that sends decisions to a
remote host over `http://` violates this condition and must be rejected by the
operator until TLS is enforced. Selectability itself is not blocked, because the
local default is loopback-only.

**Minimal change if reclassified as a precondition (NOT required; recorded for
traceability only).** In `newHTTPTransport` (or `Config.WithDefaults`), reject a
non-loopback `http://` `BaseURL` and require `https://` for non-local hosts.
Scope: `internal/providers/clef/transport.go` (or `config.go`) only; no public
interface, `decision`, or policy change. SP-003 does not perform this change.

---

## 4. L-2 — Successful response bodies have no explicit hard byte cap (LOW)

**Code evidence.**

- `internal/providers/clef/transport.go`, `SystemOne` success path (lines 80–85):
  `dec := json.NewDecoder(resp.Body); dec.Decode(&out)` — reads directly from
  `resp.Body` with **no** `io.LimitReader` bound.
- `internal/providers/clef/transport.go`, `ListModels` success path (lines
  116–119): `json.NewDecoder(resp.Body).Decode(&list)` — same, unbounded decode.
- Contrast (error path **is** bounded): `SystemOne` line 76 and `ListModels`
  line 112 both use `io.ReadAll(io.LimitReader(resp.Body, 4096))` for non-200
  bodies.

**Impact under selectability.**

A malicious or malfunctioning backend that streams a very large or endless
success-status JSON body could cause high memory use or a long decode. The
`http.Client` timeout (`transport.go` `newHTTPTransport`, from `cfg.Timeout`,
default `DefaultTimeout = 30s`) bounds wall-clock exposure, so the failure mode
is bounded in time but not in bytes. This requires an already-compromised or
hostile backend on the configured URL to trigger; it does not affect the
selection surface (`-provider`), does not change which provider is selected, and
does not make Clef unselectable. It is a defense-in-depth resource-exhaustion
hardening item.

**Classification: `POST_SELECTION_HARDENING`.**

**Condition.** None required for selectability. The pre-existing request timeout
is the only bound; adding a hard byte cap is desirable hardening to be performed
after Clef is selectable, not before.

**Minimal change if reclassified as a precondition (NOT required; recorded for
traceability only).** Wrap the success `resp.Body` in
`io.LimitReader(resp.Body, <cap>)` before `json.NewDecoder` in both
`SystemOne` and `ListModels`. Scope: `internal/providers/clef/transport.go` only;
two call sites; no interface or `decision` change. SP-003 does not perform this
change.

---

## 5. I-1 — Invalid `CLEF_TIMEOUT` falls back to the default timeout (INFO)

**Code evidence.**

- `internal/providers/clef/config.go`, `envDurationOr` (lines 86–96):
  `v := strings.TrimSpace(os.Getenv(key))`; if `v == ""` returns `fallback`;
  `d, err := time.ParseDuration(v)`; `if err != nil || d <= 0 { return fallback }`;
  otherwise returns `d`. The fallback passed is `DefaultTimeout` (30s) via
  `ConfigFromEnv`.
- `internal/providers/clef/config.go`, `WithDefaults` (lines 48–59): also
  replaces a non-positive `Timeout` with `DefaultTimeout`.
- `internal/providers/clef/transport.go`, `newHTTPTransport` (lines 43–51):
  also guards `if timeout <= 0 { timeout = DefaultTimeout }`.

**Impact under selectability.**

A mistyped or non-positive `CLEF_TIMEOUT` does not disable Clef, does not select
or deselect any provider, and does not produce a permissive or infinite timeout —
it deterministically falls back to a bounded 30s. There is no security-relevant
effect: the fallback is the safe, bounded default and is applied consistently in
three places. It does not touch the selection surface and cannot make Clef
unselectable. The finding is informational (a silent, but safe, defaulting
behavior).

**Classification: `NOT_APPLICABLE`.**

**Rationale.** The finding concerns a safe fallback on a non-selection
configuration value; it has no bearing on whether Clef is selectable and no
security-relevant failure mode (no fail-open, no unbounded timeout, no selection
change). It is therefore not applicable to the selectability decision. This is
an explicit disposition, not a silent waiver: the behavior and its evidence are
recorded here and could be surfaced in observability (SP-002 Clause S12), but it
does not gate selectability.

**Condition.** None.

---

## 6. No-finding-waived statement

Each of the four carried-forward findings appears above with exactly one
classification, a rationale, and concrete code evidence:

- **M-1** → `ACCEPTABLE_WITH_CONDITION` (§2)
- **L-1** → `ACCEPTABLE_WITH_CONDITION` (§3)
- **L-2** → `POST_SELECTION_HARDENING` (§4)
- **I-1** → `NOT_APPLICABLE` (§5)

No finding is silently waived. No finding is classified
`PRECONDITION_TO_SELECTION`, so no minimal in-scope change is a precondition of
selectability; the "Minimal change if reclassified" notes in §2–§4 are
non-authorizing traceability records only.

## 7. No-production-change confirmation

The only repository change introduced by SP-003 is this report file,
`docs/reports/clef-selectable/SP-003-hardening-decision.md`. No Go source,
configuration, policy, lifecycle, approval, authorization, or routing file is
modified. The report was produced by static inspection; the required validations
`go build ./...`, `go vet ./...`, and `go test ./...` run against the unchanged
tree.
