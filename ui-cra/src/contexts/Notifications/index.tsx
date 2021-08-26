import { createContext, Dispatch, useContext } from "react";

export interface NotificationData {
  message: string;
  variant: "success" | "danger";
}

type NotificationContext = {
  notifications: NotificationData[] | [];
  setNotifications: Dispatch<React.SetStateAction<NotificationData[] | []>>;
  setShowNotifications: Dispatch<React.SetStateAction<boolean>>;
};

export const Notification = createContext<NotificationContext | null>(null);

export default () => useContext(Notification) as NotificationContext;
