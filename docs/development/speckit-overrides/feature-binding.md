# Feature Binding Delta

Apply this delta to every managed `speckit-*` skill before the skill writes any
file.

## Add before the skill's first write

1. Run from the repository root:

   ```powershell
   pwsh scripts/speckit/team-feature.ps1 resolve -Json
   ```

2. Use the returned values for the rest of the skill:
   `FEATURE_DIR`, `FEATURE_SPEC`, `IMPL_PLAN`, `TASKS`, `OWNER`, `LOCAL_OWNER`,
   `NUMBER`, `SLUG`, `BRANCH`.

3. If resolution fails:
   - STOP and report the error verbatim.
   - Do not fall back to the newest directory under `specs/`.
   - Do not edit `.specify/feature.json` by hand.
   - The usual fixes are: check out `feat/<slug>` or run
     `pwsh scripts/speckit/team-feature.ps1 bind`.

4. Never work when the current branch is `main`, `develop`, or differs from the
   bound feature branch. Follow Constitution Principle IX and ask before any
   branch mutation.

5. If the developer names a different feature, STOP. One checkout handles one
   feature at a time.

## Add to the skill's scope guard

Do not edit these paths unless the feature's `tasks.md` explicitly assigns them:

- `.agents/skills/**`
- `.specify/**`
- `scripts/speckit/**`
- `shared/`
- `contracts/`
- migration files
- another feature's `specs/**`

Managed skills and managed templates must remain upstream. Team additions belong
in `docs/development/speckit-overrides/`.
