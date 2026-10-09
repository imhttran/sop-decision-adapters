# CLEF-010 — Selectable-Provider Readiness Review

- **Stage:** CLEF-010 (Selectable-Provider Readiness Review)
- **Repository:** sop-decision-adapters
- **Deliverable:** docs/reports/clef-provider/CLEF-010-readiness-decision.md
- **Result:** **READY_WITH_NOTES**
- **Next action:** Open selectable-provider review.

---

## 1. Direct answer to the required question

> Can Clef now be considered READY_FOR_SELECTABLE_REVIEW through the existing
> provider-neutral seam without requiring Clef-specific policy, lifecycle,
> approval, authorization, or execution-routing changes in agentic-sop?

**Yes.** Clef is **READY_WITH_NOTES** for entry into a _separate_ selectable-provider
review through the existing provider-neutral seam (`decision.Provider`, selected via
`newProvider`), and **no Clef-specific policy, lifecycle, approval, authorization, or
execution-routing change is required in `agentic-sop`**.

This is a **readiness-to-review** decision only. It does **not** authorize making Clef
selectable, does **not** enable Clef, does **not** make Clef live, and does **not** make
Clef default.

---

## 2. Repository state

- Repository: `sop-decision-adapters` (module `github.com/imhttran/sop-decision-adapters`).
- Reviewed at HEAD under a working tree that contains the CLEF-001..CLEF-009 report set
  and the Clef adapter package; this stage adds only `CLEF-010-readiness-decision.md`.
- Present Clef-provider reports under `docs/reports/clef-provider/`:
  - `CLEF-001-repository-truth.md`
  - `CLEF-002-capability-transport.md`
  - `CLEF-003-adapter-contract-mapping.md`
  - `CLEF-006-local-runtime-verification.md`
  - `CLEF-007-shadow-evaluation.md`
  - `CLEF-008-benchmark.md`
  - `CLEF-009-verification.md`
  - `CLEF-architecture-analysis.md`
  - `captures/` (live capture artifacts)
- No source file, SOP state, provider-selection default, or prior CLEF report is modified by
  this stage.

---

## 3. CLEF-001..CLEF-016 evidence

| Record                                     | Dedicated report artifact        | Implementation/test evidence (readiness-relevant)                                                                                                                                      |
| ------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CLEF-001 repository truth                  | present                          | `CLEF-001-repository-truth.md`: provider selection is a construction-time `switch` in `newProvider`; no registry/factory map in `decision` or adapters.                                |
| CLEF-002 capability/transport              | present                          | `CLEF-002-capability-transport.md`: verified Ollama-compatible `POST /v1/systemone` with model `clef-flash`; MLX/oMLX classified UNSUPPORTED.                                          |
| CLEF-003 adapter contract mapping          | present                          | `CLEF-003-adapter-contract-mapping.md`: contract mapping incl. `UNSUPPORTED` for an unsupported caller selection at the `newProvider` seam; §13 confirms no SOP policy in the adapter. |
| CLEF-004 implement provider                | not present                      | Implementation present: `internal/providers/clef/{clef,config,capability,translate,systemone,transport}.go` implement `decision.Provider`; selected via `newProvider`.                 |
| CLEF-005 provider contract & failure tests | not present (historical BLOCKED) | No adapter evidence claimed from it; historically superseded by CLEF-016/012/013/014 (`docs/history/PLAN-Clef-Provider-Integration-Decomposed.md`). Not invented.                      |
| CLEF-006 local runtime verification        | present                          | `CLEF-006-local-runtime-verification.md`: opt-in live test `internal/providers/clef/live_systemone_test.go`; live Clef cell recorded UNAVAILABLE (denominator 0), not zero.            |
| CLEF-007 shadow evaluation                 | present                          | `CLEF-007-shadow-evaluation.md`; `internal/shadow/shadow.go` + tests; §7 confirms no policy/provider-contract/governance change.                                                       |
| CLEF-008 benchmark                         | present                          | `CLEF-008-benchmark.md`: Nimble MEASURED; Clef UNAVAILABLE (zero-denominator); oMLX and 8-bit NOT_EXERCISED. **Implementation evidence no longer present:** the harness (`internal/benchmark/*` + `corpus.json`) was deleted unused in `77da7b9`. The recorded measurements stand; the code backing them does not. |
| CLEF-009 verification                      | present                          | `CLEF-009-verification.md`: gates/guards G1-G12; findings M-1/L-1/L-2/I-1; §6.7 confirms no policy/provider-contract/governance change.                                                |
| CLEF-010 readiness                         | this report                      | This readiness decision.                                                                                                                                                               |
| CLEF-011 register in selection             | not present                      | Implementation/test present: the `clef` case in `newProvider` (`cmd/sop-decision-adapter/main.go`) and `cmd/sop-decision-adapter/clef_selection_test.go`.                              |
| CLEF-012 translation tests                 | not present                      | Test present: `internal/providers/clef/translate_test.go`.                                                                                                                             |
| CLEF-013 fake-transport & failure tests    | not present                      | Test present: `internal/providers/clef/transport_test.go`, `internal/providers/clef/clef_test.go`.                                                                                     |
| CLEF-014 governance/neutrality tests       | not present                      | Test present: `internal/providers/clef/governance_test.go` (neutrality tests cited in §7).                                                                                             |
| CLEF-016 config/capability tests           | not present                      | Test present: `internal/providers/clef/config_test.go`, `internal/providers/clef/capability_test.go`.                                                                                  |

**Dedicated report artifacts vs. evidence.** CLEF-004 and CLEF-011..CLEF-016 have **no dedicated
report artifact** under `docs/reports/clef-provider/`. That does **not** mean no readiness
evidence exists for them: those tasks contributed the Clef provider implementation and the
adapter test suite (`clef_selection_test.go`, `translate_test.go`, `transport_test.go`,
`clef_test.go`, `governance_test.go`, `config_test.go`, `capability_test.go`), which is cited
elsewhere in this report (§5, §7). CLEF-005 remains a historical BLOCKED task with no adapter
evidence claimed. No evidence is invented. The readiness decision therefore relies on the
records that exist (CLEF-001, 002, 003, 006, 007, 008, 009 plus the architecture analysis)
**and** on the code/test evidence present in the tree; the missing dedicated CLEF-004/011..016
report artifacts are recorded as a documentation note, not as an absence of evidence.

---

## 4. Verified transport

- Endpoint: `POST {baseURL}/v1/systemone` (native SystemOne decision call), implemented in
  `internal/providers/clef/transport.go` (`httpTransport.SystemOne`).
- Availability check: `GET {baseURL}/api/tags` (`httpTransport.ListModels`); it never issues
  an inference request.
- Model: `clef-flash` (`DefaultModel` in `internal/providers/clef/config.go`).
- Config defaults (`config.go`): `OLLAMA_BASE_URL` → `http://localhost:11434`,
  `CLEF_MODEL` → `clef-flash`, `CLEF_TIMEOUT` → `30s`.
- Opt-in gates: `CLEF_ENABLED` (default false) enables the provider; `CLEF_LIVE_TEST` is a
  separate opt-in that gates the live contract-shape test. Enabling Clef and exercising the
  live test are independent decisions.

---

## 5. In-process provider-neutral seam (`decision.Provider` selected via `newProvider`)

- The public contract is `decision.Provider` (`decision/provider.go`): `Name()`, `Available()`,
  `Decide()`. Callers depend only on this interface, never on Clef-specific types.
- Selection is the construction-time `switch` in `newProvider(name, baseURL, model, timeout)`
  in `cmd/sop-decision-adapter/main.go`. Clef is reached only through this seam; there is no
  Clef branch in `decision/`.
- The Clef adapter (`internal/providers/clef/clef.go`) implements `decision.Provider` and
  translates/normalizes only; SystemOne wire shapes, the `noul` boolean representation, and
  Ollama shapes are confined to the adapter package.
- The selection seam already resolves `clef` (case/whitespace-insensitive) via
  `newProvider` as exercised by `cmd/sop-decision-adapter/clef_selection_test.go`.

---

## 6. Failure semantics

Fail-closed classification maps failures to normalized `decision.ErrorKind` values
(`clef.go`):

| Condition                                | Error kind              |
| ---------------------------------------- | ----------------------- |
| Disabled provider or unreachable backend | `KindUnavailable`       |
| Invalid request / unsupported operation  | `KindInvalidRequest`    |
| Undecodable response                     | `KindMalformedResponse` |
| Indeterminate response                   | `KindMalformedResponse` |
| Other transport failure                  | `KindProviderFailure`   |

No failure path produces success-shaped evidence. A disabled adapter returns
`KindUnavailable` with `ErrDisabled`. `Available()` returns `false` for a disabled
provider, an unreachable backend, or an absent model.

---

## 7. Governance boundary

- Clef is **OFF by default**: `Config{}`, `ConfigFromEnv()` without `CLEF_ENABLED`, and
  `WithDefaults()` all leave `Enable=false`; transport defaults never enable the provider
  (`config.go`).
- Clef is **non-default**: no Clef branch exists in `decision/`; the adapter is reached only
  when a caller explicitly selects it through `newProvider` (`G1`/`G2`, CLEF-009).
- The adapter defines no policy, approval, lifecycle, authorization, or execution-routing
  symbols; those belong to `agentic-sop` (`G11`/`G12`, CLEF-009 §6; CLEF-003 §13).
- SMALL/MEDIUM/LARGE routing is defined and owned by `agentic-sop` and is untouched by the
  adapter.

---

## 8. Verification results

- CLEF-009 verification gates/guards **G1-G12** confirm: OFF by default (G1/G2), explicit
  opt-in, no Clef branch in `decision/` (G11/G12).
- Default `go test ./...` contacts no live runtime (CLEF-009 §4.3).
- Required validations for this stage: `go build ./...`, `go vet ./...`, `go test ./...`
  (all pass offline at the reviewed baseline).

---

## 9. Security findings

Open but **non-blocking** (CLEF-009 §6.6/§7); **no unresolved Critical or High finding**:

| ID  | Severity      | Finding                                                  | Disposition                            |
| --- | ------------- | -------------------------------------------------------- | -------------------------------------- |
| M-1 | Medium        | URL allow-listing / SSRF surface; adapter is fail-closed | Non-blocking; carried as residual risk |
| L-1 | Low           | No TLS enforcement; local default endpoint               | Non-blocking; carried as residual risk |
| L-2 | Low           | Success-body size cap                                    | Non-blocking; carried as residual risk |
| I-1 | Informational | Silent timeout fallback behavior                         | Non-blocking; carried as residual risk |

---

## 10. Benchmark findings

From CLEF-008:

- Nimble: **MEASURED**.
- Clef: **UNAVAILABLE** — zero-denominator, **not measurable**; this is not a measured zero
  and carries no provider-quality claim.
- oMLX (MLX clef-4bit): **NOT_EXERCISED**.
- 8-bit variant: **NOT_EXERCISED**.
- Benchmark cells mirror the existing `newProvider` switch; the Clef cell is unavailable, not
  measured.

---

## 11. Residual risks

- **Unavailable live sample** (CLEF-006/007/008): the live Clef cell is unavailable
  (denominator 0), so contract/architecture verification is present but live measured
  quality is unavailable. Recorded as unavailable, never as zero or as provider quality.
- **Security findings M-1, L-1, L-2, I-1** (CLEF-009): open, non-blocking, no Critical/High.
- **Absent CLEF records** (CLEF-004, CLEF-005, CLEF-011..CLEF-016): not present in the tree;
  no readiness evidence is claimed from them.
- **MLX/oMLX clef-4bit**: informational **UNKNOWN**, out of scope and **non-gating**
  (CLEF-002 UNSUPPORTED; CLEF-006 §5 / CLEF-007 §6 / CLEF-009 §9 NOT EXERCISED).

---

## 12. Readiness result

Applying the decision rules:

- Seam, transport, failure semantics, and governance boundary are all **verified**.
- No required `agentic-sop` change was discovered.
- No unresolved Critical/High finding exists.
- Notes remain (unavailable live sample; M-1/L-1/L-2/I-1; absent CLEF records; MLX/oMLX
  UNKNOWN).

Therefore the result is exactly:

> **READY_WITH_NOTES**

### Decision trace

| Required element                             | Evidence                                                    | Result impact                   |
| -------------------------------------------- | ----------------------------------------------------------- | ------------------------------- |
| Seam (`decision.Provider` via `newProvider`) | `decision/provider.go`, `cmd/sop-decision-adapter/main.go`  | Verified → eligible             |
| Verified transport                           | `internal/providers/clef/transport.go`                      | Verified → eligible             |
| Failure semantics                            | `internal/providers/clef/clef.go`                           | Fail-closed → eligible          |
| Governance boundary                          | `internal/providers/clef/config.go`, CLEF-009 G1/G2/G11/G12 | OFF/non-default → eligible      |
| Verification findings                        | CLEF-009                                                    | Pass with notes                 |
| Security findings                            | CLEF-009 M-1/L-1/L-2/I-1                                    | Non-blocking notes              |
| Benchmark findings                           | CLEF-008                                                    | Clef unavailable → note         |
| Residual risks                               | CLEF-006/007/008/009                                        | Notes remain → READY_WITH_NOTES |

### Standing constraints (confirmed)

- Clef remains **OFF**.
- Clef remains **non-default**.
- **No `agentic-sop` change is required.**
- **SMALL/MEDIUM/LARGE routing remains unchanged.**
- Making Clef selectable requires **separate human approval**.

---

## 13. Next action

Exactly one next action is recommended:

```text
Open selectable-provider review.
```

This report does **not** enable Clef, make Clef live, or make Clef default. Entry into the
selectable-provider review is a human-gated step and does not by itself change provider
selection, defaults, or SOP policy.
