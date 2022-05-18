import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';

console.log('weave-gitops-enterprise ui:', import.meta.env.VITE_APP_VERSION);

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById('root'),
);
