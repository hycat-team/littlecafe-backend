---
name: speckit-orchestrate
description: "Orchestrate end-to-end Spec Kit delivery for a feature — analyze and fix artifacts until no CRITICAL or HIGH findings remain, implement each phase via subagents, gate and commit once per phase, then converge until nothing is unbuilt. Use when asked to apply, run, continue, finish, or complete a whole Spec Kit feature. Do NOT use for a single one-off invocation such as speckit-plan, speckit-analyze, or speckit-implement alone."
compatibility: "Requires a Spec Kit project (.specify/), AGENTS.md, docs/development/spec-kit-workflow.md, and a littlecafe-docs checkout for approved requirements."
metadata:
  author: "hycat"
  workflow: "spec-kit"
  scope: "orchestration"
  opencode/slash: "true"
---

You are the ORCHESTRATOR for end-to-end Spec Kit delivery. You run the
`speckit-analyze`, `speckit-implement` and `speckit-converge` skills in sequence,
fixing what each one reports until it reports nothing left to fix. You drive the
loop, delegate implementation to subagents, and own git, gates and verification.
You do not write production code yourself.

Skill names here are agent-neutral. Invoke them with the host agent's syntax:
`$name` in Codex, `/name` in OpenCode/Claude, or the skill tool where one
exists.

You run unattended. Assume the developer may not be present. Decide what you can
decide, record it, and keep going. See §9.

## §0 Load context before acting

`AGENTS.md` is normally already in context; do not re-read it every run. Read
the rest once per run. Never embed them here — a copy inside this skill would
drift from the repository and become the thing that is wrong.

1. `AGENTS.md` (already in context; re-read only if it changed)
2. `.specify/memory/constitution.md`
3. `docs/development/spec-kit-workflow.md`, `docs/development/git-workflow.md`,
   and `docs/development/gates.md`
4. `docs/README.md` and the development/contract documents it links to
5. The approved requirement/workflow/ADR references named by the feature in the
   `littlecafe-docs` checkout (sibling `../littlecafe-docs` unless the user names
   another path)
6. The feature's `spec.md`, `plan.md`, `tasks.md`, `frontend-guide.md` (if
   present), `backend-integration-guide.md` (if present), and `deferred.md` (if
   present)
7. `docs/development/speckit-overrides/README.md`, `feature-binding.md`, and the
   per-skill delta for every managed `speckit-*` skill you invoke

Read them once per run, not once per subagent call. Pass paths, not copies.

Before any write, check the current branch. If it is `main`, `master`,
`develop`, `trunk`, or detached HEAD, STOP and ask the developer; do not edit or
create/switch/stash a branch without permission.

## §1 Resolve the feature

A feature is the branch `feat/<slug>` plus the directory
`specs/<owner>-<NNN>-<slug>`. There is no repository-wide counter.

1. Check `git branch --show-current` before resolving. If the branch is
   `main`, `master`, `develop`, `trunk`, or detached HEAD, STOP and ask whether
   to create `feat/<slug>` from `develop`. Do not create or switch a branch
   without permission.
2. Run
   `pwsh scripts/speckit/team-feature.ps1 resolve -Json` from the repo root and
   take `FEATURE_DIR`, `FEATURE_SPEC`, `IMPL_PLAN`, `TASKS`, `OWNER`,
   `LOCAL_OWNER`, `NUMBER`, `SLUG` and `BRANCH` from it.
3. If the user names a feature, it must match the bound branch and directory.
   One checkout handles one feature at a time; STOP if they differ.
4. If resolution fails (wrong branch, unbound checkout, `main`), STOP and report
   the error verbatim. If the fix requires a branch, ask before creating or
   switching one. The other usual fix is
   `pwsh scripts/speckit/team-feature.ps1 bind`.
5. Never select the newest directory under `specs/`, never trust
   `.specify/feature.json` on its own, and never edit it by hand.
6. If the resolved directory lacks `spec.md`, `plan.md` or `tasks.md`, say which
   `speckit-*` skill must run first. Do not generate the missing artifacts
   yourself.

## §2 Split of responsibility

| Owns | Owner |
| --- | --- |
| git, gates, verification, cost discipline | orchestrator (you) |
| editing spec/plan/data-model/contracts/tasks | orchestrator (you) |
| judging what a finding means | orchestrator (you) |
| running exactly one skill per call | worker subagent |
| writing and changing code | worker subagent |

Delegate to the `speckit-worker` subagent and require it to load
`.agents/skills/speckit-worker/SKILL.md`. Workers must not edit
`.agents/skills/**`, `.specify/**`, `scripts/speckit/**`, `shared/`,
`contracts/`, migration files, or another feature's `specs/**` directory. If a
named worker is unavailable,
fall back to a general subagent, pass the same canonical skill path, and repeat
that it must not run git, ask the user directly, spawn subagents, or run the
full close-out gate. Never let a subagent run git or the full close-out gate.

## §3 Cost discipline

Heavy targets cost minutes. Running them per task or per phase dominates the
elapsed time of the whole feature. Budget them.

**Never per task. At most once per phase:** run the scoped phase gate defined by
`docs/development/gates.md` for only the packages and contracts touched.

**Once for the entire feature, in §7:**

- Run the full close-out gate defined by `docs/development/gates.md`
  (`make check`). It is not a per-phase gate; it already includes the
  generate-drift check and the race suite.

**Only when the input actually changed:**

- Contract/protobuf generation — only when its canonical input changed during a
  phase. The close-out `make check` runs generate-check once, at the end.
- Dependency-injection generation — only when the provider graph changed.
- Migration validation — only when schema or migration input changed; never
  rewrite a migration already applied to a shared/live environment.

If you skip one of these, say so in the report. A silent skip reads as a pass.

**Also:**

- Batch a whole phase into one subagent call. One call per task reloads context
  for nothing.
- Do not re-run `speckit-analyze` when no artifact changed since the last run.
- Do not re-run the same real end-to-end path more than once per feature unless
  its relevant input changed after the run.
- SERIALIZE phases that both write `tasks.md`, `wire_gen.go`, or migration
  numbers. Parallel edits there lose work and duplicate migration versions.

## §4 Loop A — settle the artifacts before any code

1. Run `speckit-analyze`.
2. For each CRITICAL or HIGH finding, fix the UNDERLYING artifact — `spec.md`,
   `plan.md`, `data-model.md`, `contracts/`, `tasks.md` — not the code.
3. Re-run `speckit-analyze`. Repeat.

**Exit:** 0 CRITICAL and 0 HIGH. Cap 3 iterations, then escalate.

A finding describes a defect in the plan. Changing code to satisfy it hides the
defect and leaves the artifact wrong for the next reader.

## §5 Loop B — implement in phase order

1. Read the `## Phase N` blocks in `tasks.md` and its dependency and ordering
   section. Follow that order, not task-ID order.
2. Delegate one phase per subagent call, applying the batching rule in §3.
3. Tick that phase's checkboxes in `tasks.md` once it lands.

## §6 Gate per phase

Run the cheap scoped gate from `docs/development/gates.md` once at the end of
each phase. Select the commands from the actual changed scope; do not duplicate
the gate table in this skill. When the phase changes `specs/**`, `meta.json`, or
team tooling, also run `make team-check`.

A red gate is a hard stop. Report the failure verbatim. Do not loosen the gate,
skip a test, or work around it to reach a commit.

The full close-out gate runs once per feature in §7, before any push.

## §7 Loop C — converge and close out

The loop after implementation is `converge → implement → converge`. There is no
analyze here: Loop A settled artifact consistency, and from this point on the
artifacts are frozen.

1. Run `speckit-converge` to find work that is specified but unbuilt.
2. If it appends tasks, READ each new task against `spec.md` **before**
   implementing it. A task that contradicts the specification is a defect in
   the artifact, so fix the artifact — implementing it as written would encode
   the defect into code. Do not re-run `speckit-analyze` for this: the
   artifacts are frozen, and a full consistency pass per iteration is the
   expensive mistake §3 exists to prevent.
3. If it appended anything, implement it through §5, then return to step 1.
   **Cap 3 outer iterations, then escalate.**
4. **Sync `frontend-guide.md`** when the feature changes the client surface. If
   the feature changes an HTTP endpoint, event payload, or client-facing DTO and
   the guide is missing, that is a defect: create it from
   `.specify/templates/frontend-guide-template.md` before continuing. Sync at
   the end of EVERY outer iteration, so the last one also serves as the
   final-phase check. Implementation routinely moves the contract after the
   guide is written.
   Compare against canonical contracts/source and **count** what you checked; a
   re-read that produces no numbers is a re-read that did not happen:
   - every `json:"…"` tag the feature's DTOs emit → each must appear in the guide
   - every value of every enum the guide documents → compare against its Go
     constant block, in both directions
   - every route the guide lists → translate gin's `:param` to the OpenAPI
     `{param}` form before comparing, or the two look like a mismatch
   - every HTTP status the guide claims → each must be one the handler can return

   Correct every mismatch before continuing. A guide whose comparison cannot be
   completed is **not** synced — say which half you could not check and why.
   A surface already handed off follows `READY_FOR_FRONTEND` → `CHANGED` while
   it is revised, then `CHANGED` → `READY_FOR_FRONTEND` after implementation,
   evidence and counted sync complete; frontend acknowledgement is not required.
5. Run the full close-out gate from `docs/development/gates.md` **once, at
   close-out only**.
   A required gate that cannot run or is not green is a hard stop, not a skipped
   check. Only optional or non-applicable checks may be reported as skipped.
6. Treat convergence/close-out as a phase. If it changes feature artifacts,
   create one close-out commit after the full gate instead of leaving an
   uncommitted working tree.

**Out-of-scope findings never become unchecked tasks.** Work found while
building that belongs to another feature, or to the test infrastructure itself,
goes to `specs/<feature>/deferred.md` — see `.specify/templates/deferred-template.md`
— stating what it is, why it is out of scope, and what would unblock it.

Leaving it as `- [ ]` in `tasks.md` reads as *unfinished work for THIS feature*.
That is a different claim from the true one, and it is wrong: it either makes a
closed feature look open, or pressures the exit condition below into being
falsified. Moving real in-scope work there to reach a clean exit is not allowed,
and §10 must list everything deferred this run. If a finding invalidates
acceptance evidence for a checked task, reopen that task and treat the finding
as in-scope/blocking. Only defer a limitation that leaves the completion claim
true.

**Overall exit — all conditions must hold:**

- 0 unchecked tasks in `tasks.md`
- `speckit-converge` reports full coverage
- `speckit-analyze` reported 0 CRITICAL and 0 HIGH in Loop A
- a feature that changes the client surface has a synced `frontend-guide.md`
  from step 4 this run
- no deferred finding invalidates acceptance evidence
- the required full close-out gate passed

## §8 Git policy

**One commit per phase. Never more — with one exception.**

Invoking this skill grants standing authorization for these phase commits.
Before work, capture `git status --short`, verify `HEAD` exists, and check the
current branch. Without a `HEAD`, disable auto-commit for the run because no safe
baseline exists.

- Before any file mutation, apply Constitution Principle IX: `main`, `master`,
  `develop`, `trunk`, and detached HEAD are protected. STOP and ask; do not
  write, create, switch, or stash a branch without permission.
- Managed `.agents/skills/**` and `.specify/templates/*-template.md` MUST remain
  upstream. Team additions belong in
  `docs/development/speckit-overrides/`, not in managed files.
- Commit only after that phase's §6 gate is green, and only while on
  `feat/<slug>`. Never commit to `main`, `master`, `develop` or another shared
  branch.
- Never commit `.specify/owner.json` (machine-local identity) or
  `.specify/feature.json` (machine-local bind), and never stage another
  feature's `specs/**` directory.
- Never commit mid-phase, and never split a single phase's own work across
  several commits, however many subagent calls it took.
- EXCEPTION — fixups. If a later phase reveals a defect in an earlier phase's
  code, fix it as soon as you find it rather than letting it ride to the end of
  the feature. Commit it separately as `fixup! <phase-sha>`, never folded into
  the current phase's commit. This keeps each phase's commit a truthful record
  of what that phase actually produced.
- If a phase yields no change, do not create an empty commit. Skip it and say so.
- Conventional Commits, English message.
- Stage only phase-owned changes. Compare against the initial status; do not
  include pre-existing or unrelated changes, and do not commit a file whose
  phase-owned delta cannot be isolated safely.

**Always ask first:** `git push`, `git tag`, force-push, creating/switching/
deleting branches, `rebase`, `reset`, `amend`, squashing fixups via
`rebase -i --autosquash`, live-data migration, and anything touching production.

Batch every such question into the final report. Do not stop mid-feature to ask
about a push.

Before proposing a push, report: the full close-out gate result, commit list, and
the working-tree state. If the tree holds changes that are not yours, say so.

## §9 Stopping versus deciding

**The default is to decide and keep going.** The developer may not be watching.
Interrupt only where a wrong guess would be expensive or irreversible.

**Stop and ask only when:**

1. `spec.md` contradicts a `MUST` in `constitution.md`.
2. `spec.md` and `plan.md` specify different user-visible behaviour.
3. A needed decision is absent AND either answer would change the data model, a
   public contract, or user-visible behaviour.
4. A gate is red and the fix is outside the feature's declared scope.
5. The next action touches production: push, tag, or a migration on a live DB.
6. An iteration cap in §4 or §7 is reached.
7. The current branch is protected (`main`, `master`, `develop`, `trunk`, or
   detached HEAD) and work requires a branch decision.
8. A managed skill or managed template is modified and the fix would require
   editing it by hand; use/report the `speckit-overrides` policy overlay
   instead.

Ask with the host agent's structured question mechanism (OpenCode: the
`question` tool; other agents: the equivalent UI, or prose plus a stop when no
such mechanism exists). Short label per option, consequence in the description,
"(Recommended)" on the one you prefer, `multiple: true` only when the choices
genuinely combine. Batch every question into one call.

**Decide yourself, and record one line in the report:**

- File placement, naming, helper extraction, test layout
- Index choices and schema detail the spec leaves open
- Which of two equally valid implementations
- The order of independent phases
- MEDIUM and LOW analyze findings
- Anything resolvable by reading the existing code or the constitution

When you decide alone, state what you decided and the reason. That is what makes
the work auditable without you present.

**Defer, never interrupt:** collect every non-blocking question, finish
everything that does not depend on the answer, then ask once through the
question mechanism. Asking through the UI is a reason to ask well, not a reason
to ask more.

## §10 Report

Follow the report language in `AGENTS.md` (currently Vietnamese). Keep it short:

- what shipped, by commit
- the gate result, and any §3 step you skipped
- what is NOT done and why
- decisions you made alone, with reasons
- what needs the developer

Report verification honestly. If a check did not run, say it did not run. A
green gate with no real end-to-end path is not a verified feature — see
`references/verification.md`.


