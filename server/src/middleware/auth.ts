import { Elysia } from 'elysia';

export const authMiddleware = new Elysia().guard({
  beforeHandle({ headers, error }) {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) {
      return error(401, 'Unauthorized');
    }
  },
});
