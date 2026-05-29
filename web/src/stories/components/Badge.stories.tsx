import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Components/Badge',
  parameters: { layout: 'centered' },
};

export default meta;

export const Default: StoryObj = {
  render: () => <span className="badge">Default</span>,
};

export const Primary: StoryObj = {
  render: () => <span className="badge badge-primary">Primary</span>,
};

export const Success: StoryObj = {
  render: () => <span className="badge badge-success">Success</span>,
};

export const Ghost: StoryObj = {
  render: () => <span className="badge badge-ghost">Ghost</span>,
};

export const Outline: StoryObj = {
  render: () => <span className="badge badge-outline">Outline</span>,
};

export const AllVariants: StoryObj = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <span className="badge">Default</span>
      <span className="badge badge-primary">Primary</span>
      <span className="badge badge-success">Success</span>
      <span className="badge badge-ghost">Ghost</span>
      <span className="badge badge-outline">Outline</span>
      <span className="badge badge-sm">Small</span>
    </div>
  ),
};
