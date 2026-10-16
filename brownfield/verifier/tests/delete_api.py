"""Mandatory: the DELETE /api/todos/:id endpoint."""

from __future__ import annotations

from pathlib import Path

from _app import running_app

TESTS = {
    "api-delete-existing": ("DELETE of an existing task answers 200 {deleted: true} and removes it", 15),
    "api-delete-missing-404": ("DELETE of a task that does not exist, or twice, answers 404", 10),
}


def run(workspace: Path, scratch: Path, check):
    with running_app(workspace, scratch) as app:
        status, _, body = app.http("DELETE", "/api/todos/1")
        _, _, after = app.http("GET", "/api/todos")
        gone_status, _, _ = app.http("GET", "/api/todos/1")
        check("api-delete-existing",
              status == 200 and body == {"deleted": True} and gone_status == 404 and after == [{"id": 2, "title": "Update onboarding guide"}],
              f"DELETE={status} {body}, GET after={after}, GET /1={gone_status}")
        again, _, _ = app.http("DELETE", "/api/todos/1")
        unknown, _, _ = app.http("DELETE", "/api/todos/999")
        check("api-delete-missing-404", again == 404 and unknown == 404, f"second DELETE={again}, unknown id={unknown}")
