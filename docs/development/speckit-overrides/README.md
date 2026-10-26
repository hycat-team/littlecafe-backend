# Team Skill Overrides

This directory is a repository policy overlay for Spec Kit managed skills. It is
**not** an official Spec Kit command/skill override mechanism.

Spec Kit officially supports:

- template overrides in `.specify/templates/overrides/`;
- command/skill overrides through a preset (`specify preset add/update`).

LittleCafe keeps managed `.agents/skills/**` and managed
`.specify/templates/*-template.md` at their upstream content. Team-specific
behavior lives here and is enforced through the constitution and `AGENTS.md`.

## Precedence

```text
Constitution
  > docs/development/speckit-overrides/
  > managed .agents/skills/<name>/SKILL.md
```

The managed skill remains authoritative for every step this directory does not
override.

## Delta-only rule

These files contain only additions or replacements, never a full copy of a
managed skill. Each override uses one of these forms:

- `Add before step X`
- `Replace step X`
- `Add after step X`

An agent must read the managed skill once and apply only the matching delta. Do
not restate the whole workflow here.

## Load order

Before running a managed `speckit-*` skill:

1. Read this file.
2. Read `feature-binding.md`.
3. Read the matching file below, if present.
4. Then read and execute the managed `.agents/skills/<name>/SKILL.md`.

| Skill | Additional delta |
|---|---|
| `speckit-analyze` | `feature-binding.md` only |
| `speckit-checklist` | `feature-binding.md` only |
| `speckit-clarify` | `feature-binding.md` only |
| `speckit-converge` | `feature-binding.md`, `converge.md`, `handoff.md` |
| `speckit-implement` | `feature-binding.md`, `implement.md`, `handoff.md` |
| `speckit-plan` | `feature-binding.md` only |
| `speckit-specify` | `feature-binding.md`, `specify.md` |
| `speckit-tasks` | `feature-binding.md`, `tasks.md`, `handoff.md` |
| `speckit-taskstoissues` | `feature-binding.md` only |

`speckit-orchestrate` and `speckit-worker` are repo-owned custom skills and are
not overridden here.

## Branch rule

Before any write or file mutation, apply
`docs/development/git-workflow.md` and Constitution Principle IX:

- `main`, `master`, `develop`, `trunk`, detached HEAD are protected.
- Do not edit there; ask the developer and wait for approval.
- Do not create, switch, or stash a branch without permission.
- Read-only work is allowed.

## Managed file rule

- Never hand-edit `.agents/skills/**` or managed
  `.specify/templates/*-template.md`.
- Template additions use `.specify/templates/overrides/`.
- Repo-owned templates use `.specify/templates/`.
- If Spec Kit changes a managed skill, update the matching delta or
  `docs/development/spec-kit-customizations.md`; do not patch the managed file.
