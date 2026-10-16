const todoService = require('../services/todoService');

function createTodoController(db) {
  return {
    list(req, res) {
      res.json(todoService.listTodos(db));
    },
    get(req, res) {
      const todo = todoService.getTodo(db, req.todoId);
      if (!todo) return res.status(404).json({ error: 'Task not found' });
      res.json(todo);
    },
    create(req, res) {
      const todo = todoService.createTodo(db, req.body.title.trim());
      res.location(`/api/todos/${todo.id}`).status(201).json(todo);
    },
    update(req, res) {
      const todo = todoService.updateTodo(db, req.todoId, req.body.title.trim());
      if (!todo) return res.status(404).json({ error: 'Task not found' });
      res.json(todo);
    },
    delete(req, res) {
      const deleted = todoService.deleteTodo(db, req.todoId);
      if (deleted === null) return res.status(501).json({ error: 'Task deletion is not implemented' });
      if (!deleted) return res.json({ deleted: false });
      res.json({ deleted: false });
    },
  };
}

module.exports = { createTodoController };
