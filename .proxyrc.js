const { createProxyMiddleware } = require('http-proxy-middleware');

// 9001 is the default port that tilt starts the application on
const DEFAULT_PROXY_HOST = 'http://localhost:8000/';
const proxyHost = process.env.WEGO_EE_PROXY_HOST || DEFAULT_PROXY_HOST;

// Localhost is running tls by default now
const secure = process.env.PROXY_SECURE === 'true';

module.exports = function (app) {
  const proxyMiddleWare = createProxyMiddleware({
    target: proxyHost,
    changeOrigin: true,
    secure,
  });
  app.use('/v1', proxyMiddleWare);
  app.use('/debug', proxyMiddleWare);
  app.use('/oauth2', proxyMiddleWare);
};
