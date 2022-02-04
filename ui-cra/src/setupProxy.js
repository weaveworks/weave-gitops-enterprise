const { createProxyMiddleware } = require('http-proxy-middleware');

const DEFAULT_PROXY_HOST = 'http://34.67.250.163:30080/';
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;
const capiServerHost = process.env.CAPI_SERVER_HOST || proxyHost;
const wegoServerHost = process.env.WEGO_SERVER_HOST || capiServerHost;

module.exports = function (app) {
  app.use(
    '/gitops',
    createProxyMiddleware({
      target: capiServerHost,
      changeOrigin: true,
    }),
  );
  app.use(
    '/v1/applications',
    createProxyMiddleware({
      target: wegoServerHost,
      changeOrigin: true,
    }),
  );
  app.use(
    '/v1',
    createProxyMiddleware({
      target: capiServerHost,
      changeOrigin: true,
    }),
  );
};
