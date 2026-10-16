# Greenfield task

A greenfield task asks a candidate to build an application from a minimal starting scaffold.

```text
greenfield/
├── task.toml
├── instruction.md
├── app/
├── environment/
│   └── Dockerfile
├── reference-solution/
├── verifier/
│   ├── tests/
│   ├── scoring.yml
│   └── mutants/
└── proof-of-work/
```

## Components

- `task.toml`: Defines task metadata, limits, and verification settings.
- `instruction.md`: States the requirements given to the candidate.
- `app/`: Contains the starting scaffold and its `config.toml` setup and run commands.
- `environment/Dockerfile`: Defines the dependencies used to run and verify the task.
- `reference-solution/`: Contains a complete implementation expected to score 100%.
- `verifier/`: Holds the checks and scoring materials used to grade behavior.
- `verifier/tests/`: Holds the scored behavioral checks.
- `verifier/scoring.yml`: Lists the checks and their weights.
- `verifier/mutants/`: Holds independent one-defect patches and their verification records.
- `proof-of-work/`: Holds candidate solutions and their run records.

## Records

- Only `instruction.md` and `app/` are candidate-visible.
- A verifier grades behavior and writes `reward.json`.
- Each mutant is applied to a fresh reference copy; its result is stored at `verifier/mutants/verifications/<name>/<run>/reward.json`, alongside an unmutated `control` result.
- A proof-of-work run at `proof-of-work/<model>-<reasoning>-<timestamp>/` contains `solution/`, `metadata.json`, `trajectory.jsonl`, and `verifications/<run>/reward.json`.
- `metadata.json` records the run's provenance; `trajectory.jsonl` records its steps.
