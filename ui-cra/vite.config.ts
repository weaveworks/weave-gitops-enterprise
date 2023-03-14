import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

const proxyHost = process.env.PROXY_HOST || 'http://localhost:8000/';
const secure = process.env.PROXY_SECURE === 'true';

const proxyCfg = {
  target: proxyHost,
  changeOrigin: true,
  secure,
};
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    // https://vitejs.dev/config/server-options.html#server-proxy
    proxy: {
      '/v1': proxyCfg,
      '/oauth': proxyCfg,
    },
  },
});
