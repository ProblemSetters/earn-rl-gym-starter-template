async function request(path = '', options = {}) {
  const response = await fetch(`/api/todos${path}`, options);
  const body = await response.json();
  if (!response.ok) throw new Error(body.error || 'The request failed. Please try again.');
  return body;
}

export const listTodos = () => request();
export const createTodo = (title) => request('', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ title }),
});
export const updateTodo = (id, title) => request(`/${id}`, {
  method: 'PATCH',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ title }),
});

export const removeTodo = (id) => request(`/${id}`, { method: 'DELETE' });
