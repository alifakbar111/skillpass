import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { ChecklistCard, type ChecklistStep } from './ChecklistCard';

function renderCard(steps: ChecklistStep[]) {
  return render(
    <MemoryRouter>
      <ChecklistCard title="Get set up" steps={steps} />
    </MemoryRouter>,
  );
}

describe('ChecklistCard', () => {
  it('renders pending steps with progress', () => {
    renderCard([
      { id: 'a', label: 'Add a headline', done: true },
      { id: 'b', label: 'Add an experience', done: false, to: '/somewhere', actionLabel: 'Add now' },
    ]);

    expect(screen.getByText('Get set up')).toBeInTheDocument();
    expect(screen.getByText('1/2 done')).toBeInTheDocument();
    expect(screen.getByText('Add an experience')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Add now' })).toHaveAttribute('href', '/somewhere');
  });

  it('invokes inline actions', () => {
    const onAction = vi.fn();
    renderCard([
      { id: 'a', label: 'Step', done: false, onAction, actionLabel: 'Start' },
      { id: 'b', label: 'Other', done: false },
    ]);

    screen.getByRole('button', { name: 'Start' }).click();
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it('disappears once everything is done', () => {
    renderCard([
      { id: 'a', label: 'One', done: true },
      { id: 'b', label: 'Two', done: true },
    ]);
    expect(screen.queryByTestId('onboarding-checklist')).not.toBeInTheDocument();
  });
});
