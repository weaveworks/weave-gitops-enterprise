const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  app.use(
    '/gitops',
    createProxyMiddleware({
      target: 'http://localhost:8090',
      changeOrigin: true,
    }),
  );
  app.use(
    '/v1',
    createProxyMiddleware({
      target: 'http://localhost:8000',
      changeOrigin: true,
    }),
  );
};
