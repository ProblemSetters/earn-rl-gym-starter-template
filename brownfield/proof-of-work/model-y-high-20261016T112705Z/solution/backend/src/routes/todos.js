const { Router } = require('express');
const { createTodoController } = require('../controllers/todoController');

function validateTitle(req, res, next) {
  const title = req.body?.title;
  if (typeof title !== 'string' || !title.trim() || title.trim().length > 200) {
    return res.status(400).json({ error: 'Title must contain between 1 and 200 characters' });
  }
  next();
}

function createTodoRouter(db) {
  const router = Router();
  const controller = createTodoController(db);
  router.param('id', (req, res, next, id) => {
    const number = Number(id);
    if (!/^[1-9]\d*$/.test(id) || !Number.isSafeInteger(number)) {
      return res.status(400).json({ error: 'Task id must be a positive integer' });
    }
    req.todoId = number;
    next();
  });
  router.get('/', controller.list);
  router.post('/', validateTitle, controller.create);
  router.get('/:id', controller.get);
  router.patch('/:id', validateTitle, controller.update);
  router.delete('/:id', controller.delete);
  return router;
}

module.exports = { createTodoRouter };
