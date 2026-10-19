# Backend Agent Guide

## 1. Startup & Navigation

- Read this `AGENTS.md` completely before starting. Read only task-relevant docs;
  never scan the whole `docs/` tree.
- **Skills are loadable workflows, not documentation**:
  - Canonical skills live in `.agents/skills/<name>/SKILL.md`. OpenCode, Codex,
    Cursor, and Gemini CLI read them directly; do not create a copy under
    `.opencode/skills`.
  - `.opencode/agents/*.md` is only a permission adapter for OpenCode; it must
    never hold a second workflow.
  - For a whole feature use `speckit-orchestrate`; for a single step use the
    matching `speckit-*` skill. Invoke with the host agent's syntax: `$name`
    (Codex), `/name` (OpenCode/Claude), or the skill tool.
- Before any Spec Kit run, read `docs/development/agent-workflow.md`.
- Before writing or changing code, read the relevant `docs/system-design/` and
  `docs/development/` documents for the task. `docs/system-design/` is the
  agent's source of implementation truth and must stay accurate.
- Before changing configuration, secrets, environment-specific behavior, or
  deployment files, read `docs/development/environments.md`.

### Authoritative sources

- Business requirements, workflows, and ADRs live in the `littlecafe-docs`
  repository.
- Environment and release policy lives in
  `littlecafe-docs/deployment/environments.md`.
- Machine-readable API/event/gRPC contracts live in `contracts/` in this
  repository.
- Source code and tests are implementation evidence; when they conflict with an
  approved contract or document, stop and report the affected scope.
- Conflict priority: Constitution > `littlecafe-docs` > `contracts/` > `specs/`
  > source.

## 2. Workflow Compliance

- **MANDATORY** — full compliance with `docs/development/agent-workflow.md`.
- Large tasks/features follow the Spec Kit workflow step by step. Small tasks
  (bug fixes, docs, config) may skip Spec Kit but MUST still follow the core
  principles and the matching gate.
- Only run Spec Kit when the user invokes it (`$speckit-*`, `/speckit-*`, or the
  skill tool). Do not start it for small changes.
- Propose and take the next step yourself instead of handing it back; stop only
  where the workflow's stop conditions say to.
- Git policy: do not commit, push, tag, branch, reset, or release unless
  explicitly requested. Sole exception: when the user invokes
  `$speckit-orchestrate`, the skill holds standing authorization for exactly one
  commit per gated phase and verified `fixup!` commits. Workers never hold this
  authorization; push and every other Git mutation still require permission.

### Gate

The table below is the single source of truth for gates; do not restate gate
commands in other documents or skills.

| Change | Gate |
| --- | --- |
| Small fix (typo, docs, config, single task) | `make lint` + `go test` for the packages touched |
| One phase of a feature | `make lint` + `go test -race -count=3` scoped to the packages touched |
| Before push/tag/close-out | `make check`; boundary check, contract lint, and Compose config/smoke equivalent to `.github/workflows/ci.yml` when the scope changes |

- `make check` includes `generate-check` (Wire/OpenAPI drift) and the
  workspace-wide `test-race`; never run it inside a task or phase loop.
- Inside a phase, run scoped `go test -race` for the touched packages only;
  never run `make test-race`.
- If a gate step is skipped because it does not apply, say so in the report. A
  silent skip reads as a pass.

## 3. Code Compliance

- **MANDATORY** — comply with `.specify/memory/constitution.md` for every
  action, whether or not the Spec Kit workflow is running. The constitution
  outranks everything for architectural boundaries and principles.
- **MANDATORY** — read and follow `docs/system-design/` for architecture,
  boundaries, contracts, data, security, runtime, and design patterns before
  writing code.
- Contracts first: OpenAPI with Gin and the `oapi-codegen` strict server
  interface; AsyncAPI/JSON Schema for events; protobuf for read-only
  cross-service queries. Generated code is committed and CI verifies drift.
- Service ownership: a service must not import another service's implementation
  or read its database. Cross-service reads use gRPC; mutations use event
  contracts and the transactional outbox.
- **Shared code ownership**: `shared/` holds only cross-cutting building blocks
  used by multiple services and owned by none. A service-owned adapter or a
  config section that only one service uses MUST live in that service
  (Google OAuth in Identity, Brevo email in Notification, payment provider in
  Wallet); `configs/base.yaml` holds shared sections, `configs/<service>.yaml`
  holds service-owned sections.
- **Environment-aware coding** — follow `docs/development/environments.md`:
  - Pure constants belong in `configs/`; env holds secrets and values that
    differ per environment. Never hardcode hosts, ports, URLs, or credentials.
  - Do not branch business logic on the environment; use config or flags.
  - When adding or changing a setting, update configs, all env examples,
    dev/prod/staging Compose, CI Compose validation, and operations docs.
  - Never commit real secrets; `.env.*` stay on the machine/VPS with mode 600.
  - Migrations must run on an empty database and stay idempotent; keep the
    `db-init` and `migrate-only` contracts.
  - Deployment-behavior changes must update `docs/operations/` and the
    environment guide.
- Write tests with the change; run codegen and check drift.
- If a feature changes an HTTP endpoint, event payload, or client-facing DTO,
  update `specs/<feature>/frontend-guide.md` and only set
  `READY_FOR_FRONTEND` when contract, implementation, test evidence, and the
  counted sync check are complete.
- **Language**: client-facing messages are localized through `shared/pkg/i18n`
  (Vietnamese is the default, English is supported); identifiers, packages, and
  exported symbols are English. Internal log messages stay English. Every
  comment inside a file whose name starts with `.env` MUST be written in English.
- Do not create entities, routes, RPCs, events, fixtures, or sample data to
  illustrate.
- Prefer Makefile targets over manual commands when an equivalent target exists.

## 4. Core Principles

- **Constitution is supreme**: `.specify/memory/constitution.md` applies to
  every task, including tasks that do not use Spec Kit. When it conflicts with
  other documents, the constitution wins for architectural boundaries and
  principles.
- **Ask with the question UI, not prose**: when a developer decision is needed,
  use the `question` tool with concrete options, state each consequence, and
  mark the preferred option with "(Recommended)". Use prose only for open-ended
  questions that options would flatten.
- **Relative paths**: use relative paths when referencing files, writing docs
  and specs, and linking code; never write absolute local paths into the
  repository.
- **Language policy**: documents under `specs/` and developer-facing reports are
  written in Vietnamese; technical identifiers stay in English. Never write
  secrets into the repository.
