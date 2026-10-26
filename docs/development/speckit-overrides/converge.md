# `speckit-converge` Delta

Apply these additions to the managed `speckit-converge` skill.

## Add to goal and context loading

Use the source hierarchy from `AGENTS.md`:

1. relevant requirement/workflow/ADR in `littlecafe-docs`;
2. machine-readable contracts in `contracts/`;
3. feature `spec.md`, `plan.md`, `tasks.md`;
4. the constitution as governing constraints.

Resolve and read the requirement/workflow/ADR links named by the feature
artifacts and the relevant canonical files under `contracts/`. If the
authoritative requirement source is unavailable, report that limitation and do
not claim full convergence. A conflict between feature artifacts and an
authoritative source is not a code gap: STOP and report the artifact conflict.

## Add after the user-input section

**Worker Git guard.** When this skill is invoked by `speckit-worker`, do not
execute an extension hook that performs any Git mutation. Report that hook to
the orchestrator for execution under the orchestrator's Git policy. Execute
non-Git hooks normally.

## Add to the scope assessment

A finding owned by another feature, shared test infrastructure, or a third-party
constraint must be recorded in `FEATURE_DIR/deferred.md`, not appended to
`tasks.md`. Include the concrete issue, scope boundary, unblocking condition,
and affected checked task IDs.

If a finding invalidates acceptance evidence for a checked task, reopen that
task and treat it as in-scope/blocking; it cannot remain only in `deferred.md`.

Do not move in-scope work to `deferred.md` merely to reach a clean result.

## Add to client and backend surface inventory

Determine whether the feature adds, changes, or removes:

- HTTP endpoints, event payloads, or client-facing DTOs;
- gRPC methods/messages, event contracts/topology, or other service-to-service
  surfaces.

Locate `frontend-guide.md` and/or `backend-integration-guide.md` when required.

## Add when appending convergence tasks

If a client-facing change is required but `frontend-guide.md` is missing, append
a convergence task to create it from
`.specify/templates/frontend-guide-template.md`.

If a backend integration change is required but
`backend-integration-guide.md` is missing, append a convergence task to create
it from `.specify/templates/backend-integration-guide-template.md`.

When no surface changes, confirm `tasks.md` records both conclusions and do not
create either guide.

## Replace the frontend-only close-out sync

Synchronize every required guide:

- `frontend-guide.md`: routes, request/response fields, enums, HTTP/error
  statuses.
- `backend-integration-guide.md`: RPCs, messages, fields, enum values, errors,
  events, routing keys/queues, compatibility.

Use counted checks against canonical contracts/source. A guide may become
`READY_FOR_FRONTEND` or `READY_FOR_BACKEND` only after implementation, test
evidence, and counted sync are complete.

## Add to the final report

Report both handoff paths/statuses, or the explicit no-change decision for each
surface, plus deferred entries created or observed.
