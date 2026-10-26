# `speckit-tasks` Delta

Apply these additions to the managed `speckit-tasks` skill.

## Add after the user-input section

**Worker Git guard.** When this skill is invoked by `speckit-worker`, do not
execute an extension hook that performs any Git mutation. Report that hook to
the orchestrator for execution under the orchestrator's Git policy. Execute
non-Git hooks normally.

## Add to task generation

Every feature must produce two explicit handoff decisions:

1. **Client surface**
   - If it changes: add tasks to create/update
     `FEATURE_DIR/frontend-guide.md` from
     `.specify/templates/frontend-guide-template.md`, maintain its lifecycle
     status, and perform a counted source/contract sync at close-out.
   - If it does not change: add one task to record that conclusion in
     `tasks.md`; do not create the guide.

2. **Backend integration surface**
   - If it changes gRPC methods/messages, event contracts/topology, or another
     service-to-service surface: add tasks to create/update
     `FEATURE_DIR/backend-integration-guide.md` from
     `.specify/templates/backend-integration-guide-template.md`, maintain its
     lifecycle status, and perform a counted sync at close-out.
   - If it does not change: add one task to record that conclusion in
     `tasks.md`; do not create the guide.

## Add to deferred-work handling

Findings outside this feature belong in `FEATURE_DIR/deferred.md` using
`.specify/templates/deferred-template.md`. In-scope work must remain an
executable task.

Never generate an unchecked task for work that is genuinely outside the
feature. Record why it is out of scope, what unblocks it, and any completed task
whose evidence it weakens.

## Add to the completion checklist

- [ ] Frontend handoff decision is explicit and `frontend-guide.md`/sync tasks
      exist exactly when required.
- [ ] Backend integration handoff decision is explicit and
      `backend-integration-guide.md`/sync tasks exist exactly when required.
- [ ] Out-of-scope work is routed to `deferred.md`, not left as an unfinished
      feature task.
