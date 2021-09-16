import React, { FC, useCallback, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import Notifications from '../../components/Layout/Notifications';
import { Notification, NotificationData } from './index';

const NotificationProvider: FC = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationData[] | []>(
    [],
  );
  const history = useHistory();

  const clearNotifications = useCallback(() => setNotifications([]), []);

  useEffect(() => {
    return history.listen(clearNotifications);
  }, [history, notifications, clearNotifications]);

  return (
    <Notification.Provider value={{ notifications, setNotifications }}>
      {notifications.length !== 0 && <Notifications />}
      {children}
    </Notification.Provider>
  );
};

export default NotificationProvider;
