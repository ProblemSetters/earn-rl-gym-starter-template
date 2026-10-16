import { useState } from 'react';

export default function TaskForm({ onAdd, disabled }) {
  const [title, setTitle] = useState('');
  async function submit(event) {
    event.preventDefault();
    if (title.trim() && await onAdd(title.trim())) setTitle('');
  }
  return (
    <form className="task-form" onSubmit={submit}>
      <label htmlFor="new-task">What needs to get done?</label>
      <div className="input-row">
        <input id="new-task" value={title} onChange={(event) => setTitle(event.target.value)}
          placeholder="Write a task for your team" maxLength={200} disabled={disabled} required />
        <button className="primary" disabled={disabled || !title.trim()} type="submit">Add task</button>
      </div>
    </form>
  );
}
