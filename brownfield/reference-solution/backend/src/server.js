const path = require('node:path');
const { createApp } = require('./app');
const { openDatabase } = require('./db/database');

const port = Number(process.env.PORT || 3000);
const filename = process.env.TASKBOARD_DB_PATH || path.resolve(__dirname, '../data/taskboard.sqlite');
const db = openDatabase(filename);
const server = createApp(db).listen(port, '127.0.0.1', () => {
  console.log(`Taskboard listening on http://127.0.0.1:${port}`);
});

function close() {
  server.close(() => {
    db.close();
    process.exit(0);
  });
  server.closeIdleConnections();
}
process.on('SIGTERM', close);
process.on('SIGINT', close);
