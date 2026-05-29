import '../src/styles/index.css';
import { withThemeByDataAttribute } from '@storybook/addon-themes';
import type { Preview } from '@storybook/react';

const preview: Preview = {
  parameters: {
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
  },
  decorators: [
    withThemeByDataAttribute({
      themes: { winter: 'winter', dark: 'dark' },
      defaultTheme: 'winter',
      attributeName: 'data-theme',
    }),
  ],
};

export default preview;
