import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Components/Loading',
  parameters: { layout: 'centered' },
};

export default meta;

export const Spinner: StoryObj = {
  render: () => <span className="loading loading-spinner loading-lg text-primary" />,
};

export const Dots: StoryObj = {
  render: () => <span className="loading loading-dots loading-lg" />,
};

export const Ring: StoryObj = {
  render: () => <span className="loading loading-ring loading-lg" />,
};

export const Bars: StoryObj = {
  render: () => <span className="loading loading-bars loading-lg" />,
};

export const AllSizes: StoryObj = {
  render: () => (
    <div className="flex items-end gap-4">
      <div className="flex flex-col items-center gap-1">
        <span className="loading loading-spinner loading-xs" />
        <span className="text-xs opacity-60">xs</span>
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className="loading loading-spinner loading-sm" />
        <span className="text-xs opacity-60">sm</span>
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className="loading loading-spinner loading-md" />
        <span className="text-xs opacity-60">md</span>
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className="loading loading-spinner loading-lg" />
        <span className="text-xs opacity-60">lg</span>
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className="loading loading-spinner loading-xl" />
        <span className="text-xs opacity-60">xl</span>
      </div>
    </div>
  ),
};
