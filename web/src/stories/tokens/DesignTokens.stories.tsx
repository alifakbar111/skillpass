import type { Meta } from '@storybook/react';

const meta: Meta = {
  title: 'Design System/Tokens',
  parameters: { layout: 'centered' },
};

export default meta;

const swatches = [
  'base-100',
  'base-200',
  'base-300',
  'neutral',
  'primary',
  'secondary',
  'accent',
  'info',
  'success',
  'warning',
  'error',
] as const;

const sizes = [
  { token: 'text-xs', size: '0.75rem', usage: 'Labels, badges, errors' },
  { token: 'text-sm', size: '0.875rem', usage: 'Body text, descriptions' },
  { token: 'text-base', size: '1rem', usage: 'Default paragraph' },
  { token: 'text-lg', size: '1.125rem', usage: 'Taglines' },
  { token: 'text-xl', size: '1.25rem', usage: 'Section subheadings' },
  { token: 'text-2xl', size: '1.5rem', usage: 'Page headings' },
  { token: 'text-3xl', size: '1.875rem', usage: 'Hero headings' },
  { token: 'text-4xl', size: '2.25rem', usage: 'Large hero' },
  { token: 'text-5xl', size: '3rem', usage: 'Landing hero' },
] as const;

export const Colors = () => (
  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
    {swatches.map((name) => (
      <div key={name} className={`p-4 rounded-lg bg-${name} text-${name}-content border`}>
        <span className="font-semibold">{name}</span>
      </div>
    ))}
  </div>
);

export const Typography = () => (
  <div>
    {sizes.map(({ token, size, usage }) => (
      <p key={token} className={`${token} mb-1`}>
        <span className="opacity-60">
          {token} ({size})
        </span>{' '}
        — {usage}
      </p>
    ))}
    <div className="mt-4">
      <p className="text-sm opacity-60 mb-1">Monospace (Fira Code):</p>
      <p className="font-mono">const hello = "world";</p>
    </div>
  </div>
);

export const Spacing = () => (
  <div className="space-y-4">
    {[1, 2, 3, 4, 6, 8].map((n) => (
      <div key={n} className="flex items-center gap-2">
        <span className="text-sm w-16">{n}</span>
        <div className={`w-${n} h-8 bg-primary rounded`} />
        <span className="text-xs opacity-60">{n * 0.25}rem</span>
      </div>
    ))}
  </div>
);
