const express = require('express');
const path = require('node:path');
const { createTodoRouter } = require('./routes/todos');

function createApp(db) {
  const app = express();
  app.disable('x-powered-by');
  app.use(express.json({ limit: '16kb' }));
  app.get('/api/health', (req, res) => res.json({ status: 'ok' }));
  app.use('/api/todos', createTodoRouter(db));
  app.use('/api', (req, res) => res.status(404).json({ error: 'API route not found' }));

  const frontendDist = path.resolve(__dirname, '../../frontend/dist');
  app.use(express.static(frontendDist));
  app.get('/', (req, res) => res.sendFile(path.join(frontendDist, 'index.html')));
  app.use((req, res) => res.status(404).json({ error: 'Route not found' }));
  app.use((error, req, res, next) => {
    if (error.status >= 400 && error.status < 500) {
      return res.status(error.status).json({ error: 'Invalid request body' });
    }
    console.error(error);
    res.status(500).json({ error: 'An unexpected error occurred' });
  });
  return app;
}

module.exports = { createApp };
