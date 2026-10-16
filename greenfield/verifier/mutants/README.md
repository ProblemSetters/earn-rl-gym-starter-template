# goredis mutants

Each patch applies independently to a fresh copy of `reference-solution/`. Apply it from the copy's root with `git apply /path/to/verifier/mutants/<name>.patch`, then grade that copy with an external task runner.

| Patch | Deliberate defect | Target verifier check | Expected score |
| --- | --- | --- | ---: |
| `ping-wrong-reply.patch` | `PING` returns `+OK` | `ping` (15 points) | 85/100 |
| `get-empty-as-null.patch` | An existing empty value is returned as null | `empty-value` (5 points) | 95/100 |
| `set-no-overwrite.patch` | `SET` preserves an existing value | `overwrite` (10 points) | 90/100 |

The patches are intentionally absent from the candidate-visible materials. The unmutated reference must score 100/100, and each patched copy must still build and fail only its listed behavioral check.

The measured control and mutant results are in `verifications/<name>/1/reward.json`. Each result came from a separate fresh reference copy with at most one patch applied. Re-run these checks after changing the reference, verifier, or a patch.
