/// <reference types="vitest/config" />
import { fileURLToPath } from 'node:url';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: Number(process.env.DEV_PORT) || 4200,
    proxy: {
      '/api': process.env.API_PROXY_TARGET || 'http://localhost:1234',
      // Locally stored uploads (avatars) are served by the API server.
      '/uploads': process.env.API_PROXY_TARGET || 'http://localhost:1234',
    },
  },
  test: {
    environment: 'happy-dom',
    // threads pool: the default forks pool breaks under Bun on Windows
    // ("File URL path must be an absolute path").
    pool: 'threads',
    // Absolute path: vite-node on Windows rejects relative file URLs.
    setupFiles: [fileURLToPath(new URL('./src/test/setup.ts', import.meta.url))],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
});
