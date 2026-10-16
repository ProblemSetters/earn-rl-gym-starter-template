#!/usr/bin/env python3
"""Black-box RESP tests for goredis."""

from __future__ import annotations

import json
import os
import signal
import socket
import subprocess
import time
import tomllib
from pathlib import Path

HERE = Path(__file__).resolve().parent
PORT = 6379
TESTS = [
    ("starts", "config.toml sets the project up and the server starts listening on port 6379", 10),
    ("ping", "PING returns the RESP simple string PONG", 15),
    ("set-get", "SET stores and GET returns a bulk string", 15),
    ("overwrite", "SET replaces an existing value", 10),
    ("empty-value", "GET returns an empty stored value as an empty bulk string", 5),
    ("missing-get", "GET of a missing key returns a null bulk string", 10),
    ("del", "DEL counts existing distinct keys and removes them", 20),
    ("arity-errors", "invalid command arity returns an ERR and keeps the connection usable", 10),
    ("multiple-clients", "the server serves separate client connections", 5),
]


def encode(*args: str) -> bytes:
    parts = [f"*{len(args)}\r\n".encode()]
    for arg in args:
        raw = arg.encode()
        parts.extend((f"${len(raw)}\r\n".encode(), raw, b"\r\n"))
    return b"".join(parts)


def read_line(sock: socket.socket) -> bytes:
    data = bytearray()
    while not data.endswith(b"\r\n"):
        chunk = sock.recv(1)
        if not chunk:
            raise EOFError("server closed the connection")
        data.extend(chunk)
    return bytes(data[:-2])


def read_reply(sock: socket.socket):
    prefix = sock.recv(1)
    if prefix == b"+":
        return ("simple", read_line(sock).decode())
    if prefix == b"-":
        return ("error", read_line(sock).decode())
    if prefix == b":":
        return ("integer", int(read_line(sock)))
    if prefix == b"$":
        length = int(read_line(sock))
        if length == -1:
            return ("bulk", None)
        chunks = bytearray()
        while len(chunks) < length + 2:
            chunk = sock.recv(length + 2 - len(chunks))
            if not chunk:
                raise EOFError("server closed during bulk reply")
            chunks.extend(chunk)
        if chunks[-2:] != b"\r\n":
            raise ValueError("bad bulk string terminator")
        return ("bulk", bytes(chunks[:-2]).decode())
    raise ValueError(f"invalid RESP reply prefix {prefix!r}")


def request(sock: socket.socket, *args: str):
    sock.sendall(encode(*args))
    return read_reply(sock)


def main() -> int:
    workspace = Path(os.environ.get("WORKSPACE", "/work/candidate"))
    out = Path(os.environ.get("OUT_DIR", "/logs/verifier"))
    task_dir = Path(os.environ.get("TASK_DIR", HERE.parents[1]))
    metadata = tomllib.loads((task_dir / "task.toml").read_text())["metadata"]
    out.mkdir(parents=True, exist_ok=True)
    results = {}
    details = {}

    def check(test_id: str, passed: bool, detail: str = ""):
        results[test_id] = passed
        if not passed:
            details[test_id] = detail or "behavior did not match the RESP contract"

    if not workspace.is_dir():
        results = {"verifier-setup": False}
        tests = [{"id": "verifier-setup", "title": "the verifier has a candidate workspace",
                  "score": 0, "max_score": 1, "status": "runtime_error", "type": "deterministic"}]
    else:
        later = [test_id for test_id, _, _ in TESTS if test_id != "starts"]
        for test_id in later:
            check(test_id, False, "the server did not start")

        # The candidate says how to start its server in config.toml. The runner has already run its [setup] commands
        # and left their output in setup.log; a failed or skipped setup is a failed candidate, not a verifier error.
        problem = ""
        command = ""
        setup_log = out / "setup.log"
        log_text = setup_log.read_text(errors="replace") if setup_log.exists() else ""
        if "setup skipped:" in log_text or "setup failed" in log_text:
            problem = log_text[-500:].strip()
        else:
            try:
                command = tomllib.loads((workspace / "config.toml").read_text())["run"]["command"].strip()
            except (OSError, KeyError, TypeError, AttributeError, tomllib.TOMLDecodeError) as error:
                problem = f"config.toml has no usable [run].command: {error}"
            if not command and not problem:
                problem = "config.toml [run].command is empty"

        proc = None
        client = None
        if not problem:
            proc = subprocess.Popen(command, shell=True, cwd=workspace, stdout=subprocess.DEVNULL,
                                    stderr=subprocess.DEVNULL, start_new_session=True)
            deadline = time.monotonic() + 10
            while time.monotonic() < deadline:
                if proc.poll() is not None:
                    problem = f"the start command exited with {proc.returncode}"
                    break
                try:
                    client = socket.create_connection(("127.0.0.1", PORT), timeout=0.5)
                    break
                except OSError:
                    time.sleep(0.05)
            if client is None and not problem:
                problem = f"nothing listened on 127.0.0.1:{PORT} within 10s"
        check("starts", client is not None, problem)

        if client is not None:
            for test_id in later:
                results.pop(test_id)
                details.pop(test_id, None)
                results[test_id] = False
                details[test_id] = "did not run"
            try:
                client.settimeout(2)
                with client:
                    check("ping", request(client, "PING") == ("simple", "PONG"))
                    set_reply = request(client, "SET", "alpha", "one")
                    get_reply = request(client, "GET", "alpha")
                    check("set-get", set_reply == ("simple", "OK") and get_reply == ("bulk", "one"),
                          f"SET={set_reply!r}, GET={get_reply!r}")
                    request(client, "SET", "alpha", "two")
                    check("overwrite", request(client, "GET", "alpha") == ("bulk", "two"))
                    empty_set = request(client, "SET", "empty-value-key", "")
                    empty_get = request(client, "GET", "empty-value-key")
                    check("empty-value", empty_set == ("simple", "OK") and empty_get == ("bulk", ""),
                          f"SET={empty_set!r}, GET={empty_get!r}")
                    check("missing-get", request(client, "GET", "missing") == ("bulk", None))
                    request(client, "SET", "beta", "b")
                    deleted = request(client, "DEL", "alpha", "missing", "alpha", "beta")
                    after = (request(client, "GET", "alpha"), request(client, "GET", "beta"))
                    missing_del = request(client, "DEL", "absent")
                    check("del", deleted == ("integer", 2) and after == (("bulk", None), ("bulk", None))
                          and missing_del == ("integer", 0),
                          f"DEL={deleted!r}, remaining={after!r}, missing DEL={missing_del!r}")
                    bad_ping = request(client, "PING", "extra")
                    no_key_del = request(client, "DEL")
                    after_error = request(client, "GET", "missing")
                    check("arity-errors", bad_ping[0] == "error" and bad_ping[1].startswith("ERR ")
                          and no_key_del[0] == "error" and no_key_del[1].startswith("ERR ")
                          and after_error == ("bulk", None),
                          f"PING={bad_ping!r}, DEL={no_key_del!r}, later GET={after_error!r}")
                    request(client, "SET", "cross-client-key", "shared")
                with socket.create_connection(("127.0.0.1", PORT), timeout=2) as second:
                    second.settimeout(2)
                    check("multiple-clients", request(second, "GET", "cross-client-key") == ("bulk", "shared"))
            except (OSError, EOFError, ValueError, ConnectionError) as error:
                for test_id in later:
                    if results[test_id] is False and details.get(test_id) == "did not run":
                        details[test_id] = str(error)
        if proc is not None:
            try:
                os.killpg(proc.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
            try:
                proc.wait(timeout=3)
            except subprocess.TimeoutExpired:
                os.killpg(proc.pid, signal.SIGKILL)
                proc.wait()

        tests = []
        for test_id, title, points in TESTS:
            passed = results.get(test_id, False)
            tests.append({"id": test_id, "title": title, "score": points if passed else 0,
                          "max_score": points, "status": "passed" if passed else "failed",
                          "type": "deterministic"})

    log = [f"FAIL {test_id}: {message}" for test_id, message in details.items()]
    if "verifier-setup" in results:
        log.append("RUNTIME_ERROR candidate workspace missing")
    result = {"schema_version": 1, "task": metadata["id"], "task_version": metadata["version"],
              "score": sum(test["score"] for test in tests), "max_score": sum(test["max_score"] for test in tests),
              "tests": tests}
    (out / "reward.json").write_text(json.dumps(result, indent=2) + "\n")
    (out / "verifier.log").write_text("\n".join(log) + ("\n" if log else ""))
    print(f"[verifier] score={result['score']}/{result['max_score']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
