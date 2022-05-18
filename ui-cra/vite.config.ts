import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import { viteStaticCopy } from 'vite-plugin-static-copy';
// @ts-ignore
import svgrPlugin from 'vite-plugin-svgr';

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
  build: {
    outDir: 'build',
    commonjsOptions: {
      // Fix for some old libs like dagre
      // https://github.com/vitejs/vite/issues/5759#issuecomment-1034461225
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
    // TODO: open issue on vite an explore options
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
