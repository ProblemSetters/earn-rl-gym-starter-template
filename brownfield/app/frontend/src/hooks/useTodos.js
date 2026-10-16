import { useEffect, useState } from 'react';
import { createTodo, listTodos, updateTodo } from '../api/todos.js';

export function useTodos() {
  const [todos, setTodos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    let active = true;
    listTodos()
      .then((items) => { if (active) setTodos(items); })
      .catch((error) => { if (active) setError(error.message); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  async function save(operation) {
    setError('');
    setSaving(true);
    try {
      await operation();
      return true;
    } catch (error) {
      setError(error.message);
      return false;
    } finally {
      setSaving(false);
    }
  }

  const addTodo = (title) => save(async () => {
    const todo = await createTodo(title);
    setTodos((current) => [...current, todo]);
  });
  const renameTodo = (id, title) => save(async () => {
    const todo = await updateTodo(id, title);
    setTodos((current) => current.map((item) => item.id === id ? todo : item));
  });

  async function deleteTodo(id) {
    // TODO: Call DELETE /api/todos/:id and update the list after success.
  }

  return { todos, loading, saving, error, addTodo, renameTodo, deleteTodo };
}
