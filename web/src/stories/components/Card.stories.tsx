import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Components/Card',
  parameters: { layout: 'centered' },
};

export default meta;

export const Default: StoryObj = {
  render: () => (
    <div className="card bg-base-200 p-4 w-80">
      <h3 className="font-semibold text-xl">Card Title</h3>
      <p className="text-sm opacity-70 mt-2">This is a standard card using bg-base-200 with p-4 padding.</p>
    </div>
  ),
};

export const Bordered: StoryObj = {
  render: () => (
    <div className="card bg-base-100 border border-base-300 p-4 w-80">
      <h3 className="font-semibold text-xl">Bordered Card</h3>
      <p className="text-sm opacity-70 mt-2">This card uses bg-base-100 with a border for separation.</p>
    </div>
  ),
};
