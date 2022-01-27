import React, { FC, useCallback, useState } from 'react';
import { Alert } from '../../types/kubernetes';
import { request } from '../../utils/request';
import { useInterval } from '../../utils/use-interval';
import { Alerts } from './index';
import useNotifications from './../Notifications';

const ALERTS_POLL_INTERVAL = 5000;

const AlertsProvider: FC = ({ children }) => {
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const { notifications, setNotifications } = useNotifications();

  const alertsUrl = '/gitops/api/alerts';

  const fetchAlerts = useCallback(() => {
    abortController?.abort();

    const newAbortController = new AbortController();
    setAbortController(newAbortController);
    request('GET', alertsUrl, {
      cache: 'no-store',
      signal: newAbortController.signal,
    })
      .then(res => {
        setAlerts(res.alerts);
      })
      .catch(err => {
        if (
          err.name !== 'AbortError' &&
          notifications?.some(
            notification => err.message === notification.message,
          ) === false
        ) {
          setNotifications([
            ...notifications,
            { message: err.message, variant: 'danger' },
          ]);
        }
      })
      .finally(() => setAbortController(null));
  }, [abortController, notifications, setNotifications]);

  useInterval(() => fetchAlerts(), ALERTS_POLL_INTERVAL, true, []);

  return <Alerts.Provider value={{ alerts }}>{children}</Alerts.Provider>;
};

export default AlertsProvider;
