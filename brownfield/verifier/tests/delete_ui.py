"""Mandatory: deleting a task through the real UI, then the SQLite file."""

from __future__ import annotations

from pathlib import Path

from _app import run_browser, running_app

TESTS = {
    "delete-button-removes-item": ("Delete removes the selected task from the list", 15),
    "delete-removes-sqlite-row": ("the selected task is deleted from SQLite", 20),
    "delete-removes-only-that-item": ("deleting one task removes only that task and leaves the other intact", 10),
    "delete-persists-after-reload": ("a deleted task stays gone after the page is reloaded", 15),
    "delete-failure-keeps-item": ("when the DELETE request fails the task stays and an error is shown", 15),
}


def run(workspace: Path, scratch: Path, check):
    with running_app(workspace, scratch) as app:
        results = run_browser(app, "delete_ui.cjs")
        rows = app.rows() if app.database.exists() else None

    def outcome(name):
        return results.get(name, (False, "the scenario did not run"))

    check("delete-button-removes-item", *outcome("removed"))
    check("delete-persists-after-reload", *outcome("reload"))
    check("delete-failure-keeps-item", *outcome("failure"))
    if rows is None:
        for test_id in ("delete-removes-sqlite-row", "delete-removes-only-that-item"):
            check(test_id, False, "the app did not create its SQLite database")
        return
    target = [row for row in rows if row[0] == 1 or row[1] == "Plan sprint review"]
    other = [row for row in rows if row == (2, "Update onboarding guide")]
    check("delete-removes-sqlite-row", not target, f"expected no row for Plan sprint review; found {target}")
    # The other task is always still there in an app that deletes nothing, so this only passes once the
    # selected task is really gone as well.
    check("delete-removes-only-that-item", not target and len(other) == 1,
          f"expected Plan sprint review gone and Update onboarding guide kept; rows are {rows}")
