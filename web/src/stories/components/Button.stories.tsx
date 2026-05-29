import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Components/Button',
  parameters: { layout: 'centered' },
};

export default meta;

export const Primary: StoryObj = {
  render: () => (
    <button type="button" className="btn btn-primary">
      Primary
    </button>
  ),
};

export const Ghost: StoryObj = {
  render: () => (
    <button type="button" className="btn btn-ghost">
      Ghost
    </button>
  ),
};

export const Outline: StoryObj = {
  render: () => (
    <button type="button" className="btn btn-outline">
      Outline
    </button>
  ),
};

export const Success: StoryObj = {
  render: () => (
    <button type="button" className="btn btn-success">
      Success
    </button>
  ),
};

export const Danger: StoryObj = {
  render: () => (
    <button type="button" className="btn btn-error">
      Error
    </button>
  ),
};

export const Disabled: StoryObj = {
  render: () => (
    <button type="button" className="btn btn-primary" disabled>
      Disabled
    </button>
  ),
};

export const AllVariants: StoryObj = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <button type="button" className="btn btn-primary">
        Primary
      </button>
      <button type="button" className="btn btn-secondary">
        Secondary
      </button>
      <button type="button" className="btn btn-ghost">
        Ghost
      </button>
      <button type="button" className="btn btn-outline">
        Outline
      </button>
      <button type="button" className="btn btn-success">
        Success
      </button>
      <button type="button" className="btn btn-error">
        Error
      </button>
      <button type="button" className="btn btn-primary" disabled>
        Disabled
      </button>
    </div>
  ),
};
