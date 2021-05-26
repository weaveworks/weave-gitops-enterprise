import React, { FC, useState } from 'react';
import { Alert } from '../../types/kubernetes';
import { request } from '../../utils/request';
import { useInterval } from '../../utils/use-interval';
import { Alerts } from './index';

const ALERTS_POLL_INTERVAL = 5000;

const AlertsProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [error, setError] = React.useState<string | null>(null);

  const alertsUrl = '/gitops/api/alerts';

  const fetchAlerts = () => {
    // abort any inflight requests
    abortController?.abort();

    const newAbortController = new AbortController();
    setAbortController(newAbortController);
    setLoading(true);
    request('GET', alertsUrl, {
      cache: 'no-store',
      signal: newAbortController.signal,
    })
      .then(res => {
        setAlerts(res.alerts);
        setError(null);
      })
      .catch(err => {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      })
      .finally(() => {
        setLoading(false);
        setAbortController(null);
      });
  };

  useInterval(() => fetchAlerts(), ALERTS_POLL_INTERVAL, true, []);

  return (
    <Alerts.Provider
      value={{
        alerts,
        error,
      }}
    >
      {loading && !alerts ? 'loader' : children}
    </Alerts.Provider>
  );
};

export default AlertsProvider;
