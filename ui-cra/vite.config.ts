/// <reference types="vitest" />

import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import { viteStaticCopy } from 'vite-plugin-static-copy';
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

const localAlias = process.env.PROXY_LOCAL
  ? { '@weaveworks/weave-gitops': `${process.env.PWD}/../../weave-gitops/ui` }
  : {};

//
// https://vitejs.dev/config/
//
export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/setupTests.ts'],
    deps: {
      // So node esm module resolution doesn't work on this yet..
      // So transform it into commonjs
      inline: ['@weaveworks/weave-gitops'],
    },
  },
  resolve: {
    // This doesn't seem to be observed in package.json/resolutions so force it here for dev mode
    dedupe: ['@material-ui/styles'],
    alias: {
      ...localAlias,
    },
  },
  build: {
    // Same as CRA for now.
    outDir: 'build',
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
