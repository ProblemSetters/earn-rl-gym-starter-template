const fs = require('node:fs');
const path = require('node:path');
const { DatabaseSync } = require('node:sqlite');

function openDatabase(filename) {
  if (filename !== ':memory:') fs.mkdirSync(path.dirname(filename), { recursive: true });
  const db = new DatabaseSync(filename);
  db.exec(`CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL
  )`);
  if (db.prepare('SELECT COUNT(*) AS count FROM todos').get().count === 0) {
    const insert = db.prepare('INSERT INTO todos (title) VALUES (?)');
    insert.run('Plan sprint review');
    insert.run('Update onboarding guide');
  }
  return db;
}

module.exports = { openDatabase };
