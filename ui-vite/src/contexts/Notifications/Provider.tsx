import React, { FC, useCallback, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { Notification, NotificationData } from './index';

const NotificationProvider: FC = ({ children }) => {
  const [notification, setNotification] =
    useState<NotificationData | null>(null);
  const history = useHistory();

  // clear notifications after a specific period of time if the history doesn't change?
  const clear = useCallback(() => {
    setNotification(null);
  }, []);

  useEffect(() => {
    return history.listen(clear);
  }, [clear, history]);

  return (
    <Notification.Provider value={{ notification, setNotification }}>
      {children}
    </Notification.Provider>
  );
};

export default NotificationProvider;
