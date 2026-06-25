import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { AvatarUploader } from '@/components/jobseeker/AvatarUploader';

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe('AvatarUploader', () => {
  it('renders upload button', () => {
    render(<AvatarUploader name="John" onUploaded={vi.fn()} />, { wrapper });
    expect(screen.getByRole('button', { name: /change profile photo/i })).toBeInTheDocument();
  });

  it('shows initial when no avatar', () => {
    render(<AvatarUploader name="John" onUploaded={vi.fn()} />, { wrapper });
    expect(screen.getByText('J')).toBeInTheDocument();
  });

  it('shows avatar image when provided', () => {
    render(<AvatarUploader name="John" avatarUrl="/avatars/me.jpg" onUploaded={vi.fn()} />, { wrapper });
    const img = screen.getByRole('img', { name: /john avatar/i });
    expect(img).toHaveAttribute('src', '/avatars/me.jpg');
  });
});
