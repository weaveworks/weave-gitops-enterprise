import { FC, useEffect, useState } from 'react';
// import { useNavigate } from 'react-router-dom';
import { Notification, NotificationData } from './index';
import { useLocation } from 'react-router-dom';

const NotificationsProvider: FC = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationData[] | []>(
    [],
  );
  const location = useLocation();

  // Clear notifications when navigating to a new page
  useEffect(() => {
    setNotifications([]);
  }, [setNotifications, location]);

  return (
    <Notification.Provider value={{ notifications, setNotifications }}>
      {children}
    </Notification.Provider>
  );
};

export default NotificationsProvider;
