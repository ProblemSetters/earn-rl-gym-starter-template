#!/usr/bin/env python3
"""Run zero-point hidden gates before the scored browser and API checks."""

from __future__ import annotations

import importlib.util
import json
import os
import sys
import tempfile
import tomllib
from pathlib import Path

HERE = Path(__file__).resolve().parent
HIDDEN = HERE.parent / "hidden-tests"
sys.path.insert(0, str(HERE))
NOT_TEST_FILES = {"run_tests.py"}


def load_modules(directory: Path) -> list:
    modules = []
    for path in sorted(directory.glob("*.py")):
        if path.name in NOT_TEST_FILES or path.name.startswith("_"):
            continue
        spec = importlib.util.spec_from_file_location(path.stem, path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        for name in ("TESTS", "run"):
            if not hasattr(module, name):
                raise ValueError(f"{path.name} must define {name}")
        modules.append(module)
    if not modules:
        raise ValueError(f"no test files found in {directory}")
    return modules


def run_module(module, workspace: Path) -> tuple[dict, list[str]]:
    reported: dict[str, tuple[bool, str]] = {}
    notes: list[str] = []

    def check(test_id: str, passed: bool, evidence: object = "") -> None:
        if test_id not in module.TESTS:
            raise KeyError(f"{test_id!r} is not declared in TESTS")
        reported[test_id] = (bool(passed), str(evidence))

    with tempfile.TemporaryDirectory() as scratch:
        try:
            module.run(workspace, Path(scratch), check)
        except Exception as error:  # noqa: BLE001 - a crash while exercising the candidate fails the unreported tests
            notes.append(f"CRASH {module.__name__}: {type(error).__name__}: {error}")
    return reported, notes


def main() -> int:
    workspace = Path(os.environ.get("WORKSPACE", "/work/candidate"))
    out = Path(os.environ.get("OUT_DIR", "/logs/verifier"))
    task_dir = Path(os.environ.get("TASK_DIR", HERE.parents[1]))
    metadata = tomllib.loads((task_dir / "task.toml").read_text(encoding="utf-8"))["metadata"]
    out.mkdir(parents=True, exist_ok=True)
    log: list[str] = []

    def finish(tests: list[dict], hidden_tests: list[dict]) -> int:
        result = {
            "schema_version": 1, "task": metadata["id"], "task_version": metadata["version"],
            "score": sum(item["score"] for item in tests), "max_score": sum(item["max_score"] for item in tests),
            "tests": tests,
            "hidden-tests": hidden_tests,
        }
        (out / "reward.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
        (out / "verifier.log").write_text("\n".join(log) + ("\n" if log else ""), encoding="utf-8")
        print(f"[verifier] score={result['score']}/{result['max_score']}")
        return 0

    def verifier_error(message: str) -> int:
        log.append(f"RUNTIME_ERROR {message}")
        return finish([{"id": "verifier-setup", "title": "the verifier could run", "score": 0, "max_score": 1,
                        "status": "runtime_error", "type": "deterministic"}], [])

    if not workspace.is_dir():
        return verifier_error(f"workspace not found: {workspace}")
    try:
        hidden_modules = load_modules(HIDDEN)
        scored_modules = load_modules(HERE)
        if any(not getattr(module, "GATED", False) for module in hidden_modules):
            raise ValueError("every hidden test module must declare GATED = True")
        all_ids = [test_id for module in hidden_modules + scored_modules for test_id in module.TESTS]
        if len(all_ids) != len(set(all_ids)):
            raise ValueError("a test id is declared more than once")
    except Exception as error:  # noqa: BLE001 - a broken test file is the verifier's fault, not the candidate's
        return verifier_error(f"could not load the test files: {type(error).__name__}: {error}")

    tests, hidden_tests = [], []
    hidden_passed = True
    for module in hidden_modules:
        reported, notes = run_module(module, workspace)
        log.extend(notes)
        for test_id, (title, _) in module.TESTS.items():
            passed, evidence = reported.get(test_id, (False, "the test was not reached"))
            if not passed:
                log.append(f"FAIL {test_id}: {evidence}")
                hidden_passed = False
            hidden_tests.append({"id": test_id, "title": title, "max_score": 0, "score": 0,
                                 "status": "passed" if passed else "failed", "type": "deterministic"})

    for module in scored_modules:
        reported, notes = run_module(module, workspace) if hidden_passed else ({}, [])
        log.extend(notes)
        for test_id, (title, points) in module.TESTS.items():
            entry = {"id": test_id, "title": title, "max_score": points, "type": "deterministic"}
            if not hidden_passed:
                log.append(f"SKIPPED {test_id}: a hidden gate failed")
                tests.append({**entry, "score": 0, "status": "skipped"})
                continue
            passed, evidence = reported.get(test_id, (False, "the test was not reached"))
            if not passed:
                log.append(f"FAIL {test_id}: {evidence}")
            tests.append({**entry, "score": points if passed else 0, "status": "passed" if passed else "failed"})
    return finish(tests, hidden_tests)


if __name__ == "__main__":
    raise SystemExit(main())
