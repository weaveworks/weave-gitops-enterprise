const { createProxyMiddleware } = require("http-proxy-middleware");

const DEFAULT_PROXY_HOST = "http://34.67.250.163:30080/";
const proxyHost = process.env.PROXY_HOST || DEFAULT_PROXY_HOST;
const gitopsHost = process.env.GITOPS_HOST || proxyHost;
const capiServerHost = process.env.CAPI_SERVER_HOST || proxyHost;

module.exports = function (app) {
  app.use(
    "/gitops",
    createProxyMiddleware({
      target: gitopsHost,
      changeOrigin: true,
    })
  );
  app.use(
    "/v1",
    createProxyMiddleware({
      target: capiServerHost,
      changeOrigin: true,
    })
  );
};
