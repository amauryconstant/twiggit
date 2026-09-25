# Prompt: run the golang-cli evals

Paste everything below the line into a fresh agent session started at the repository root (`/home/amaury/Projects/go-cli-skill`).

---

You are orchestrating an adversarial evaluation of the agent skill at `skills/golang-cli/`. The skill is a drop-in replacement for `samber/cc-skills-golang@golang-cli`. Your job is to measure how much the skill improves answers, which is the pass rate with the skill minus the pass rate without it. Work through the phases in order. Each phase ends on a checkable condition; do not start the next phase until it holds.

## Inputs

- Eval suite: `skills/golang-cli/evals/evals.json`, with 62 evals and 300 assertions. Each eval has `id`, `name`, `description`, `prompt`, `trap`, and `assertions[{id, text}]`.
- Skill under test: `skills/golang-cli/SKILL.md`, plus the `references/` and `assets/examples/` files it links to.
- Report format reference: `/home/amaury/Projects/_external/cc-skills-golang/EVALUATIONS.md`, the section `golang-cli — v1.0.0` and the file's `<style>` block. The rules behind it are in `/home/amaury/Projects/_external/cc-skills-golang/CLAUDE.md` under "Evaluation".
- Workspace: `/tmp/golang-cli-workspace/`. Every intermediate file goes there.

## Rules that protect the measurement

1. **Executors are blind.** An executor receives only the eval's `prompt`, plus the skill instruction in the with-skill condition. It never sees `trap`, `assertions`, `description`, `evals.json`, this file, or the other condition's output.
2. **Isolate the baseline.** The without-skill condition must have no access to this skill or any overlapping Go/CLI skill:
   - Check that no `golang-cli` skill is installed in `~/.claude/skills/` or `./.claude/skills/`.
   - Check that no cc-skills-golang plugin is enabled in `~/.claude/settings.json` or `./.claude/settings.json`.
   - Run each executor with its working directory set to an empty folder, `/tmp/golang-cli-workspace/runs/<id>/<condition>/`.
   - Tell the executor not to read anything under `/home/amaury/Projects/`.
3. **Load only the skill under test.** The with-skill executor may read only files under `skills/golang-cli/`, and never `skills/golang-cli/evals/`. Pass the skill as an absolute path.
4. **Keep the conditions identical.** Use the same model, the same tools, and the same prompt wording, except for the one skill instruction. If session hooks or output styles are active (for example caveman or ponytail modes), they apply to both conditions equally. Record which ones were active.
5. **Leave the skill and the evals unchanged.** This run measures; it does not fix anything.

## Phase 1 — Preflight

- Parse `evals.json`. Confirm 62 evals and 300 assertions, and that every assertion id starts with `<eval id>.`.
- Run the isolation checks from rule 2 and record what you found.
- Record the executor model id, the judge model id, the date, the skill's `metadata.version` from the SKILL.md frontmatter, and the git commit (`git rev-parse --short HEAD`).
- Write all of this to `/tmp/golang-cli-workspace/run.json`.

**Done when** `run.json` exists and every isolation check passed. If a check fails, stop and report it; do not work around it.

## Phase 2 — Execute (124 runs)

For every eval, launch two executor subagents, one per condition. Run them in parallel batches of at most 10.

Without-skill executor prompt:

```text
You are a senior Go engineer. Answer the request below completely, including all code in fenced blocks. Do not read files under /home/amaury/Projects/. Do not ask clarifying questions; make reasonable assumptions and state them.

<request>
{prompt}
</request>
```

With-skill executor prompt (the same, plus one instruction):

```text
You are a senior Go engineer. Before answering, read /home/amaury/Projects/go-cli-skill/skills/golang-cli/SKILL.md and follow it; read the references and asset files it links to when they bear on the request. Read nothing else under /home/amaury/Projects/, and never read the evals/ directory. Answer the request below completely, including all code in fenced blocks. Do not ask clarifying questions; make reasonable assumptions and state them.

<request>
{prompt}
</request>
```

For each run:

- Save the executor's final answer verbatim to `/tmp/golang-cli-workspace/runs/<id>/<condition>/answer.md`, where `<condition>` is `with` or `without`.
- For with-skill runs, also save which skill files the executor read to `files_read.txt`. That shows which references are actually used.

**Done when** all 124 `answer.md` files exist and none is empty. Re-run any failed or truncated run once, and log any run that fails twice in `run.json`.

## Phase 3 — Grade (62 judge runs)

For every eval, launch one judge subagent. Give it:

- the eval's `prompt` and `assertions`, but not the `trap`
- the two answers, labeled `Answer A` and `Answer B`, in random order. Record the mapping in `/tmp/golang-cli-workspace/runs/<id>/mapping.json`; the judge never learns which is which.

Judge instructions:

```text
Grade each answer against each assertion independently. An assertion passes only if the answer clearly does what it states, in code or in explicit guidance; implied, partial, or "could be added" does not pass. Where an assertion offers alternatives ("X or Y"), either satisfies it. For every verdict give at most 12 words of evidence quoting or pointing at the answer. Return only JSON:
{"A": {"<assertion id>": {"pass": true|false, "evidence": "..."}}, "B": {...}}
```

Save each judge's output to `/tmp/golang-cli-workspace/runs/<id>/grades.json`, and map A/B back to with/without.

**Done when** every eval has one verdict per assertion per condition, 600 verdicts in total. Re-run any judge whose JSON fails to parse or has missing ids.

## Phase 4 — Aggregate and flag

Compute:

- overall with/without pass counts and percentages
- the delta in percentage points
- the uplift (with ÷ without, to 2 decimal places, with a `×` suffix)
- the same numbers per eval

Save everything to `/tmp/golang-cli-workspace/results.json`.

Flag each eval group using the upstream anti-pattern rules:

| Pattern | Meaning | Suggested action |
| --- | --- | --- |
| with = without = 100% | Tests common knowledge | Redesign or cut the eval |
| with = without = 0% | Coverage gap in the skill, or an eval that asks for the wrong approach | Check the prompt for an explicit wrong instruction; otherwise the skill needs the rule |
| with = without, partial | Mixes common-knowledge and gap assertions | Split the group |
| with < without | The skill hurts | Read both answers; name the skill line that misled |

Also list every assertion that fails with the skill, with the judge's evidence. These are the skill's blind spots.

**Done when** `results.json` holds the totals, the per-eval scores, the flags, and the list of failures with the skill.

## Phase 5 — Report

1. **Create `EVALUATIONS.md`** at the repository root.
   - Wrap the file in `<!-- prettier-ignore-start -->` / `<!-- prettier-ignore-end -->`.
   - Copy the upstream `<style>` block, with the `.g` and `.r` classes.
   - Add a one-row summary table: skill, version, assertions, with %, without %, delta, uplift.
   - Add a `## \`golang-cli\` — v<version>` section in the upstream format:
     - an overall table
     - a `<details>` block with the model, and "Grading: LLM-as-judge (blind, randomized A/B)"
     - a per-eval header row, with bold name, description and colored score spans
     - `a.b` assertion rows with ✓/✗ spans, plus short evidence after each ✗
2. **Add a section at the end** called `### Findings` with:
   - the flagged groups and suggested actions from Phase 4
   - the 10 largest per-eval deltas
   - the references that with-skill executors never read (from the `files_read.txt` files), which are candidates for pruning
3. **Run the markdown checks:** `npx prettier --check EVALUATIONS.md` and `npx markdownlint-cli2@0.23.3 EVALUATIONS.md`.
4. **Do not commit.** Report the headline numbers, the flagged groups, and the path to `EVALUATIONS.md`.

**Done when** `EVALUATIONS.md` passes both checks and every one of the 62 evals appears in it with its with and without scores.

## Cost note

The run launches about 186 subagents: 124 executors and 62 judges. If you must cut cost, keep the complete set of assertions for fewer evals rather than grading every eval partially. Evals 1–12 are the upstream baseline, so run those first.

For partial reruns, `tools/evals/{execute,grade,compare}.py` run Phases 2–4 headlessly with the same prompts: `--ws <dir> --ids 8,9,12 --reps 3`, and `compare.py --old <earlier workspaces>` for baselines. `evals/trigger-evals.json` holds the should/shouldn't-trigger queries for skill-creator's description optimizer.
