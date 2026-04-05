import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

// Mirror the production Docker shape during development by serving the frontend
// from one origin and proxying /api requests to the Go backend.
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [react()],
    server: {
      proxy: {
        '/api': {
          target: env.VITE_DEV_API_PROXY_TARGET || 'http://127.0.0.1:8080',
          changeOrigin: true,
        },
      },
    },
  };
});
