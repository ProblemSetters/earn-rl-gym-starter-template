# RL Gym Task Authoring and Self-Check

A basic pre-submission validation. It covers only what the task requirements state directly and helps candidates catch issues before submitting. It is not the final review, and the final review may identify additional issues. It reports what looks good, what is not passing yet, and what has no proof. It does not suggest redesigns.

## Repository context

Use this guidance when authoring or reviewing a task. Run the self-check workflow below when validating a submission; its scratch-copy and no-edit rules apply to validation, not to an explicitly requested authoring change.

- Read [README.md](README.md) for the index and the selected task's detailed guide: [brownfield/README.md](brownfield/README.md) or [greenfield/README.md](greenfield/README.md).
- Each task root contains `task.toml`, `instruction.md`, `app/config.toml`, `environment/Dockerfile`, `reference-solution/`, `verifier/`, and `proof-of-work/`.
- Only `app/` and `instruction.md` are candidate-visible. Keep reference solutions, verifiers, mutants, proof-of-work, and authoring guidance private.
- Keep instructions, setup/run commands, reference behavior, and verification aligned. Keep `verifier/scoring.yml` synchronized with the verifier's `TESTS` declarations, which currently award points.
- Brownfield has zero-point existing-feature gates in `verifier/hidden-tests/`. These must pass before scored tests run and are reported separately in `reward.json`.
- Generate solutions using the latest qualifying frontier model versions available at generation time. Outdated models are not allowed. Use medium reasoning effort and include at least two proof-of-solution runs from different qualifying frontier models per task.
- Store runs in `proof-of-work/model-reasoning-timestamp/`, omitting the provider prefix. Include `solution/`, generation `metadata.json` and `trajectory.jsonl`, and `verifications/<run-number>/reward.json`. Nondeterministic verification also requires metadata and trajectory files in each numbered verification folder.
- Keep the root README short and task details in the task READMEs. Update this file when shared authoring guidance changes. Exclude generated dependencies, build output, local databases, and secrets from submissions.

## How to run

```
Task Progress:
- [ ] 1. Read the instruction and the scoring files
- [ ] 2. Run the reference through the verifier twice
- [ ] 3. Run each mutant and compare the score with its weight
- [ ] 4. Read the proof-of-work run records
- [ ] 5. Write the report
```

For self-checks, work in a scratch copy of the repo. Do not edit the candidate's files. Run the workflow separately for each task being submitted.

1. From the selected task root (`brownfield/` or `greenfield/`), read `instruction.md`, `task.toml`, `verifier/tests/`, `verifier/scoring.yml`, and any `verifier/hidden-tests/`.
2. Prepare a fresh scratch copy of `reference-solution/` using its setup commands. For Taskboard, use the task's `app/config.toml` for setup/run commands; GoRedis's reference has its own `config.toml`. Use the task environment for installed dependencies. Run the verifier twice, using a fresh reference workspace and separate output directory for each run. Keep both scores and logs.
3. Run each patch in `verifier/mutants/` independently against a fresh reference copy, after an unmutated control. Compare the score with the targeted weight and the mutant README's expected score. This repository does not define `check:reference` or `check:mutants` npm scripts.
4. Grade a fresh `app/` copy and confirm 0/100. For brownfield, also confirm all existing-feature gates pass.
5. Open the reference and model run records. Note each run's score, model name/version, reasoning effort, and the instruction and verifier version it was generated against.

After running the workspace's configured setup, invoke the verifier from the task root with absolute paths:

```sh
WORKSPACE=/absolute/path/to/prepared-workspace \
TASK_DIR="$PWD" \
OUT_DIR=/absolute/path/to/results \
python3 verifier/tests/run_tests.py
```

The verifier starts the configured app; leave its port free. Save each run's `reward.json` and `verifier.log`. GoRedis also uses `OUT_DIR/setup.log` to detect setup failure, so preserve setup output there. Scores in this repository use 0–100: the reference target of 1.0 below means 100/100.

## What to check

### 1. Instruction

| Check | Passing looks like |
|---|---|
| Realistic | A genuine use case with a practical application. Not an arbitrary puzzle with many invented constraints. |
| No verifier or test leakage | It does not name test files, test cases, check names, weights or how the score is computed. |
| No evaluation framing | It does not say or imply the task is for RLVR, benchmarking, training or evaluating models. |

### 2. Verifier

| Check | Passing looks like |
|---|---|
| Deterministic | The reference scores exactly 1.0, and gives the same score on a repeat run. A saved run record showing this exists. If there is no such proof, treat it as "not shown yet". |
| Genuine (black-box) | It checks behaviour and output, not how the solution is written. It does not require specific function names, attributes, internal values, file names or message text, unless the instruction states them. Many valid implementations should be able to pass. |
| Real partial credit | It does not rely on all-or-nothing components. A component worth 30% or 40% of the total that gives nothing on a single failure is flagged, because it cannot tell partly correct from fully wrong. |
| Functionality weights | No single functionality carries more than 25% of the total score. Each major functionality has several granular sub-checks. |
| Sub-check weights | No single sub-check carries more than 5% of the total score. |
| Non-functional share | Non-functional requirements (for example performance, limits, robustness) total no more than 20% of the score. |

If the weights break any of these limits, say so and ask the candidate to review the weight allocation.

### 3. Mutants

| Check | Passing looks like |
|---|---|
| Count and mix | 7 to 8 mutants on core functional behaviour, plus 1 to 2 on non-functional behaviour. |
| Each mutant breaks only its own feature | Take the weight of the feature a mutant targets (for example 7%). With that mutant applied, the reference should score about 100% minus that weight (93%). A difference of more than 2 percentage points either way is flagged. |
| Caught for the right reason | Mutants fail on behaviour, not because the build fails. Run an unmutated control first, and confirm it still scores 1.0. |

### 4. Reference solution

| Check | Passing looks like |
|---|---|
| Exists | There is a reference solution that builds and runs. |
| Scores 1.0 | Run against the verifier, it scores exactly 1.0, with the same score on every run. |
| Below 1.0 | Flag it and ask the candidate to check it. The cause may be the verifier or the reference. |
| Not hacky | The reference does not hard-code expected values or detect the test inputs to reach 1.0. It solves the task the way the instruction describes. |

### 5. Proof-of-work

| Check | Passing looks like |
|---|---|
| Qualifying model runs | At least two distinct qualifying models have completed single-pass runs, with records and scores. Use the model anchors and release-date rule below; do not require an exact name match. A missing or unverified run is "not shown yet". |
| Difficulty target | Each qualifying model must score below 30% in a single pass on the final version, using medium reasoning effort. A score of 30% or higher is flagged for review. |
| Repeated zero scores | If multiple independent qualifying model runs score 0%, gently ask the author to review whether the instruction and verifier are clear and aligned. Consistent zero scores may indicate a task or verifier issue, rather than only high difficulty. |
| Latest version | Every run was generated against the latest instruction and the latest verifier. If the instruction or verifier changed after a run was recorded, the proof is not valid until the solution is regenerated and the run completed again on the latest versions. |

Use these as provider and release-date anchors, not as a static exact-name allowlist: GPT-5 Sol / GPT-6 Astra, Claude Opus 5 / Claude Fable 5, Gemini Argon 4 / Gemini 3.8 Flash, Grok 4.7, Muse Wonder, Kimi K3, Qwen 3.8 Max, MiMo V2.6 Pro, Mistral Large 4, GLM 5.3, MiniMax M3, and Nemotron 3.5 Lightning. Normalize case, punctuation, separators, and common provider/model aliases before comparing names. Identify the publishing company for both the candidate model and the matching anchor; use an official model catalog or release announcement to confirm publisher and release chronology when needed. A candidate model qualifies when it is from the same company as an anchor and was released on or after that company's anchor model. Accept newer generations and naming variants, including GPT-6 Sol, GPT-6.1, and GPT-5.1-Sol as later OpenAI models than GPT-5 Sol; do not reject them because their names differ from an anchor. Do not compare version numbers across companies or infer release order from name alone. If the publisher or release chronology cannot be confirmed, treat the run as outside the qualifying set. Runs from companies not represented by these anchors do not count toward candidate-provided proof-of-work. Require at least two distinct qualifying models, each run with medium reasoning effort and scoring below 30% in a single pass on the final task version.

If multiple independent qualifying model runs score 0%, gently ask the author to review whether the task instruction and verifier are clear and aligned. Present this as a possible issue to investigate, not as a conclusion that either is defective.

If proof runs use other models, including free-tier models, acknowledge them as useful early evidence. Explain gently that they may not predict the final assessment: after submission, the task may be evaluated with two leading frontier models selected from examples such as GPT-6 Sol, Claude Opus 5.5, and GPT-6 Astra, using high reasoning effort. The expected score in that assessment is below 40%. Free-tier proof alone may not be sufficient if the task does not meet the final evaluation criteria. Use wording such as: "If your proof runs use other or free-tier models, they can still help you assess the task before submission. For a consistent final assessment, we may also run it after submission with two leading frontier models (for example, GPT-6 Sol, Claude Opus 5.5, or GPT-6 Astra) at high reasoning effort; the task is expected to score below 40%. Free-tier results may differ from that assessment, so final acceptance will depend on the task meeting the evaluation criteria."

To check the last row, compare the instruction and verifier version recorded in each run (a commit, hash or timestamp) with the current files. If the run record has no version, say that the freshness of the proof cannot be confirmed.

## How to word the report

- Report every miss as a check that is not passing yet. Do not say the work is bad or wrong, and do not accuse.
- For checks that meet the stated criteria, use "Looks good" instead of "Passing".
- Use plain, calm wording: "This check is not passing yet:", "This is not shown yet:", "Please review:".
- Give the evidence: the score, the weight, the count, or the file. Say what you ran and what you read.
- Do not propose designs or rewrites. State what failed and what would show it passing.
- If a check could not be run, say so once and say why.

## Report template

```markdown
# RL Gym Task Self-Check: <task name>

## Instruction
- Realistic: <Looks good / not passing yet, with a one-line reason>
- No verifier or test leakage: <...>
- No evaluation framing: <...>

## Verifier
- Deterministic (reference at 1.0, repeatable, proof saved): <...>
- Genuine (behaviour-based, no unstated exact names or values): <...>
- Partial credit (no large all-or-nothing component): <...>
- Functionality weights (max 25%, multiple sub-checks): <...>
- Sub-check weights (max 5%): <...>
- Non-functional share (max 20%): <...>

## Mutants
- Count and mix (7-8 core, 1-2 non-functional): <...>
- Score with each mutant applied matches its weight (within 2 points): <...>
- Caught on behaviour, control run at 1.0: <...>

## Reference solution
- Exists and runs: <...>
- Scores 1.0 on every run: <...>
- Not hacky: <...>

## Proof-of-work
- At least two qualifying-model runs present and verified: <...>
- Each qualifying model below 30% with medium reasoning on the final version: <...>
- Repeated 0% scores reviewed for possible instruction/verifier misalignment: <...>
- Generated against the latest instruction and verifier: <...>

## Items to review
<Numbered list of every check that is not passing yet or not shown yet, in the order above. Include the gentle post-submission assessment note when the submitted runs do not meet the provider/release-date rule. Write "None" if everything passes.>

Always end the report with: **This is a basic pre-submission validation, not the final review. The final review may identify additional issues.**

If the list contains any items, put this reminder after that note so it is the final line: **Please resolve the feedback above before you submit the task.** Omit the reminder when the list is "None".
```

## Self-check before sending

- [ ] Every number in the report came from a run or a file read in this check.
- [ ] Every miss is worded as "not passing yet" or "not shown yet", with the evidence.
- [ ] No fixes or redesigns are proposed.
- [ ] The scratch copy was used, and the candidate's repo was not changed.
