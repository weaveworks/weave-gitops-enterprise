import { createContext, Dispatch, useContext } from 'react';

export interface NotificationData {
  message: string;
  variant: 'success' | 'danger';
}

type NotificationContext = {
  notification: NotificationData | null;
  setNotification: Dispatch<React.SetStateAction<NotificationData | null>>;
};

export const Notification = createContext<NotificationContext | null>(null);

export default () => useContext(Notification) as NotificationContext;
