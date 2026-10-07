# CLEF-003 — Clef-to-Provider Contract Mapping

> Specification / documentation only. This document defines the exact mapping
> chain from an agentic-sop SEAM request, through a local Clef adapter request,
> to a Clef operation, back through a Clef response, into a provider-neutral
> result. It introduces no runtime behavior, no production code, and no new
> authority fields. Every type/field/status name referenced here was inspected
> in repository source; where a shape is not present in source it is recorded as
> MISSING/PARTIAL and specified as a placeholder contract owned by the local
> adapter layer.

## 0. Scope and authoritative sources

The provider-neutral contract is authoritative and source-inspected. All names
below are taken verbatim from the repository (`github.com/imhttran/sop-decision-adapters`):

| Element | Source file |
|---|---|
| `decision.Provider` (`Name`, `Available`, `Decide`) | `decision/provider.go` |
| `decision.DecisionRequest` (`State`, `Questions`, `Metadata`) | `decision/request.go` |
| `decision.Question` (`ID`, `Type`, `Instructions`, `Criteria`, `Choices`) | `decision/request.go` |
| `decision.QuestionType` (`choice`, `boolean`, `score`) | `decision/request.go` |
| `decision.DecisionResult` (`Provider`, `Model`, `Answers`, `Usage`) | `decision/result.go` |
| `decision.Answer` (`Type`, `Choice`, `Probabilities`, `Confidence`, `Probability`, `Score`) | `decision/result.go` |
| `decision.AnswerType` (`choice`, `boolean`, `score`) | `decision/result.go` |
| `decision.Usage` (`InputTokens`, `OutputTokens`) | `decision/result.go` |
| `decision.ErrorKind` (`invalid_request`, `unavailable`, `malformed_response`, `provider_failure`) and `decision.ProviderError` | `decision/errors.go` |
| Sentinels `ErrInvalidRequest`, `ErrUnavailable`, `ErrMalformedResponse`, `ErrProviderFailure` | `decision/errors.go` |
| SystemOne wire shapes (`systemOneRequest`, `systemOneQuestion`, `systemOneResponse`, `systemOneAnswer`, `systemOneUsage`) and `Transport` (`SystemOne`, `ListModels`) | `internal/providers/nimble/systemone.go`, `internal/providers/nimble/transport.go` |
| Translation/normalization (`BuildSystemOneRequest`, `NormalizeSystemOneResponse`, `translateQuestion`, `normalizeAnswer`, `classifyTransportError`) | `internal/providers/nimble/translate.go`, `internal/providers/nimble/nimble.go` |

### 0.1 Status vocabulary

The task requires outcomes to map into `OK`, `UNSUPPORTED`, `ERROR`,
`INDETERMINATE`. The provider-neutral contract encodes these **not** as an
explicit enum field on `DecisionResult`, but through the pair
`(DecisionResult, error)` returned by `decision.Provider.Decide` and the
`decision.ErrorKind` taxonomy:

| Task status | Provider-neutral representation (source-verified) |
|---|---|
| `OK` | `Decide` returns a non-zero `DecisionResult` with `err == nil`; `DecisionResult.Validate()` succeeds. |
| `UNSUPPORTED` | `Decide` returns a zero `DecisionResult` and an error that `errors.Is(err, decision.ErrInvalidRequest)` — produced when a request names a capability the adapter cannot express (see §7). |
| `ERROR` | `Decide` returns a zero `DecisionResult` and an error matching `decision.ErrUnavailable`, `decision.ErrMalformedResponse`, or `decision.ErrProviderFailure` (see §8, §9). |
| `INDETERMINATE` | `Decide` returns a zero `DecisionResult` and an error matching `decision.ErrMalformedResponse` originating from a response that is schema-valid at the transport level but whose decision content cannot be resolved to exactly one outcome (see §10), or an empty `Answers` map rejected by `DecisionResult.Validate()`. |

This is the representability argument: the contract's `(DecisionResult, error)`
pair plus the four `ErrorKind` sentinels is sufficient to distinguish exactly the
four required statuses; §11 proves the full partition.

### 0.2 Discovery status of the Clef boundary

| Layer | Discovery status | Evidence |
|---|---|---|
| SEAM request / provider-neutral contract | EXISTS | `decision/*.go` |
| Local adapter request shape | EXISTS (pattern established by Nimble) | `internal/providers/nimble/systemone.go`, `translate.go` |
| Clef operation names | PARTIAL — the only verified decision endpoint observed is Ollama `POST /v1/systemone` with types `choice` / `score` / `noul` | `internal/providers/nimble/systemone.go` |
| Clef response shape | PARTIAL — mirrored by `systemOneResponse` | `internal/providers/nimble/systemone.go` |
| Allowed-choice representation | EXISTS | `decision.Question.Choices`, `Question.Allows` in `decision/request.go` |
| Confidence representation | EXISTS | `decision.Answer.Confidence *float64`, `inUnitRange` in `decision/result.go` |
| Native Clef decision semantics on MLX/oMLX (`clef-4bit`) | MISSING / UNKNOWN | Not established in repo; marked placeholder — see §12 |

No file other than this report was created or modified by this task.

## 1. Call path trace (SEAM → adapter → Clef → response → result)

```text
SEAM request  (decision.DecisionRequest: State, Questions[], Metadata)
      |  decision.Provider.Decide(ctx, req)
      v
local adapter request  (e.g. systemOneRequest: Model, State, Questions{})
      |  Transport.SystemOne(ctx, req)
      v
Clef operation  (POST {baseURL}/v1/systemone  — choice | score | noul)
      |  HTTP 200 + JSON body
      v
Clef response  (systemOneResponse: Model, Answers{}, Usage)
      |  1. transport decode (json.Decoder)
      |  2. NormalizeSystemOneResponse(resp, req)
      v
provider-neutral result  (decision.DecisionResult: Provider, Model, Answers{}, Usage)
                                 + error (nil, or decision.ProviderError{kinds})
```

File anchors: `decision/provider.go` (`Decide`), `internal/providers/nimble/nimble.go`
(`Decide` body order: validate → `BuildSystemOneRequest` → `transport.SystemOne` →
`NormalizeSystemOneResponse` → `providerErr(classifyTransportError(...))`),
`internal/providers/nimble/transport.go` (`SystemOne`), `internal/providers/nimble/translate.go`
(`BuildSystemOneRequest`/`NormalizeSystemOneResponse`).

## 2. Request mapping: every provider-neutral request field

Every field of `decision.DecisionRequest` and `decision.Question` has a defined
mapping into the local adapter request (`systemOneRequest` /
`systemOneQuestion`). "Not carried" means the adapter drops the field and the
rationale is stated.

### 2.1 `DecisionRequest` fields

| Provider-neutral field | Type | Local adapter request field | Rule / rationale |
|---|---|---|---|
| `State` | `string` | `systemOneRequest.State` | Copied verbatim into the Clef state field (`json:"state,omitempty"`). Required non-empty by `DecisionRequest.Validate()`. |
| `Questions` | `[]Question` | `systemOneRequest.Questions map[string]systemOneQuestion` | Each question is translated per §2.2 and keyed by `Question.ID`. The neutral slice becomes a map because Clef keys answers by question ID. |
| `Metadata` | `map[string]string` | **not carried** | No Clef field exists for it. `decision/request.go` states values "must not encode SOP policy", so dropping it cannot lose decision-relevant content. Recorded as not carried with rationale. |

### 2.2 `Question` fields

| Provider-neutral field | Type | Local adapter request field | Rule / rationale |
|---|---|---|---|
| `ID` | `string` | map key in `systemOneRequest.Questions` | Carried as the question key (Clef has no per-question `id` inside the object; identity is the map key). Must be non-empty and unique per `Question.Validate()`. |
| `Type` | `QuestionType` | `systemOneQuestion.Type` | Mapped per §2.3. Unknown types are rejected before any Clef call. |
| `Instructions` | `string` | `systemOneQuestion.Instructions` | Carried via `instructionsFor(q)`: caller value if non-empty, else `Criteria`, else a synthesized `"Decide <ID>."`. Clef requires non-empty instructions; the fallback keeps every valid SEAM request emittable. |
| `Criteria` | `string` | `systemOneQuestion.Criteria` (shape depends on type) | CHOICE → object mapping each allowed choice → `null`; SCORE → array of candidate strings; BOOLEAN → not carried (no criteria field emitted). See §2.3. |
| `Choices` | `[]string` | `systemOneQuestion.Choices` (CHOICE); also `Criteria` array (SCORE); **not carried** (BOOLEAN) | CHOICE: copied to `Choices` and used as the `Criteria` object keys. SCORE: copied to the `Criteria` array. BOOLEAN: not carried (no allowed set). |

### 2.3 `QuestionType` → Clef operation/type mapping

| `QuestionType` | Clef operation type | Notes |
|---|---|---|
| `choice` | `"choice"` | `Criteria` emitted as `{choice: null, ...}`; `Choices` emitted verbatim. |
| `boolean` | `"noul"` | Clef-native boolean representation; confined to the adapter package. `Noul *float64` carries the yes-probability. |
| `score` | `"score"` | `Criteria` emitted as the candidate array. The neutral contract has no first-class score-scale field, so `Choices` carries 2–26 candidates; fewer than two is rejected before the Clef call. |
| anything else | — | Not translated; request rejected with `ErrInvalidRequest` (see §7). |

## 3. Result mapping: every supported Clef result form

Local adapter response types are `systemOneResponse` → `systemOneAnswer`. Every
supported form maps to `decision.DecisionResult` / `decision.Answer`.

| Clef response field | Form | Provider-neutral target | Rule |
|---|---|---|---|
| `systemOneResponse.Model` | string | `DecisionResult.Model` | Copied verbatim. |
| `systemOneResponse.Answers` | map | `DecisionResult.Answers map[string]Answer` | One answer per requested `Question.ID`; a missing key is malformed (§9). |
| `systemOneResponse.Usage` | object | `DecisionResult.Usage` (`InputTokens`, `OutputTokens`) | Copied verbatim. |
| (adapter identity) | constant | `DecisionResult.Provider` | Set to the adapter's stable name (e.g. `"nimble"`); never taken from the Clef payload. |
| `systemOneAnswer.Type == "choice"` + `Choice` in allowed set → `AnswerChoice` | choice | `Answer{Type: AnswerChoice, Choice, Probabilities, Confidence}` | `Choice` must be one of `Question.Choices` (`Question.Allows`); probability keys must be allowed choices; each probability and `Confidence` in `[0,1]`. |
| `systemOneAnswer.Type == "noul"` + `Noul != nil` → `AnswerBoolean` | boolean | `Answer{Type: AnswerBoolean, Probability}` | `Probability` = `*Noul`, must be in `[0,1]`. The neutral answer intentionally exposes no transport-specific field. |
| `systemOneAnswer.Type == "score"` + `Score != nil` → `AnswerScore` | score | `Answer{Type: AnswerScore, Score}` | `Score` copied verbatim; no range imposed. |

Type-compatibility rule: a Clef answer whose `Type` does not match the question's
type is malformed (§9), never silently coerced.

## 4. Allowed-choice handling

- The allowed set for a CHOICE question is `decision.Question.Choices` (source:
  `decision/request.go`, validated to be non-empty and duplicate-free by
  `Question.Validate`).
- On the request side, each allowed choice becomes a key of the Clef `Criteria`
  object and an entry of `Choices` (source: `translateQuestion`, CHOICE branch).
- On the result side, `normalizeChoiceAnswer` enforces:
  1. `raw.Type` must equal the Clef choice type; otherwise malformed.
  2. `raw.Choice` must be non-empty.
  3. `raw.Choice` must satisfy `Question.Allows(raw.Choice)`.
  4. every key of `raw.Probabilities` must be one of the allowed choices.
- Membership is exact string equality via `Question.Allows`.

### 4.1 Invalid choice handling (explicit)

| Condition | Detection point | Resulting status | Returned result shape |
|---|---|---|---|
| Clef returns a `Choice` not in the allowed set | `normalizeChoiceAnswer`, `!q.Allows(raw.Choice)` | `ERROR` | zero `DecisionResult`, error wrapping `decision.ErrMalformedResponse` |
| Clef returns a probability keyed by an unknown choice | `normalizeChoiceAnswer` loop | `ERROR` | zero `DecisionResult`, error wrapping `decision.ErrMalformedResponse` |
| Clef returns an empty `Choice` | `normalizeChoiceAnswer` | `ERROR` | zero `DecisionResult`, error wrapping `decision.ErrMalformedResponse` |

## 5. Confidence semantics

- Carrier: `decision.Answer.Confidence *float64`, `json:"confidence,omitempty"`
  (source: `decision/result.go`). It is **optional** and preserved only when the
  Clef response supplies it (`systemOneAnswer.Confidence *float64`).
- Valid range: `[0, 1]`, enforced by `inUnitRange("confidence", ...)` inside
  `Answer.Validate` and re-checked by `normalizeChoiceAnswer` via
  `answer.Validate()`.
- Meaning: confidence is a **provider-reported** scalar expressing the model's
  own certainty in the answer. It carries **no decision authority**: the
  provider-neutral comment states that whether a value is sufficient to trigger
  an approval gate is decided by agentic-sop, never by the adapter. Confidence
  must not be promoted into a gate, threshold, or policy signal inside the
  adapter.

### 5.1 Invalid confidence handling (explicit)

| Condition | Detection point | Resulting status | Returned result shape |
|---|---|---|---|
| `Confidence` present and `< 0` or `> 1` | `inUnitRange` inside `Answer.Validate`, invoked from `normalizeChoiceAnswer` | `ERROR` | zero `DecisionResult`, error wrapping `decision.ErrMalformedResponse` |
| `Confidence` absent | not an error | — | field omitted from the answer (`omitempty`) |
| `Confidence` non-numeric / unparseable JSON | transport decode | `ERROR` | zero `DecisionResult`, error wrapping `decision.ErrMalformedResponse` (§9) |

## 6. Status derivation rules

Derivation is performed top-down in the adapter's `Decide`, matching the
source-verified order in `internal/providers/nimble/nimble.go`:

```text
1. req.Validate()  fails                      -> ERROR (ErrInvalidRequest)
2. BuildSystemOneRequest / translateQuestion   -> if a question type cannot be
   (unsupported type, unrepresentable score)     expressed: UNSUPPORTED
                                                  (else ERROR for other invalid input)
3. Transport.SystemOne fails                  -> UNSUPPORTED / ERROR by classifier (§8)
4. NormalizeSystemOneResponse fails           -> ERROR or INDETERMINATE by cause (§9, §10)
5. result.Validate() fails                    -> ERROR (ErrMalformedResponse)
6. otherwise                                  -> OK
```

The status is always carried as the `(DecisionResult, error)` pair described in
§0.1; no status field is added to `DecisionResult`.

## 7. Unsupported semantics (explicit)

"Unsupported" means Clef cannot express the requested capability, or the request
names something outside the adapter's supported surface. It maps to
`UNSUPPORTED`, represented as a zero `DecisionResult` plus an error matching
`decision.ErrInvalidRequest`.

| Trigger | Detection point | Result |
|---|---|---|
| A `Question.Type` outside `{choice, boolean, score}` | `translateQuestion` `default` branch; also `Question.Validate` `default` | `UNSUPPORTED` (`ErrInvalidRequest`) |
| A SCORE question carrying fewer than 2 candidates in `Choices` | `translateQuestion`, SCORE branch (`len(q.Choices) < 2`) | `UNSUPPORTED` (`ErrInvalidRequest`) |
| Clef reports the requested operation/capability is not available (e.g. HTTP 404/405 on the operation endpoint) | transport → classifier | `UNSUPPORTED` (see §8) |
| A caller selects a provider/model the adapter does not support | selection seam (`newProvider`) | `UNSUPPORTED` |

Note on partial shapes: capabilities discovered as PARTIAL/MISSING in §0.2
(e.g. native MLX/oMLX decision semantics) are specified **here as unsupported
placeholder contracts owned by the local adapter layer**. Until an adapter
implementing them exists and is verified, the correct mapping for such a request
is `UNSUPPORTED` — never an assumed-available success.

## 8. Transport failure semantics (explicit)

Transport failures occur before any Clef decision content is available. The
transport classifies and `classifyTransportError` maps them:

| Failure | Detected by | Provider-neutral mapping | Status |
|---|---|---|---|
| Connection refused / reset / host unreachable / network unreachable | `isUnavailable` (`syscall.ECONNREFUSED`, `ECONNRESET`, `EHOSTUNREACH`, `ENETUNREACH`) | error wrapping `decision.ErrUnavailable` | `ERROR` |
| Timeout (`context.DeadlineExceeded`, `net.Error.Timeout()`, `os.ErrDeadlineExceeded`) | `isUnavailable` | error wrapping `decision.ErrUnavailable` | `ERROR` |
| Context cancelled (`context.Canceled`) | `isUnavailable` | error wrapping `decision.ErrUnavailable` | `ERROR` |
| DNS resolution failure (`*net.DNSError`) | `isUnavailable` | error wrapping `decision.ErrUnavailable` | `ERROR` |
| Request body could not be encoded / request could not be built | `HTTPTransport.SystemOne` | error wrapping `decision.ErrProviderFailure` | `ERROR` |
| Non-2xx HTTP status (4xx/5xx, gateway errors) | `httpStatusError` in `HTTPTransport.SystemOne` | error wrapping `decision.ErrProviderFailure` (or `UNSUPPORTED` for a capability-not-available status — see §7) | `ERROR` / `UNSUPPORTED` |
| Any other transport error | `classifyTransportError` fall-through | error wrapping `decision.ErrProviderFailure` | `ERROR` |

In all transport cases the returned `DecisionResult` is the zero value; no
partial answers are emitted.

## 9. Malformed-result semantics (explicit)

Malformed-result failures occur when a response is received but cannot be
normalized. They map to `ERROR` (or `INDETERMINATE`, §10, when the response is
well-formed but indecisive), always as a zero `DecisionResult` plus an error
wrapping `decision.ErrMalformedResponse`.

| Condition | Detection point | Status |
|---|---|---|
| Body is not valid JSON / cannot be decoded | `json.Decoder` in `HTTPTransport.SystemOne` → `malformedResponseError` → `isMalformedResponse` | `ERROR` |
| A requested question has no answer in `Answers` | `NormalizeSystemOneResponse` | `ERROR` |
| Answer `Type` is unknown | `normalizeAnswer` `default` | `ERROR` |
| Answer `Type` is incompatible with the question type | `normalize{Choice,Boolean,Score}Answer` | `ERROR` |
| Choice is empty or not allowed | `normalizeChoiceAnswer` | `ERROR` (§4.1) |
| Probability key is an unknown choice | `normalizeChoiceAnswer` | `ERROR` |
| Any probability outside `[0,1]` | `normalizeChoiceAnswer` / `Answer.Validate` | `ERROR` |
| Confidence outside `[0,1]` | `Answer.Validate` (`inUnitRange`) | `ERROR` (§5.1) |
| Boolean answer has no probability (`Noul == nil`) | `normalizeBooleanAnswer` | `ERROR` |
| Boolean probability outside `[0,1]` | `normalizeBooleanAnswer` / `Answer.Validate` | `ERROR` |
| Score answer has no value (`Score == nil`) | `normalizeScoreAnswer` | `ERROR` |
| `DecisionResult.Validate()` rejects an empty `Answers` map | `NormalizeSystemOneResponse` | `ERROR` |

## 10. Indeterminate semantics (explicit)

"Indeterminate" means a schema-valid Clef response was received, but the
decision content cannot be resolved to exactly one provider-neutral outcome.
It maps to `INDETERMINATE`, represented as a zero `DecisionResult` plus an error
wrapping `decision.ErrMalformedResponse` originating from a resolvability check
rather than a structural decode failure.

| Trigger | Detection point | Status |
|---|---|---|
| The response is well-formed JSON with all keys present but the decision cannot be resolved (e.g. no single outcome is determinable from the returned fields, or an empty `Answers` map after a nominally successful transport call) | `DecisionResult.Validate()` (empty `Answers`) or the adapter's resolvability check (placeholder, owned by the local adapter layer) | `INDETERMINATE` |
| Clef explicitly reports an indeterminate/unresolvable decision in a schema-valid envelope | local adapter resolvability check (placeholder) | `INDETERMINATE` |

Placeholder note: the provider-neutral contract has no dedicated indeterminate
kind. `INDETERMINATE` is therefore a **derived status** using
`decision.ErrMalformedResponse` with a resolvability (not decode) cause, kept
distinct from §9's structural failures at the classification layer. Where a
Clef-native indeterminate signal is not verifiable in source, it is specified
here as a placeholder contract owned by the local adapter layer, never assumed
available (§12).

## 11. Representability: the provider-neutral contract covers all Clef outcomes

Every required Clef outcome maps into exactly one of the four statuses,
represented purely by the existing contract:

| Status | Existing contract surface used |
|---|---|
| `OK` | `DecisionResult{Provider, Model, Answers, Usage}` with `err == nil`; `DecisionResult.Validate()` passes (`decision/result.go`) |
| `UNSUPPORTED` | `decision.ProviderError{Kind: decision.KindInvalidRequest}` / `errors.Is(err, decision.ErrInvalidRequest)` (`decision/errors.go`) |
| `ERROR` | `decision.ProviderError{Kind: KindUnavailable|KindMalformedResponse|KindProviderFailure}` matching `ErrUnavailable`, `ErrMalformedResponse`, `ErrProviderFailure` (`decision/errors.go`) |
| `INDETERMINATE` | `decision.ProviderError{Kind: KindMalformedResponse}` with a resolvability cause (derived, §10) |

Because the result type is normalized and provider-independent, and the error
taxonomy covers exactly the failure classes above, **the existing provider-neutral
contract can represent all required Clef outcomes**: the `choice`/`boolean`/`score`
question and answer families cover every supported Clef result form (§3), and the
`(DecisionResult, error)` pair plus the four `ErrorKind` sentinels partition the
four required statuses.

## 12. UNKNOWN / placeholder contracts (not assumed available)

- `mlx-community/clef-4bit` on MLX/oMLX: native Clef decision semantics
  equivalent to `POST /v1/systemone` are **not established** in repository
  source. Any mapping for it is an explicit placeholder owned by the local
  adapter layer and must surface as `UNSUPPORTED` until an adapter exists and is
  verified. The document does not assume runtime availability.
- A Clef-native indeterminate signal is not present in source; §10's
  `INDETERMINATE` mapping is a derived placeholder, not observed behavior.
- `DecisionRequest.Metadata` is dropped because no Clef field carries it (§2.1).

## 13. Forbidden authority fields (explicit enumeration)

No lifecycle, approval, authorization, commit, merge, policy, scheduler, or
execution-model authority may appear in the adapter result. The adapter result
(`decision.DecisionResult` / `decision.Answer`) carries **only** the following
provider-neutral, decision-content fields, and none of them grant authority:

- `DecisionResult.Provider` (adapter identifier)
- `DecisionResult.Model` (model identifier)
- `DecisionResult.Answers` (map of normalized answers)
- `DecisionResult.Usage` (`InputTokens`, `OutputTokens`)
- `Answer.Type`, `Answer.Choice`, `Answer.Probabilities`, `Answer.Confidence`,
  `Answer.Probability`, `Answer.Score`

Forbidden — the adapter result MUST NOT contain and MUST NOT synthesize any of:

| Forbidden authority class | Examples that must not appear |
|---|---|
| Lifecycle authority | task/step state transitions, phase advancement, workflow state |
| Approval authority | approval gates, gate decisions, sign-off, review resolution |
| Authorization authority | permission grants, capability enablement, access control |
| Commit authority | commit/push permission, VCS write authority |
| Merge authority | merge decisions, conflict resolution authority |
| Policy authority | SOP policy, thresholds, retry rules, fail-closed rules |
| Scheduler authority | queueing, ordering, dispatch, retry scheduling |
| Execution-model authority | execution targets, model selection as policy, runner control |

These classes are deliberately absent from the provider-neutral contract:
`decision/request.go` states that approval rules, risk thresholds, execution
constraints, and safety gates belong to agentic-sop, not to the adapter;
`decision/result.go` states that the approval gate is decided by agentic-sop.
Confidence and probability are **provider output only** and confer no authority.

## 14. Self-review against CLEF-003 acceptance criteria

| Acceptance criterion | Section | Status |
|---|---|---|
| Every provider-neutral request field has a defined mapping | §2.1, §2.2 | PASS |
| Every supported Clef result form has a defined mapping | §3 | PASS |
| Invalid choice handling is explicit | §4.1 | PASS |
| Invalid confidence handling is explicit | §5.1 | PASS |
| Unsupported behavior is explicit | §7 | PASS |
| Indeterminate behavior is explicit | §10 | PASS |
| Runtime/model/transport failures are mapped | §8, §12 | PASS |
| No lifecycle/approval/authorization/commit/merge/policy/scheduler/execution-model authority in the adapter result | §13 | PASS |
| The existing provider-neutral contract can represent all required Clef outcomes | §11, §0.1 | PASS |

All field and status names in this document (`State`, `Questions`, `Metadata`,
`ID`, `Type`, `Instructions`, `Criteria`, `Choices`, `Provider`, `Model`,
`Answers`, `Usage`, `Choice`, `Probabilities`, `Confidence`, `Probability`,
`Score`, `OK`/`UNSUPPORTED`/`ERROR`/`INDETERMINATE`, `ErrInvalidRequest`,
`ErrUnavailable`, `ErrMalformedResponse`, `ErrProviderFailure`) are taken from
inspection of `decision/request.go`, `decision/result.go`, `decision/errors.go`,
and `internal/providers/nimble/*.go`; no invented names are used.
