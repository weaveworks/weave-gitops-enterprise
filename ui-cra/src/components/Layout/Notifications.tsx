import React, { FC, useEffect } from 'react';
import {
  makeStyles,
  createTheme,
  ThemeProvider,
} from '@material-ui/core/styles';
import { createStyles } from '@material-ui/styles';
import { muiTheme } from '../../muiTheme';
import theme from 'weaveworks-ui-components/lib/theme';
import { ReactComponent as ErrorIcon } from '../../assets/img/error-icon.svg';
import { ReactComponent as SuccessIcon } from '../../assets/img/success-icon.svg';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const Notifications: FC = () => {
  const { notifications, setNotifications, setShowNotifications } =
    useNotifications();

  // clean up notifications when the user navigates away
  // const cleanUp = () => {
  //   setShowNotifications(false);
  //   setNotifications([]);
  // };

  useEffect(() => {
    notifications.forEach(notification => toast(notification.message));
  }, [notifications]);

  return (
    <div>
      <ToastContainer
        position="bottom-center"
        autoClose={20000}
        hideProgressBar
        newestOnTop={true}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
      />
    </div>
  );
};

export default Notifications;
