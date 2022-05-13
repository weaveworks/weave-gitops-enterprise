import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import svgrPlugin from 'vite-plugin-svgr';
import { viteStaticCopy } from 'vite-plugin-static-copy';

const DEFAULT_PROXY_HOST = 'http://34.67.250.163:30080/';
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;
const capiServerHost = process.env.CAPI_SERVER_HOST || proxyHost;
// Localhost is running tls by default now
const secure = process.env.PROXY_SECURE === 'true';

const proxyConfig = {
  target: capiServerHost,
  changeOrigin: true,
  secure,
};

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    outDir: 'build',
    commonjsOptions: {
      ignoreTryCatch: false,
    },
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
    svgrPlugin({
      svgrOptions: {
        // ...svgr options (https://react-svgr.com/docs/options/)
      },
    }),
    // Needed to see the svg's during dev
    viteStaticCopy({
      targets: [
        {
          src: './node_modules/@weaveworks/weave-gitops/*.svg',
          dest: 'node_modules/.vite/deps',
        },
      ],
    }),
  ],
});
