# Taskboard

A React task workspace backed by an Express API and SQLite. Node.js 22.12+ is required.

## Get started

```sh
npm ci
npm run build
npm start
```

Open http://127.0.0.1:3000. The Node server serves the built frontend and API on the same port.

For development, run these commands in two terminals from this directory:

```sh
npm run dev:api
npm run dev
```

Vite serves the frontend at http://127.0.0.1:5173 and proxies `/api` to the backend.
Set `PORT` for the backend port and `TASKBOARD_DB_PATH` for the SQLite filename.
The default database is `backend/data/taskboard.sqlite`. An empty database is seeded with
**Plan sprint review** and **Update onboarding guide**.

## Structure

```text
frontend/
  index.html                 Vite HTML entry
  vite.config.js             React plugin and development API proxy
  src/
    main.jsx                 React entry
    App.jsx                  Workspace layout
    components/              Task form, list, and editable task rows
    hooks/useTodos.js        Task state and mutation handlers
    api/todos.js             HTTP client
    styles.css               Responsive workspace styles
backend/
  src/
    server.js                Process startup and shutdown
    app.js                   Express middleware and frontend serving
    routes/todos.js          Routes and request validation
    controllers/todoController.js
    services/todoService.js  SQLite queries
    db/database.js           Database initialization
  test/api.test.cjs          API integration checks
```

## API

| Method | Path | Behavior |
| --- | --- | --- |
| GET | `/api/health` | Health check |
| GET | `/api/todos` | List tasks |
| GET | `/api/todos/:id` | Fetch one task |
| POST | `/api/todos` | Create with `{ "title": "..." }` |
| PATCH | `/api/todos/:id` | Rename with `{ "title": "..." }` |
| DELETE | `/api/todos/:id` | Defined; returns 501 until implemented |

Titles must contain 1–200 characters after trimming. Invalid input returns 400; missing tasks
return 404. All API responses are JSON.

## Exercise

Deletion is intentionally unfinished in `backend/src/services/todoService.js` and
`frontend/src/hooks/useTodos.js`. The Delete buttons and controller wiring already exist.
Complete the service and connect the UI handler. Keep list, add, and edit working.

Run `npm test` for the API checks and `npm run build` for the frontend build.
The API check currently expects the baseline's 501 response; update that assertion when
implementing deletion. The gym's separate browser verifier scores the finished deletion flow.

The gym images include dependencies in `/opt/taskboard-deps/node_modules`. The task's setup
links these into the workspace and builds with Vite without network access.
