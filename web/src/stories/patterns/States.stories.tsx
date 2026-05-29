import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Patterns/States',
  parameters: { layout: 'centered' },
};

export default meta;

export const LoadingState: StoryObj = {
  render: () => (
    <div className="flex items-center justify-center p-8 bg-base-200 rounded-lg w-80">
      <span className="loading loading-spinner loading-lg text-primary" />
    </div>
  ),
};

export const EmptyState: StoryObj = {
  render: () => (
    <div className="p-8 bg-base-200 rounded-lg text-center w-80">
      <p className="opacity-50">No items found</p>
    </div>
  ),
};

export const ErrorState: StoryObj = {
  render: () => (
    <div className="p-4 bg-base-200 rounded-lg w-80">
      <p className="text-error text-sm" role="alert">
        Something went wrong. Please try again.
      </p>
    </div>
  ),
};
