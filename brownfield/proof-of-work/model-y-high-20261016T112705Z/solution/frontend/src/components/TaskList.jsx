import TaskItem from './TaskItem.jsx';

export default function TaskList({ todos, loading, saving, onRename, onDelete }) {
  if (loading) return <p className="empty-state" role="status">Loading your tasks…</p>;
  if (!todos.length) return <div className="empty-state"><h3>A little room to focus</h3><p>Add your first task to get started.</p></div>;
  return <ul className="task-list" aria-label="Tasks">{todos.map((todo) => (
    <TaskItem key={todo.id} todo={todo} onRename={onRename} onDelete={onDelete} disabled={saving} />
  ))}</ul>;
}
