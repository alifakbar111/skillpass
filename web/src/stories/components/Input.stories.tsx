import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Components/Input',
  parameters: { layout: 'centered' },
};

export default meta;

export const Text: StoryObj = {
  render: () => (
    <label className="form-control w-full max-w-xs">
      <span className="label-text">Full Name</span>
      <input className="input input-bordered w-full" placeholder="John Doe" />
    </label>
  ),
};

export const Email: StoryObj = {
  render: () => (
    <label className="form-control w-full max-w-xs">
      <span className="label-text">Email</span>
      <input className="input input-bordered w-full" type="email" placeholder="john@example.com" />
    </label>
  ),
};

export const Password: StoryObj = {
  render: () => (
    <label className="form-control w-full max-w-xs">
      <span className="label-text">Password</span>
      <input className="input input-bordered w-full" type="password" />
    </label>
  ),
};

export const Textarea: StoryObj = {
  render: () => (
    <label className="form-control w-full max-w-xs">
      <span className="label-text">Description</span>
      <textarea className="textarea textarea-bordered w-full" rows={3} placeholder="Write something..." />
    </label>
  ),
};

export const WithError: StoryObj = {
  render: () => (
    <label className="form-control w-full max-w-xs">
      <span className="label-text">Email</span>
      <input className="input input-bordered w-full border-error" type="email" defaultValue="invalid" />
      <span className="text-error text-xs mt-1">Please enter a valid email</span>
    </label>
  ),
};
