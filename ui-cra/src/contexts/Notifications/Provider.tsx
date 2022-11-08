import React, { FC, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { useListVersion } from '../../hooks/versions';
import { Notification, NotificationData } from './index';

const NotificationProvider: FC = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationData[] | []>(
    [],
  );
  const history = useHistory();
  const { data, error } = useListVersion();
  const entitlement = data?.entitlement;

  useEffect(() => {
    return history.listen(() => setNotifications([]));
  }, [history, notifications, entitlement]);

  return (
    <Notification.Provider value={{ notifications, setNotifications }}>
      {children}
    </Notification.Provider>
  );
};

export default NotificationProvider;
