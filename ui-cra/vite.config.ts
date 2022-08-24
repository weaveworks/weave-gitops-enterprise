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
    // This needs dedup'ing when hot-reloading weave-gitops
    dedupe: ['@material-ui/styles'],
    alias: {
      // Allow (encourage) importing icons via full path, so that vite doesn't transform the
      // entire '@material-ui/icons' module during dev/build.
      // (10000 icons can take 30s when running `vite build`)
      '@material-ui/icons': '@material-ui/icons/esm',

      // In case an import or an import in a dependency (weave-gitops) imports something
      // "too deep" in mui and skips mui's CJS/ESM compat layers. e.g.
      // "import Button from @material-ui/core/Button/Button";
      // (vs "import Button from @material-ui/core/Button" which works okay)
      '@material-ui/core/': '@material-ui/core/esm/',

      ...localAlias,
    },
  },
  server: {
    proxy: {
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
