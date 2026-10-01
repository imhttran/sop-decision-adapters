# Documentation

Index of the documentation for `sop-decision-adapters`. Start with the project
[README](../README.md).

## Authority model

```text
requirements/  →  architecture/  →  specs/  →  plans/  →  execution
```

Requirements state the problem and goals; architecture states the structure;
specs state normative behavior; plans describe future work. When two documents
disagree, the higher-authority one wins, and the lower one SHOULD link to it
rather than restate it.

## When to read what

| Document | Read it when you want to know… |
| -------- | ------------------------------ |
| [requirements/PRD.md](requirements/PRD.md) | the problem, goals, non-goals, users, and success criteria. |
| [architecture/OVERVIEW.md](architecture/OVERVIEW.md) | the layering, package responsibilities, and the SOP boundary. |
| [specs/decision-contract.md](specs/decision-contract.md) | the normative public decision API (Provider, request, result, errors). |
| [specs/nimble-systemone.md](specs/nimble-systemone.md) | the normative Nimble ↔ `/v1/systemone` mapping, availability, and error normalization. |
| [specs/julia-adapter.md](specs/julia-adapter.md) | the normative Julia (ONNX) Runner contract, normalization rules, and error mapping. |
| [reference/cli.md](reference/cli.md) | the CLI commands, flags, and exit codes. |
| [reference/configuration.md](reference/configuration.md) | the environment variables and their defaults. |
| [guides/getting-started.md](guides/getting-started.md) | how to build and make a first decision. |
| [guides/testing.md](guides/testing.md) | how to run the offline and live test suites. |
| [plans/PLAN.md](plans/PLAN.md) | the phase roadmap and status. |
| [plans/BACKLOG.md](plans/BACKLOG.md) | deferred and out-of-scope work. |
| [history/PLAN-Phase-1.1.md](history/PLAN-Phase-1.1.md) | the completed Phase 1.1 plan (historical; non-normative). |

## Where new documentation goes

- A change to **what the project must do** → `requirements/PRD.md`.
- A change to **how the system is structured** → `architecture/OVERVIEW.md`.
- A change to **normative contract behavior** → the relevant file in `specs/`.
- A change to **future work** → `plans/`.
- A **completed or superseded** artifact → `history/`, with a non-normative
  banner linking back here.

## Related

- Project [README](../README.md)
