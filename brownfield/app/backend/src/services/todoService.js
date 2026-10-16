function listTodos(db) {
  return db.prepare('SELECT id, title FROM todos ORDER BY id').all();
}

function getTodo(db, id) {
  return db.prepare('SELECT id, title FROM todos WHERE id = ?').get(id);
}

function createTodo(db, title) {
  const result = db.prepare('INSERT INTO todos (title) VALUES (?)').run(title);
  return getTodo(db, Number(result.lastInsertRowid));
}

function updateTodo(db, id, title) {
  const result = db.prepare('UPDATE todos SET title = ? WHERE id = ?').run(title, id);
  return result.changes > 0 ? getTodo(db, id) : undefined;
}

function deleteTodo(db, id) {
  // TODO: Delete the selected row and return whether a row was deleted.
  return null;
}

module.exports = { listTodos, getTodo, createTodo, updateTodo, deleteTodo };
