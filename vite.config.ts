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
    strictPort: true,
    host: '127.0.0.1',
    proxy: {
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: false },
      '/healthz': { target: 'http://127.0.0.1:8080', changeOrigin: false },
    },
  },
});
