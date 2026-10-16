# Taskboard mutants

Each patch applies independently to a fresh copy of `reference-solution/`. Apply it from the copy's root with `git apply /path/to/verifier/mutants/<name>.patch`, then grade that copy with an external task runner.

| Patch | Deliberate defect | Target verifier check | Expected score |
| --- | --- | --- | ---: |
| `delete-wrong-success-body.patch` | Successful DELETE reports `{deleted: false}` | `api-delete-existing` (15 points) | 85/100 |
| `delete-missing-returns-ok.patch` | DELETE of a missing task returns 200 | `api-delete-missing-404` (10 points) | 90/100 |
| `delete-swallows-failure.patch` | A failed DELETE leaves the task visible without showing an error | `delete-failure-keeps-item` (15 points) | 85/100 |

The patches stay in `verifier/`, outside the candidate-visible materials. The unmutated reference must score 100/100, and each patched copy must still start and fail only its listed behavioral check.

The measured control and mutant results are in `verifications/<name>/1/reward.json`. Each result came from a separate fresh reference copy with at most one patch applied. Re-run these checks after changing the reference, verifier, or a patch.
