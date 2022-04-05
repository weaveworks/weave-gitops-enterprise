import React, { FC, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { Alert } from '../../types/kubernetes';
import { request } from '../../utils/request';
import { Alerts } from './index';
import useNotifications from './../Notifications';

const ALERTS_POLL_INTERVAL = 5000;

const AlertsProvider: FC = ({ children }) => {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const { notifications, setNotifications } = useNotifications();

  const alertsUrl = '/gitops/api/alerts';

  const fetchAlerts = () =>
    request('GET', alertsUrl, {
      cache: 'no-store',
    });

  const { error, data } = useQuery<{ data: any }, Error>(
    'alerts',
    () => fetchAlerts(),
    {
      refetchInterval: ALERTS_POLL_INTERVAL,
    },
  );

  useEffect(() => {
    if (data) {
      setAlerts(data.data.alerts);
    }

    if (
      error &&
      notifications?.some(
        notification => error.message === notification.message,
      ) === false
    ) {
      setNotifications([
        ...notifications,
        { message: error.message, variant: 'danger' },
      ]);
    }
  }, [data, error, notifications, setNotifications]);

  return <Alerts.Provider value={{ alerts }}>{children}</Alerts.Provider>;
};

export default AlertsProvider;
