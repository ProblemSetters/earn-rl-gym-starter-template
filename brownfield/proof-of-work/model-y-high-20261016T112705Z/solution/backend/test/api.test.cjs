const assert = require('node:assert/strict');
const { test } = require('node:test');
const { spawn } = require('node:child_process');
const { mkdtempSync, rmSync } = require('node:fs');
const { tmpdir } = require('node:os');
const path = require('node:path');
const net = require('node:net');
const { once } = require('node:events');

test('task API persists changes, validates input, and supports deletion', async () => {
  const socket = net.createServer().listen(0, '127.0.0.1');
  await once(socket, 'listening');
  const port = socket.address().port;
  await new Promise(resolve => socket.close(resolve));
  const scratch = mkdtempSync(path.join(tmpdir(), 'taskboard-test-'));
  const child = spawn(process.execPath, ['--experimental-sqlite', 'backend/src/server.js'], {
    cwd: path.resolve(__dirname, '../..'),
    env: { ...process.env, PORT: String(port), TASKBOARD_DB_PATH: path.join(scratch, 'test.sqlite') },
    stdio: 'pipe',
  });
  let logs = '';
  child.stderr.on('data', chunk => { logs += chunk; });
  const request = (route, options) => fetch(`http://127.0.0.1:${port}/api${route}`, options);
  try {
    let ready = false;
    for (let attempt = 0; attempt < 100; attempt++) {
      try { ready = (await request('/todos')).ok; } catch {}
      if (ready) break;
      if (child.exitCode !== null) throw new Error(logs);
      await new Promise(resolve => setTimeout(resolve, 50));
    }
    assert.ok(ready, `server did not start: ${logs}`);
    assert.deepEqual(await (await request('/todos/1')).json(), { id: 1, title: 'Plan sprint review' });
    const json = (method, body) => ({ method, headers: { 'content-type': 'application/json' }, body: JSON.stringify(body) });
    assert.equal((await request('/todos', json('POST', { title: '  ' }))).status, 400);
    assert.equal((await request('/todos', json('POST', null))).status, 400);
    assert.equal((await request('/todos', { method: 'POST', headers: { 'content-type': 'application/json' }, body: '{' })).status, 400);
    const created = await request('/todos', json('POST', { title: '  Ship release  ' }));
    assert.equal(created.status, 201);
    const todo = await created.json();
    assert.equal(todo.title, 'Ship release');
    const updated = await request(`/todos/${todo.id}`, json('PATCH', { title: 'Release shipped' }));
    assert.equal(updated.status, 200);
    assert.equal((await (await request(`/todos/${todo.id}`)).json()).title, 'Release shipped');
    assert.equal((await request('/todos/99999')).status, 404);
    assert.equal((await request('/todos/99999', json('PATCH', { title: 'Missing' }))).status, 404);
    assert.equal((await request('/todos/nope')).status, 400);
    const deletion = await request('/todos/1', { method: 'DELETE' });
    assert.equal(deletion.status, 200);
    assert.deepEqual(await deletion.json(), { deleted: true });
    assert.equal((await (await request('/todos')).json()).length, 2);
    assert.equal((await request('/todos/1')).status, 404);
    assert.equal((await request('/todos/1', { method: 'DELETE' })).status, 404);
    assert.equal((await request('/todos/nope', { method: 'DELETE' })).status, 400);
  } finally {
    const exited = once(child, 'exit');
    child.kill('SIGTERM');
    if (child.exitCode === null) await exited;
    rmSync(scratch, { recursive: true, force: true });
  }
});
