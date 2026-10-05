import { defineConfig } from 'vite';
import solid from '@solidjs/vite-plugin';
import stylex from '@stylexjs/unplugin';

export default defineConfig({
  plugins: [
    stylex.vite({
      devMode: 'full',
      useCSSLayers: true,
    }),
    solid(),
  ],
  server: {
    port: 3000,
  },
});
