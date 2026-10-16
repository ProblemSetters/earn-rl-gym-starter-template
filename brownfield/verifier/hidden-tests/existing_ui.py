"""The existing list and add flow, checked independently of deletion."""

from __future__ import annotations

from pathlib import Path

from _app import run_browser, running_app

TESTS = {
    "ui-lists-and-adds-tasks": ("the UI lists the tasks and adds a new one", 1),
}
GATED = True


def run(workspace: Path, scratch: Path, check):
    with running_app(workspace, scratch) as app:
        results = run_browser(app, Path(__file__).with_suffix(".cjs"))
    check("ui-lists-and-adds-tasks", *results.get("listAdd", (False, "the scenario did not run")))
