import React, { FC, useCallback, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { NotificationDialog } from '../../components/Layout/Notification';
import { Notification, NotificationData } from './index';

const NotificationProvider: FC = ({ children }) => {
  const [notification, setNotification] =
    useState<NotificationData | null>(null);
  const history = useHistory();

  const clear = useCallback(() => {
    setNotification(null);
  }, []);

  useEffect(() => {
    return history.listen(clear);
  }, [clear, history]);

  return (
    <Notification.Provider value={{ notification, setNotification }}>
      {notification ? <NotificationDialog /> : null}
      {children}
    </Notification.Provider>
  );
};

export default NotificationProvider;
