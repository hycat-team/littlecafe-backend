<!--
Sync Impact Report:
- Version change: 1.1.0 → 1.2.0 (MINOR — format restructure plus additive,
  compatible rules; no principle is redefined and no previously valid artefact
  becomes invalid).
  Reasoning: the constitution is translated to English and reorganized into the
  standard format (Core Principles, Technology Stack, Code Style, Access
  Control, Governance) used by the sibling backend. Every LittleCafe technical
  decision is preserved verbatim in meaning. The additive rules below were
  already true in code or contracts, so they document existing practice rather
  than change it.
- Modified principles:
  - All eight principles renumbered and rewritten in English; content unchanged.
  - Section V split into a Core Principle (Security & Privacy) and an expanded
    Access Control (RBAC) & Data Privacy section.
- Added sections:
  - Technology Stack & Constraints (consolidates decisions already fixed in
    AGENTS.md, configs/, and the previous constitution).
  - Code Style, Naming & Language Standards (additive, evidence-backed rules).
  - Validation checks under Governance (mirrors CI jobs).
- Added rules (evidence):
  - Client-facing messages in Vietnamese, identifiers/packages/symbols in
    English (evidence: apperror/validation messages are Vietnamese; all
    identifiers are English).
  - Client-facing JSON DTO fields in camelCase (evidence: `traceId` in
    contracts/openapi, `messageId`/`eventType` in contracts/asyncapi).
  - Repository lookups prefer the `Get` prefix (evidence:
    shared/domain/repositories/generic.go).
  - PR validation covers layer boundaries, error shape, JSON naming, and
    generated-contract drift (evidence: .github/workflows/ci.yml).
  - Shared-code ownership: `shared/` holds only cross-cutting building blocks;
    service-owned adapters and config sections live in the owning service
    (implemented in this change set; see Governance).
- Not adopted (no evidence or unsettled): English-only code comments (code
  comments are currently mixed-language); mapper/_helper file conventions;
  managed-platform portability (LittleCafe is VPS-only by decision).
- Templates requiring updates: verified — no update needed.
- Follow-up TODOs:
  - Resolved in this change: error-stack tests now exist for apperror,
    errorutils, the error-handler middleware, and the gRPC error bridge.
  - The gRPC error bridge now carries `code` via `google.rpc.ErrorInfo.Reason`.
    Field-level errors are reserved for
    `google.rpc.BadRequest.FieldViolations` and stay HTTP-only until a business
    RPC needs them. Tests cover both directions in `contracts/grpcx/errors_test.go`.
  - `traceId` is populated from `X-Trace-ID` set by Nginx; direct service calls
    without the header produce an empty `traceId`. Documented in
    docs/system-design/design-pattern.md.
-->

<!--
Sync Impact Report:
- Version change: 1.2.0 → 1.3.0 (MINOR — additive localization rules; one
  language-boundary constraint changes from Vietnamese-only to localized
  VI/EN).
  Reasoning: API responses, validation, email/SMS, and notification content must
  be served in the caller's language while Vietnamese remains the default.
  Identifiers, packages, exported symbols, and internal log messages stay
  English.
- Modified sections:
  - Code Style, Naming & Language Standards (language boundaries).
- Added rules:
  - Client-facing messages resolve through `shared/pkg/i18n`; the default locale
    is `vi` and `en` is supported.
  - Every message key MUST exist in both `vi` and `en` bundles; a missing
    translation fails validation.
  - gRPC propagates the locale via metadata; event envelopes carry an optional
    `locale` field for outbound notifications.
- Migration/compatibility: error response shape is unchanged; only the language
  of `title`, `detail`, `errors[].message`, and success `message` varies.
- Templates requiring updates: verified — no update needed.
-->

# LittleCafe Backend Constitution

## Core Principles

### I. Authoritative Sources & Workflow
Requirements, business rules, workflows, and ADRs live in the `littlecafe-docs`
repository. Machine-readable contracts under `contracts/` are the integration
source of truth for this backend. Source code and tests are implementation
evidence; they MUST NOT be used to change approved intent. A feature MUST NOT be
implemented until the decisions it depends on are Confirmed. When a missing
choice would change a public contract, the data model, the architecture, or
user-visible behavior, the agent MUST stop and ask instead of deciding.

### II. Service Boundaries & Data Ownership
The backend has seven services: Identity, Order, Activity, Wallet, AI,
Notification, and Reporting. Content belongs to Activity. Each service owns its
own database and migrations. Importing another service's implementation or
reading its tables is prohibited. Cross-service reads use gRPC; cross-service
mutations use RabbitMQ events.

### III. Architecture & Contracts
Every service uses Clean Architecture with Wire. HTTP is OpenAPI contract-first
with Gin and the `oapi-codegen` strict server interface. Events use
AsyncAPI/JSON Schema; protobuf is reserved for read-only cross-service
contracts. Generated code is committed and CI verifies drift. Creating sample
entities, endpoints, RPCs, or events is prohibited.

### IV. Data & Consistency
Identifiers use UUIDv7. Money is `int64`/`BIGINT` VND. Timestamps use
`TIMESTAMPTZ`; the business calendar is `Asia/Ho_Chi_Minh`. Important mutations
require an `Idempotency-Key`. Transactional outbox, inbox/dedup, retry, and DLQ
are mandatory for events. Order or Activity owns the initiating saga;
Notification and Reporting participate by choreography.

### V. Security & Privacy
Email/password uses Argon2id; Google OAuth/OIDC is supported. Access tokens are
ES256. Refresh tokens are opaque random Base64URL values and only their digest
is stored, with rotation, revoke, and reuse detection. Service-to-service calls
use mTLS; a user JWT travels separately when a service acts on behalf of a user.
RBAC combines ownership, Guardian-Child, and assignment. Sensitive data uses
AES-GCM; bank data exists only as an encrypted withdrawal snapshot. PII MUST NOT
appear in logs, events, or fixtures beyond the minimum required.

### VI. API & Realtime
REST prefix is `/api/v1`. Errors are RFC 9457 Problem Details with `code`,
`traceId`, and field errors. Cursor pagination is the default; offset is only
for admin tables. SSE has a REST fallback and a Redis Pub/Sub relay. Redis is
never an authoritative data source.

### VII. Runtime & Quality
Runtime is a single VPS with Docker Compose; images are pinned to GHCR by Git
SHA. The production migration job runs before rollout under an advisory lock.
Secrets and certificates live outside Git. Zap JSON, Prometheus, and
OpenTelemetry are the observability standard. Every PR runs formatting, vet,
race tests, contract lint, codegen drift, build, and boundary checks appropriate
to the change scope.

### VIII. Open Decisions
3D asset storage, payout entitlement, retention, performance, RPO/RTO,
accessibility, AI provider/model/vector store/retrieval policy, and any business
policy marked Undecided in the docs are open. Scaffolding, sample code, and
available dependencies are not decisions. Agents MUST NOT decide these.

## Technology Stack & Constraints

- **Language & workspace**: Go 1.26 multi-module workspace (`go.work`) with nine
  modules: seven services plus `shared` and `contracts`. Cross-service
  implementation imports MUST fail at compile time.
- **HTTP**: Gin with the `oapi-codegen` strict server interface; OpenAPI is the
  machine-readable contract.
- **Inter-service**: gRPC/protobuf for read-only queries; RabbitMQ for
  asynchronous mutation via the transactional outbox.
- **Storage**: one PostgreSQL instance (pgvector image) with seven databases and
  seven roles; Redis for cache, rate limit, cooldown, token blacklist, and SSE
  Pub/Sub.
- **Migrations**: Goose per service, SQL files embedded in the service, applied
  with a direct connection under an advisory lock. Editing an applied migration
  on a shared or live environment is prohibited.
- **DI**: Google Wire per service; `wire_gen.go` is committed and drift-checked.
- **Deployment**: Docker Compose on one VPS, images from GHCR pinned by digest,
  Nginx as the single entry point.
- **Observability**: Zap structured logging, worker run logs, messaging counters
  and gauges; Prometheus/OpenTelemetry profiles exist as scaffolding.

## Code Style, Naming & Language Standards

- **Language boundaries**: client-facing messages (errors, validation, UI text,
  email/SMS, notification content) MUST be localized through
  `shared/pkg/i18n`, with Vietnamese (`vi`) as the default and English (`en`) as
  a supported locale. Every message key MUST exist in both locale bundles.
  Identifiers, package names, exported symbols, database names, and internal log
  messages MUST be English.
- **JSON naming**: client-facing DTO fields MUST use camelCase (`traceId`,
  `eventType`, `messageId`), matching the approved contracts.
- **Repository naming**: lookup methods MUST prefer the `Get` prefix
  (`GetByID`, `GetByIDs`, `GetAll`).
- **Error shape**: every HTTP error MUST use the centralized `apperror`
  package and the RFC 9457 shape defined in
  `docs/system-design/design-pattern.md`; business error codes follow the
  `PREFIX_NNNN` registry in `littlecafe-docs/api/error-codes.md`.
- **No fabricated artifacts**: do not create entities, routes, RPCs, events,
  fixtures, or sample data to illustrate.

## Access Control (RBAC) & Data Privacy

- **System roles**: only `ADMIN` and `USER` exist as RoleSlug values; adding a
  new system role is prohibited.
- **Ownership**: every resource has an owner check; data belonging to a child
  MUST pass the Guardian-Child relationship check before access.
- **Assignment**: operational actions (staff check-in, service fulfillment)
  require an assignment or explicit permission, not merely authentication.
- **PII**: personally identifiable information MUST NOT appear in logs, events,
  fixtures, or reports beyond the minimum required. Reports use aggregated,
  anonymized data.
- **Bank data**: no reusable bank profile is stored; each withdrawal keeps only
  an AES-GCM encrypted snapshot bound to the withdrawal ID.

## Governance

- Principle changes require a superseding ADR, updated traceability, and a
  constitution version bump.
- **Frontend integration guide**: a feature that changes an HTTP endpoint, an
  event payload, or a client-facing DTO MUST ship
  `specs/<feature>/frontend-guide.md`. The guide uses one of `DRAFT`,
  `IN_PROGRESS`, `READY_FOR_FRONTEND`, `CHANGED`, `DEPRECATED`; references
  canonical contracts instead of copying schemas; states before → after; and may
  only be set to `READY_FOR_FRONTEND` after the contract is settled,
  implementation and test evidence are complete, and routes/fields/enums/status
  were compared with counted results. A feature with no client-surface change
  MUST record that conclusion in `tasks.md` instead of creating a guide.
- **Deferred work**: findings that are genuinely outside the feature MUST be
  recorded in `specs/<feature>/deferred.md` with the problem, why it is out of
  scope, what unblocks it, and any ticked task whose evidence it weakens.
  In-scope work MUST NOT be moved to deferred to reach a clean exit condition.
- **Git policy**: by default agents do not commit, push, tag, or change
  branch/history. When the user invokes `$speckit-orchestrate`, the orchestrator
  holds standing authorization to create one Conventional Commit per phase that
  passed its gate and verified `fixup!` commits, containing only phase-owned
  files. Workers MUST NOT commit. Push, tag, branch mutation, rebase, reset,
  amend, squash, live-data migration, and production actions always require
  explicit permission.
- **Validation checks**: every PR must verify layer boundaries
  (`deployments/scripts/check-boundaries.sh`), the centralized error shape, JSON
  naming for client-facing DTOs, message-bundle key parity for every supported
  locale, and generated-contract drift. Gate commands are
  defined in `AGENTS.md`; that table is the single source of truth and is not
  duplicated here.
- **Shared-code ownership**: `shared/` contains only cross-cutting building
  blocks used by multiple services and owned by none (config loader, database,
  messaging, caching, security primitives, presentation framework, generic
  repositories). A service-owned adapter MUST live in that service: Google OAuth
  in Identity, Brevo email in Notification, payment provider in Wallet, password
  and refresh-token crypto in Identity, vector entity in AI. Configuration
  follows the same rule: `configs/base.yaml` holds shared sections and
  `configs/<service>.yaml` holds sections owned by one service.

**Version**: 1.3.0 | **Ratified**: 2026-10-17 | **Last Amended**: 2026-10-17
