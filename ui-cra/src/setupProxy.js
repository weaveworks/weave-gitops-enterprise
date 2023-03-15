const { createProxyMiddleware } = require('http-proxy-middleware');

const proxyHost = process.env.WEGO_EE_PROXY_HOST || 'http://localhost:8000/';
// Accept self-signed certificates etc (for local development).
// If you are using demo-01 etc you can set PROXY_SECURE=true and it should work.
const secure = process.env.PROXY_SECURE === 'true';

module.exports = function (app) {
  const proxyMiddleWare = createProxyMiddleware({
    target: proxyHost,
    changeOrigin: true,
    secure,
  });
  app.use('/v1', proxyMiddleWare);
  app.use('/oauth2', proxyMiddleWare);
};
