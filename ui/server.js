const fs = require('fs');
const express = require('express');
const setupProxy = require('../.proxyrc.js');

const app = express();

// static files
app.use(express.static('build'));

// serve index.html on react-router's browserHistory paths
// LIST OUT PATHS EXPLICITLY SO PROXY_HOST WILL STILL WORK.
//
app.use(
  [
    '/clusters',
    '/templates',
    '/applications',
    '/application_add',
    '/application_detail',
    '/application_remove',
    '/sign_in',
    '/oauth',
  ],
  (req, res, next) => {
    const writeIndexResponse = (err, result) => {
      if (err) {
        return next(err);
      }
      res.set('content-type', 'text/html');
      res.send(result);
      res.end();
    };
    fs.readFile('build/index.html', writeIndexResponse);
  },
);

// proxy
setupProxy(app);

const port = process.env.PORT || 5001;
const server = app.listen(port, () => {
  let { address } = server.address();
  if (address.indexOf(':') !== -1) {
    address = `[${address}]`;
  }
  console.log('weave-gitops listening at http://%s:%s', address, port);
});
