import { defineConfig } from 'vite'
import reactRefresh from '@vitejs/plugin-react-refresh'
import svgrPlugin from 'vite-plugin-svgr'

const DEFAULT_PROXY_HOST = 'http://34.67.250.163:30080/';
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;
const gitopsHost = process.env.GITOPS_HOST || proxyHost;
const capiServerHost = process.env.CAPI_SERVER_HOST || proxyHost;
const wegoServerHost = process.env.WEGO_SERVER_HOST || proxyHost;

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [reactRefresh(), svgrPlugin()],
  server: {
    proxy: {
      '/gitops': {
        target: gitopsHost,
        changeOrigin: true,
      },
  '/v1/applications': {
      target: wegoServerHost,
      changeOrigin: true,
    },
    '/v1': {
      target: capiServerHost,
      changeOrigin: true,
    },
  }
}
})
