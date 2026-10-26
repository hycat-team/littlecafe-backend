# `speckit-implement` Delta

Apply these additions to the managed `speckit-implement` skill.

## Add after the user-input section

**Worker Git guard.** When this skill is invoked by `speckit-worker`, do not
execute an extension hook that performs any Git mutation. Report that hook to
the orchestrator for execution under the orchestrator's Git policy. Execute
non-Git hooks normally.

## Add to context loading

`IF EXISTS`: read `frontend-guide.md`, `backend-integration-guide.md`, and
`deferred.md` from `FEATURE_DIR`.

Determine whether the feature changes:

- an HTTP endpoint, event payload, or client-facing DTO; or
- a gRPC method/message, event contract/topology, or other service-to-service
  surface.

If the client surface changes, `frontend-guide.md` is required. If the backend
integration surface changes, `backend-integration-guide.md` is required. If a
surface does not change, `tasks.md` must state that conclusion and no guide for
that surface should be created.

## Add to task execution

If work discovered during implementation is genuinely outside the feature:

- create or append `FEATURE_DIR/deferred.md` from
  `.specify/templates/deferred-template.md`;
- record the concrete problem, why it is outside scope, what unblocks it, and
  any checked task whose evidence it weakens;
- do not move in-scope work there to obtain a clean task list.

If the finding invalidates acceptance evidence for a completed task, reopen that
task and treat the work as in-scope/blocking.

## Add to close-out

- For a changed client surface, compare `frontend-guide.md` with canonical
  contracts/source and fill counts for routes, fields, enums, and HTTP/error
  statuses. Only then may its status become `READY_FOR_FRONTEND`.
- For a changed backend integration surface, compare
  `backend-integration-guide.md` with canonical proto/AsyncAPI/event contracts
  and fill counts for RPCs, messages, fields, enum values, errors, and events.
  Only then may its status become `READY_FOR_BACKEND`.
- For an unchanged surface, confirm `tasks.md` records that conclusion and the
  corresponding guide was not created.
- Out-of-scope findings must be in `deferred.md`; no in-scope work may be hidden
  there.
