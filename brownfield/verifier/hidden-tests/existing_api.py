"""API behaviour that already works, checked independently of deletion."""

from __future__ import annotations

from pathlib import Path

from _app import SEEDED, running_app

TESTS = {
    "api-lists-todos": ("GET /api/todos lists the seeded tasks in id order", 2),
    "api-gets-one-todo": ("GET /api/todos/:id returns one task, or 404", 2),
    "api-creates-todo": ("POST /api/todos creates a trimmed task and answers 201 with a Location", 2),
    "api-rejects-invalid-input": ("POST with a blank title, or a malformed id, answers 400", 1),
    "api-updates-todo": ("PATCH /api/todos/:id renames a task, or answers 404", 2),
}
GATED = True


def run(workspace: Path, scratch: Path, check):
    with running_app(workspace, scratch) as app:
        seeded = [{"id": i, "title": t} for i, t in SEEDED]
        status, _, body = app.http("GET", "/api/todos")
        check("api-lists-todos", status == 200 and body == seeded, f"GET={status} {body}")

        one, _, body = app.http("GET", "/api/todos/1")
        missing, _, _ = app.http("GET", "/api/todos/999")
        check("api-gets-one-todo", one == 200 and body == seeded[0] and missing == 404, f"GET /1={one} {body}, GET /999={missing}")

        status, headers, body = app.http("POST", "/api/todos", {"title": "  Ship release  "})
        listed = app.http("GET", "/api/todos")[2]
        check("api-creates-todo",
              status == 201 and body == {"id": 3, "title": "Ship release"} and headers.get("Location") == "/api/todos/3"
              and listed[-1] == body,
              f"POST={status} {body} Location={headers.get('Location')}")

        blank = app.http("POST", "/api/todos", {"title": "   "})[0]
        bad_id = app.http("GET", "/api/todos/abc")[0]
        check("api-rejects-invalid-input", blank == 400 and bad_id == 400, f"blank title={blank}, id 'abc'={bad_id}")

        status, _, body = app.http("PATCH", "/api/todos/1", {"title": "Renamed review"})
        unknown = app.http("PATCH", "/api/todos/999", {"title": "x"})[0]
        check("api-updates-todo",
              status == 200 and body == {"id": 1, "title": "Renamed review"} and unknown == 404
              and app.http("GET", "/api/todos/1")[2] == body,
              f"PATCH={status} {body}, unknown id={unknown}")
