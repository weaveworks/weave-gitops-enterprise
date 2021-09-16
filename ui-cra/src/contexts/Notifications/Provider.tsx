import React, { FC, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import Notifications from '../../components/Layout/Notifications';
import { Notification, NotificationData } from './index';

const NotificationProvider: FC = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationData[] | []>(
    [],
  );
  const history = useHistory();

  useEffect(() => {
    return history.listen(() => setNotifications([]));
  }, [history, notifications]);

  return (
    <Notification.Provider value={{ notifications, setNotifications }}>
      {notifications.length !== 0 && <Notifications />}
      {children}
    </Notification.Provider>
  );
};

export default NotificationProvider;
