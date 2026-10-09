import { defineConfig } from '@vben/vite-config';
import { loadEnv } from 'vite';

export default defineConfig(async ({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            target: env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8081',
            ws: true,
          },
        },
      },
    },
  };
});
