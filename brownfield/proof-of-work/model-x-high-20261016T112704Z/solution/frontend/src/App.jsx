import TaskForm from './components/TaskForm.jsx';
import TaskList from './components/TaskList.jsx';
import { useTodos } from './hooks/useTodos.js';

export default function App() {
  const { todos, loading, saving, error, addTodo, renameTodo, deleteTodo } = useTodos();
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <a className="brand" href="/"><span className="brand-icon" aria-hidden="true">t.</span>Taskboard</a>
        <p className="workspace-label">YOUR WORKSPACE</p>
        <nav aria-label="Workspace"><a className="nav-link" href="/" aria-current="page"><span aria-hidden="true">▤</span> All tasks</a></nav>
        <p className="sidebar-note">Clear the noise.<br />Make space for good work.</p>
      </aside>
      <main>
        <header className="topbar"><span>Workspace / <strong>All tasks</strong></span><span className="workspace-badge">Team workspace</span></header>
        <div className="content">
          <p className="eyebrow">A LITTLE CLARITY, EVERY DAY</p>
          <div className="page-heading"><div><h1>All tasks</h1><p>Keep the work moving, one task at a time.</p></div><span className="count-badge">{loading ? '…' : todos.length} tasks</span></div>
          <section className="task-panel" aria-label="Task management">
            <TaskForm onAdd={addTodo} disabled={loading || saving} />
            {error && <p className="error" role="alert">{error}</p>}
            <div className="list-heading"><h2>Your task list</h2><span>Everything in one place</span></div>
            <TaskList todos={todos} loading={loading} saving={saving} onRename={renameTodo} onDelete={deleteTodo} />
          </section>
          <p className="footer-note">Small steps. Steady progress.</p>
        </div>
      </main>
    </div>
  );
}
