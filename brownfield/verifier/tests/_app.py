"""Start the candidate's Taskboard on a free port with a fresh SQLite file, for the test files next to this one."""

from __future__ import annotations

import contextlib
import json
import os
import socket
import sqlite3
import subprocess
import time
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

SEEDED = [(1, "Plan sprint review"), (2, "Update onboarding guide")]


class App:
    def __init__(self, base_url: str, database: Path, workspace: Path, env: dict):
        self.base_url, self.database, self.workspace, self.env = base_url, database, workspace, env

    def http(self, method: str, path: str, body=None):
        """One API call. Returns (status, headers, parsed JSON or None)."""
        data = json.dumps(body).encode() if body is not None else None
        request = Request(self.base_url + path, data=data, method=method,
                          headers={"Content-Type": "application/json"} if data is not None else {})
        try:
            with urlopen(request, timeout=5) as response:
                status, headers, raw = response.status, response.headers, response.read()
        except HTTPError as error:
            status, headers, raw = error.code, error.headers, error.read()
        try:
            return status, headers, json.loads(raw) if raw else None
        except ValueError:
            return status, headers, None

    def rows(self) -> list[tuple[int, str]]:
        with sqlite3.connect(self.database) as connection:
            return connection.execute("SELECT id, title FROM todos ORDER BY id").fetchall()


@contextlib.contextmanager
def running_app(workspace: Path, scratch: Path):
    with socket.socket() as probe:
        probe.bind(("127.0.0.1", 0))
        port = probe.getsockname()[1]
    database = scratch / "taskboard.sqlite"
    env = os.environ.copy()
    env.update({"PORT": str(port), "TASKBOARD_DB_PATH": str(database)})
    process = subprocess.Popen(["node", "--experimental-sqlite", "backend/src/server.js"], cwd=workspace, env=env,
                               stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        base_url = f"http://127.0.0.1:{port}"
        deadline = time.monotonic() + 10
        while True:
            if process.poll() is not None or time.monotonic() > deadline:
                raise RuntimeError("the app did not start")
            try:
                with urlopen(f"{base_url}/api/todos", timeout=0.5):
                    break
            except (OSError, URLError):
                time.sleep(0.1)
        yield App(base_url, database, workspace, env)
    finally:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


def run_browser(app: App, script: str | Path, timeout: int = 60) -> dict:
    """Run a Playwright script against the app. It prints one JSON object: {scenario: [passed, detail]}."""
    script_path = Path(script)
    if not script_path.is_absolute():
        script_path = Path(__file__).parent / script_path
    completed = subprocess.run(["node", str(script_path), app.base_url], cwd=app.workspace, env=app.env,
                               capture_output=True, text=True, timeout=timeout, check=False)
    try:
        return json.loads(completed.stdout.strip().splitlines()[-1])
    except (ValueError, IndexError):
        raise RuntimeError(f"{script} gave no result (exit {completed.returncode}): {completed.stderr[-500:]}") from None
