import { HttpResponse, http } from 'msw';
import { setupServer } from 'msw/node';
import { afterAll, afterEach, beforeAll } from 'vitest';
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';

export const handlers = [
  http.get('/api/v1/auth/me', ({ request }) => {
    const authHeader = request.headers.get('Authorization');
    if (!authHeader) {
      return new HttpResponse(null, { status: 401 });
    }
    return HttpResponse.json({
      id: 'user-1',
      email: 'test@example.com',
      username: 'testuser',
      name: 'Test User',
      role: 'jobseeker',
      isVerified: true,
    });
  }),

  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({
      accessToken: 'test-access-token',
      refreshToken: 'test-refresh-token',
      user: {
        id: 'user-1',
        email: 'test@example.com',
        username: 'testuser',
        name: 'Test User',
        role: 'jobseeker',
        isVerified: true,
      },
    });
  }),

  http.post('/api/v1/auth/logout', () => {
    return HttpResponse.json({ message: 'logged out' });
  }),

  http.post('/api/v1/auth/refresh', () => {
    return HttpResponse.json({ accessToken: 'refreshed-token' });
  }),

  http.get('/api/v1/applications/me', () => {
    return HttpResponse.json([]);
  }),

  http.get('/api/v1/jobs', () => {
    return HttpResponse.json([]);
  }),

  http.get('/api/v1/industries', () => {
    return HttpResponse.json([{ id: 'ind-1', name: 'Technology', description: 'Tech industry' }]);
  }),

  http.get('/api/v1/notifications', () => {
    return HttpResponse.json({
      notifications: [],
      unreadCount: 0,
    });
  }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());

export { server };
