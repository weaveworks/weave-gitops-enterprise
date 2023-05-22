import { FC, useEffect, useState } from 'react';
// import { useNavigate } from 'react-router-dom';
import { Notification, NotificationData } from './index';
import { useOnLocationChange } from '../../utils/nav';

const NotificationsProvider: FC = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationData[] | []>(
    [],
  );
  // const navigate = useNavigate();
  const locationChange = useOnLocationChange(setNotifications);

  useEffect(() => {
    setNotifications([]);
    return locationChange;
  }, [setNotifications, locationChange]);

  // useEffect(() => {
  //   return history.listen(() => setNotifications([]));
  // }, [history]);

  return (
    <Notification.Provider value={{ notifications, setNotifications }}>
      {children}
    </Notification.Provider>
  );
};

export default NotificationsProvider;
