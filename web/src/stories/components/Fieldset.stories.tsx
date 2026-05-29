import type { Meta, StoryObj } from '@storybook/react';

const meta: Meta = {
  title: 'Components/Fieldset',
  parameters: { layout: 'centered' },
};

export default meta;

export const RoleSelector: StoryObj = {
  render: () => (
    <fieldset className="fieldset max-w-sm">
      <legend className="fieldset-legend">I am a…</legend>
      <div className="flex gap-2">
        <button type="button" className="btn btn-primary flex-1">
          Jobseeker
        </button>
        <button type="button" className="btn btn-outline flex-1">
          Company
        </button>
      </div>
    </fieldset>
  ),
};

export const AccountDetails: StoryObj = {
  render: () => (
    <fieldset className="fieldset max-w-sm">
      <legend className="fieldset-legend">Account Details</legend>
      <div className="space-y-3">
        <input className="input input-bordered w-full" placeholder="Full Name" />
        <input className="input input-bordered w-full" placeholder="Email" />
        <input className="input input-bordered w-full" type="password" placeholder="Password" />
      </div>
    </fieldset>
  ),
};
