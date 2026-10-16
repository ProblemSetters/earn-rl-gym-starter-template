import { useState } from 'react';

export default function TaskItem({ todo, onRename, onDelete, disabled }) {
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(todo.title);
  async function submit(event) {
    event.preventDefault();
    if (title.trim() && await onRename(todo.id, title.trim())) setEditing(false);
  }
  return (
    <li className="task-item">
      <span className="task-marker" aria-hidden="true" />
      {editing ? (
        <form className="edit-form" onSubmit={submit}>
          <label className="sr-only" htmlFor={`task-${todo.id}`}>Task title</label>
          <input id={`task-${todo.id}`} value={title} onChange={(event) => setTitle(event.target.value)}
            maxLength={200} required autoFocus disabled={disabled} />
          <button type="submit" disabled={disabled || !title.trim()}>Save</button>
          <button type="button" disabled={disabled} onClick={() => setEditing(false)}>Cancel</button>
        </form>
      ) : (
        <>
          <div className="task-copy"><span>{todo.title}</span><small>TSK-{String(todo.id).padStart(3, '0')}</small></div>
          <div className="task-actions">
            <button type="button" disabled={disabled} onClick={() => { setTitle(todo.title); setEditing(true); }}>Edit</button>
            <button className="delete" type="button" disabled={disabled} onClick={() => onDelete(todo.id)}>Delete</button>
          </div>
        </>
      )}
    </li>
  );
}
