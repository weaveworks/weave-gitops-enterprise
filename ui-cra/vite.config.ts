import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
// @ts-ignore
import svgrPlugin from 'vite-plugin-svgr';
import tsconfigPaths from 'vite-tsconfig-paths';

const DEFAULT_PROXY_HOST = 'https://demo-01.wge.dev.weave.works/';
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;
// Localhost is running tls by default now
const secure = process.env.PROXY_SECURE === 'true';

const proxyConfig = {
  target: proxyHost,
  changeOrigin: true,
  secure,
};

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    // This doesn't seem to be observed in package.json/resolutions so force it here for dev mode
    dedupe: ["@material-ui/styles"],
  },
  build: {
    outDir: 'build',
  },
  optimizeDeps: {
    esbuildOptions: {
      target: 'es2020',
    },
  },
  server: {
    proxy: {
      '/gitops': proxyConfig,
      '/v1': proxyConfig,
      '/oauth2': proxyConfig,
    },
  },
  plugins: [
    react(),
    tsconfigPaths(),
    svgrPlugin(),
  ],
});
