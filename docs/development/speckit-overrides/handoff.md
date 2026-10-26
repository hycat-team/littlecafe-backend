# Handoff Delta

Apply these rules to `speckit-implement`, `speckit-converge`, and
`speckit-tasks`.

## Two independent handoffs

A feature may require one, both, or neither handoff.

### Frontend handoff

Create/update:

```text
specs/<owner>-<NNN>-<slug>/frontend-guide.md
```

from `.specify/templates/frontend-guide-template.md` when the feature changes:

- HTTP endpoints;
- event payloads visible to clients;
- client-facing DTOs;
- client-facing enums or error behavior.

Status: `DRAFT | IN_PROGRESS | READY_FOR_FRONTEND | CHANGED | DEPRECATED`.

`READY_FOR_FRONTEND` is valid only when the contract is settled, implementation
and test evidence are complete, and the counted sync is complete.

### Backend integration handoff

Create/update:

```text
specs/<owner>-<NNN>-<slug>/backend-integration-guide.md
```

from `.specify/templates/backend-integration-guide-template.md` when the feature
changes:

- gRPC methods/messages;
- event contracts/topology;
- other service-to-service integration surfaces.

Status: `DRAFT | IN_PROGRESS | READY_FOR_BACKEND | CHANGED | DEPRECATED`.

`READY_FOR_BACKEND` is valid only when the contract is settled, implementation
and test evidence are complete, and the counted sync is complete.

## Counted sync

Do not copy full schemas into a guide. Reference canonical contracts.

### Frontend guide

1. Compare routes with OpenAPI and router/handler code after normalizing path
   params.
2. Compare request/response fields with canonical schemas/generated types.
3. Compare every documented enum in both directions with its canonical
   definition.
4. Compare success and error statuses with handler behavior and Problem Details
   contracts.
5. Record numeric counts and test evidence in the guide.

### Backend integration guide

1. Compare RPCs and message fields with the canonical proto.
2. Compare error codes/trailers and deadlines with implementation and
   `contracts/grpcx` behavior.
3. Compare events, routing keys, queues, and producer/consumer pairs with
   AsyncAPI/topology.
4. Compare compatibility notes (field numbers, `reserved`, event schema) with
   contracts.
5. Record numeric counts and test evidence in the guide.

### Status

- Any remaining `N` blocks the corresponding `READY_FOR_*` status.
- A previously ready surface transitions `READY_FOR_*` → `CHANGED` while being
  revised, then back to `READY_FOR_*` when the counted sync and evidence are
  complete. No acknowledgement from the consuming team is required.

## No-change decision

When a surface does not change, `tasks.md` must state that conclusion. Do not
create the corresponding guide.

## Deferred work

Findings outside the feature go to `FEATURE_DIR/deferred.md` using
`.specify/templates/deferred-template.md`. In-scope work must not be deferred to
obtain a clean result.

## Scope guard

Do not edit `.agents/skills/**`, managed `.specify/templates/*-template.md`,
another feature's `specs/**`, `shared/`, `contracts/`, or migrations unless the
feature's `tasks.md` explicitly assigns it.
