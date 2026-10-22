# Implementation Plan: [FEATURE]

**Branch**: `feat/[slug]` | **Feature**: `specs/[owner]-[NNN]-[slug]` | **Date**: [DATE] | **Spec**: [link]

**Input**: Feature specification from `specs/[owner]-[NNN]-[slug]/spec.md`

**Note**: This template is filled in by the `$speckit-plan` command; its definition describes the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

**Language/Version**: Go (workspace version in `go.work`)

**Primary Dependencies**: [e.g., Gin, oapi-codegen, Wire, Goose, pgx or NEEDS CLARIFICATION]

**Storage**: [e.g., PostgreSQL via Goose migrations, or N/A]

**Testing**: [e.g., `go test -race` scoped to touched packages, or NEEDS CLARIFICATION]

**Target Platform**: Linux containers (Docker Compose / production)

**Project Type**: Multi-service Go workspace

**Services touched**: [identity, order, activity, wallet, ai, notification, reporting, shared]

**Performance Goals**: [domain-specific, or NEEDS CLARIFICATION]

**Constraints**: [domain-specific, or NEEDS CLARIFICATION]

**Scale/Scope**: [domain-specific, or NEEDS CLARIFICATION]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

[Gates determined based on `.specify/memory/constitution.md`]

## Project Structure

### Documentation (this feature)

```text
specs/[owner]-[NNN]-[slug]/
├── plan.md              # This file ($speckit-plan command output)
├── research.md          # Phase 0 output ($speckit-plan command)
├── data-model.md        # Phase 1 output ($speckit-plan command)
├── quickstart.md        # Phase 1 output ($speckit-plan command)
├── contracts/           # Phase 1 output ($speckit-plan command)
└── tasks.md             # Phase 2 output ($speckit-tasks command - NOT created by $speckit-plan)
```

### Source Code (repository root)

```text
contracts/
├── openapi/             # HTTP contract + generated server types
├── asyncapi/            # event contract
└── proto/               # gRPC read contracts

services/<service>/
├── cmd/<service>/       # main
├── internal/            # transport, service, repository, di
└── database/migrations/ # Goose migrations for this service

shared/                  # cross-cutting building blocks
configs/                 # shared + per-service configuration
deployments/             # Compose files
```

**Structure Decision**: [Document the selected structure and reference the real directories captured above]

## Shared Files & Conflict Hotspots

> Required for team work. List every file outside this feature's own source and
> `specs/[owner]-[NNN]-[slug]/` that the feature will create or change. Two open
> features must not touch the same row. If a row is owned by another open
> feature, STOP and re-plan; do not edit it on this branch.

| Path | Why this feature must touch it | Owner / other feature | Merge point |
|------|-------------------------------|-----------------------|-------------|
| [e.g., `contracts/openapi/components.yaml`] | [e.g., adds POST /meals] | [e.g., none] | [e.g., after main merge, before PR] |
| [e.g., `services/diet/internal/di/wire.go`] | [e.g., new provider] | | |
| [e.g., `services/diet/database/migrations/`] | [e.g., one new migration] | | |
| [e.g., `configs/diet.yaml`] | [e.g., new setting] | | |
| [e.g., `go.work`] | [e.g., new module] | | |

**Migration rule**: only one open feature may add migrations to the same service.
The later feature rebases `main` and regenerates its migration; never hand-edit
an applied migration or a generated snapshot.

**Contract rule**: `contracts/` is shared. Announce the change in this plan and
merge contract changes serially; the next feature rebases and regenerates.

**Generated code rule**: `wire_gen.go`, OpenAPI types, protobuf stubs, and other
generated files are regenerated after merge and verified by CI drift checks. Do
not hand-edit generated files.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
