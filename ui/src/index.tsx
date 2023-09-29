import React from 'react';
import ReactDOM from 'react-dom';
import AppContainer from './App';
import reportWebVitals from './reportWebVitals';

console.log('weave-gitops-enterprise ui:', process.env.REACT_APP_VERSION);

ReactDOM.render(
  <React.StrictMode>
    <AppContainer />
  </React.StrictMode>,
  document.getElementById('root'),
);

reportWebVitals();
