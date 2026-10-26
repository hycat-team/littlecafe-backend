# `speckit-specify` Delta

Apply these changes after the skill has generated the concise feature short name
and before it creates the feature directory. Keep every other managed step.

## Replace the feature-creation step

The managed `speckit-specify` step that creates the feature directory with
`create-new-feature.ps1` is replaced by the team helper flow below.

Features are named `specs/<owner>-<NNN>-<slug>` and each feature has exactly one
branch, `feat/<slug>`. `<NNN>` counts per owner, never repository-wide. Slugs
are unique across the repository; if the branch already exists with someone
else's `meta.json`, STOP and ask for a different slug.

1. Confirm the current branch is `feat/<slug>` and its suffix equals the chosen
   slug. If they differ, STOP and ask which is correct. A `before_specify` hook
   that creates a branch must follow the same rule; the hook never chooses the
   directory name.
2. Ensure this machine has an owner handle:

   ```powershell
   pwsh scripts/speckit/team-feature.ps1 owner
   ```

   If none is configured, ask the developer once through the structured question
   mechanism. The handle is a short prefix for the feature directory: `duck`
   produces `specs/duck-001-user-auth`; 2-20 lowercase letters, digits or
   hyphens. Offer the sanitized git email local-part and git user name as
   suggestions. Then run:

   ```powershell
   pwsh scripts/speckit/team-feature.ps1 owner -Handle <handle>
   ```

   Never invent a handle and never create a feature with an empty owner.

3. Push the branch before creating the feature so the remote ref reserves the
   slug, if it is not on origin yet:

   ```powershell
   git push -u origin feat/<slug>
   ```

4. Create the feature:

   ```powershell
   pwsh scripts/speckit/team-feature.ps1 new -Slug <slug> -Description "<feature description>" -Json
   ```

   Add `-AllowExistingBranch` only when the remote `feat/<slug>` already has
   commits beyond local HEAD and you own it.

5. Parse `FEATURE_DIR`, `SPEC_FILE`, `FEATURE_NUM`, `OWNER`, `BRANCH`. Treat
   `FEATURE_DIR` as `SPECIFY_FEATURE_DIRECTORY`.
6. If the helper fails, STOP and report the error verbatim. Do not invent a
   number or slug.

## Do not

- Do not call `create-new-feature.ps1`.
- Do not scan `specs/` for the next number.
- Do not edit `.specify/feature.json` by hand.
- Do not edit `meta.json` except for `status`, `services`, and `description`.
- Do not create more than one feature per invocation.
