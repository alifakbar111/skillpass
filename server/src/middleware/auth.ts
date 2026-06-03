import { Elysia } from 'elysia';

export const authMiddleware = new Elysia().guard({
  beforeHandle({ headers, set }) {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) {
      set.status = 401;
      return { error: 'Unauthorized' };
    }
  },
});
