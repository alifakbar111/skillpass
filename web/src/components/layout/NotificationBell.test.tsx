import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { NotificationBell } from '@/components/layout/NotificationBell';

vi.mock('@/lib/notifications', () => ({
  getNotifications: vi.fn().mockResolvedValue({ notifications: [], unreadCount: 0 }),
  markNotificationRead: vi.fn(),
  markAllNotificationsRead: vi.fn(),
  clearAllNotifications: vi.fn(),
  subscribeToNotifications: vi.fn().mockResolvedValue({
    close: vi.fn(),
    addEventListener: vi.fn(),
  }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe('NotificationBell', () => {
  it('renders bell icon button', async () => {
    render(<NotificationBell />, { wrapper });
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /notifications/i })).toBeInTheDocument();
    });
  });
});
