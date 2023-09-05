import React from 'react';
import ReactDOM from 'react-dom';
import reportWebVitals from './reportWebVitals';
import AppContainer from './App';

console.log('weave-gitops-enterprise ui:', process.env.REACT_APP_VERSION);

ReactDOM.render(
  <React.StrictMode>
    <AppContainer />
  </React.StrictMode>,
  document.getElementById('root'),
);

reportWebVitals();
