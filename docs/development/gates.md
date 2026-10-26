# Gates

Single source of truth for gates. Do not restate gate commands in other
documents or skills.

| Change | Gate |
| --- | --- |
| Small fix (typo, docs, config, single task) | `make lint` + `go test` for packages touched; add `make team-check` when `specs/**`, `.specify/**` or `scripts/speckit/**` changed |
| One phase of a feature | `make lint` + `go test -race -count=3` scoped to packages touched + `make team-check` |
| Before push/tag/close-out | `make check` (includes `team-check`); plus boundary/contract/Compose checks from `.github/workflows/ci.yml` when scope changes |

- `make check` includes `generate-check`, workspace `test-race`, and
  `team-check`; never run it inside a task or phase loop.
- `make check` also runs `coverage-check`: business-code statement coverage
  must stay at or above `COVER_MIN` (default `80`). The scope excludes `cmd/`,
  `internal/di/`, `internal/bootstrap/`, `database/migrations/`, generated
  files, `contracts/openapi/`, and `shared/infrastructure/`; adjust
  `COVER_EXCLUDE` in the Makefile when the boundary changes.
- Inside a phase, run scoped `go test -race` only; never run `make test-race`.
- If a gate step does not apply, say so in the report; a silent skip reads as a
  pass.
