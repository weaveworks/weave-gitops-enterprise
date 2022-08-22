/// <reference types="vitest/globals" />

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

//
// https://vitejs.dev/config/
//
export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/setupTests.ts'],
    deps: {
      inline: ['@weaveworks/weave-gitops'],
    },
  },
  // ssr: {
  //   noExternal: ['styled-components'],
  // },
  resolve: {
    // This doesn't seem to be observed in package.json/resolutions so force it here for dev mode
    dedupe: ['@material-ui/styles'],
    // alias: {
    //   '@weaveworks/weave-gitops': '/Users/simon/weave/weave-gitops/ui',
    // },
  },
  build: {
    outDir: 'build',
  },
  optimizeDeps: {
    // include: ['@weaveworks/weave-gitops'],
    // esbuildOptions: {
    //   target: 'es2020',
    // },
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
    svgrPlugin(),
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
