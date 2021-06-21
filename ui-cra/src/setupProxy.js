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
      target: 'http://35.232.103.122:30809/',
      changeOrigin: true,
    }),
  );
};
