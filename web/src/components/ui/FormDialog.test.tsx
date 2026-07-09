import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { FormDialog } from '@/components/ui/FormDialog';

describe('FormDialog', () => {
  it('renders title and children when open', () => {
    render(
      <FormDialog open title="Test Dialog" onClose={vi.fn()}>
        <p>Dialog content</p>
      </FormDialog>,
    );
    expect(screen.getByText('Test Dialog')).toBeInTheDocument();
    expect(screen.getByText('Dialog content')).toBeInTheDocument();
  });

  it('renders cancel button', () => {
    render(
      <FormDialog open title="Test" onClose={vi.fn()}>
        <p>Content</p>
      </FormDialog>,
    );
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('does not have open attribute when closed', () => {
    const { container } = render(
      <FormDialog open={false} title="Hidden" onClose={vi.fn()}>
        <p>Should not be visible</p>
      </FormDialog>,
    );
    expect(container.querySelector('dialog')).not.toHaveAttribute('open');
  });
});
