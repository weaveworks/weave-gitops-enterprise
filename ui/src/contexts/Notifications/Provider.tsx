import { FC, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { Notification, NotificationData } from './index';

const NotificationsProvider: FC = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationData[] | []>(
    [],
  );
  const history = useHistory();

  useEffect(() => {
    return history.listen(() => setNotifications([]));
  }, [history]);

  return (
    <Notification.Provider value={{ notifications, setNotifications }}>
      {children}
    </Notification.Provider>
  );
};

export default NotificationsProvider;
