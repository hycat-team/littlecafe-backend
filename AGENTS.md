# Backend Agent Guide

> Read once per session if the host has not already injected this file. Do not
> re-read it for every task; re-read only if it changed.

## 1. Startup

- Skills are workflows: `speckit-orchestrate` for a whole feature, the matching
  `speckit-*` for one step. Canonical: `.agents/skills/<name>/SKILL.md`; do not
  duplicate them under `.opencode/skills`. `.opencode/agents/*.md` is only a
  permission adapter.
- Before Spec Kit: read `docs/development/spec-kit-workflow.md` and
  `docs/development/gates.md`. Before branch/PR/release work, read
  `docs/development/git-workflow.md`.
- Before feature files: `pwsh scripts/speckit/team-feature.ps1 resolve -Json`.
- Before code: read the relevant `docs/system-design/` and `docs/development/`
  docs. Before config/secrets/env/deploy: read
  `docs/development/environments.md`.
- Only run Spec Kit when the user invokes it; not for small changes. Small tasks
  still follow the invariants and gate below.

## 2. Sources

- Requirements/ADRs: `littlecafe-docs`; contracts: `contracts/`.
- Environment/release policy: `littlecafe-docs/deployment/environments.md`.
- Conflict priority: Constitution > littlecafe-docs > contracts > specs > source.
- Source/tests are evidence; stop and report conflicts.

## 3. Invariants

- **Constitution**: `.specify/memory/constitution.md` is supreme for every task.
- **Boundaries**: no cross-service import or DB access. Reads via gRPC;
  mutations via events + transactional outbox.
- **Contracts first**: OpenAPI/Gin `oapi-codegen`, AsyncAPI/JSON Schema, protobuf
  for reads. Generated code committed; CI checks drift.
- **Shared code**: `shared/` is cross-cutting only; service-owned adapters and
  config live in the service. `configs/base.yaml` shared, `configs/<service>.yaml`
  owned.
- **Team**: 1 feature = 1 owner = 1 branch. `specs/<owner>-<NNN>-<slug>` on
  `feat/<slug>`; numbers per owner, slugs global. No `create-new-feature.ps1`, no
  `specs/registry.md`, no other feature's `specs/**`, and no shared file without
  `plan.md`. `.agents/skills/**`, `.specify/**`, `scripts/speckit/**` are
  lead-owned.
- **Migrations**: one open feature per service; rebase `main`; never edit applied
  migrations; run on an empty DB and stay idempotent.
- **Frontend handoff**: update `specs/<owner>-<NNN>-<slug>/frontend-guide.md` on
  client-surface changes; set `READY_FOR_FRONTEND` only when contract, code,
  tests, and counted sync are complete.
- **Secrets**: never commit; `.env.*` local/VPS mode 600.
- **Git**: no commit/push/tag/branch/reset/release unless requested.
  `$speckit-orchestrate` gets one commit per gated phase + verified `fixup!`;
  workers never. Commit only on `feat/<slug>`; never commit
  `.specify/owner.json`, `.specify/feature.json`, or another feature's `specs/**`.
- **Language**: i18n via `shared/pkg/i18n`; identifiers/logs English; specs and
  reports Vietnamese; `.env*` comments English.
- **Paths**: relative only.

## 4. Gate

- Read `docs/development/gates.md` before running a gate; it is the single source
  of truth.
- `make check` includes `generate-check`, workspace `test-race`, and
  `team-check`; never run it inside a task/phase loop. Inside a phase, run scoped
  `go test -race` only.
- If a gate step does not apply, say so; a silent skip reads as a pass.

## 5. Task routing

| Task | Read first |
| --- | --- |
| Spec Kit feature | `docs/development/spec-kit-workflow.md`, matching skill |
| Git / PR / release | `docs/development/git-workflow.md` |
| Architecture / boundaries | `docs/system-design/architecture.md` |
| API / events / gRPC | `contracts/`, `docs/system-design/design-pattern.md`, `docs/system-design/event-driven-design.md` |
| Data / migrations | `docs/system-design/data-and-migrations.md`, `docs/development/spec-kit-workflow.md` |
| Config / env / deploy | `docs/development/environments.md`, `docs/operations/deployment.md` |
| Security / runtime / AI | `docs/system-design/security.md`, `docs/system-design/runtime-observability.md`, `docs/system-design/ai-service.md` |
| Frontend handoff | `docs/development/spec-kit-workflow.md`, `.specify/templates/frontend-guide-template.md` |

## 6. Behavior

- Use the `question` tool for developer decisions, with consequences and
  "(Recommended)"; prose only for open-ended questions.
- Propose and take the next step; stop only at documented stop conditions.
- Write tests with the change; run codegen; prefer Makefile targets.
- Do not create entities, routes, RPCs, events, fixtures, or sample data to
  illustrate.
