import React, { FC, useEffect } from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { createStyles } from '@material-ui/styles';
import theme from 'weaveworks-ui-components/lib/theme';
import { ReactComponent as ErrorIcon } from '../../assets/img/error-icon.svg';
import { ReactComponent as SuccessIcon } from '../../assets/img/success-icon.svg';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const useStyles = makeStyles(() =>
  createStyles({
    content: {
      display: 'flex',
      boxShadow: theme.boxShadow.light,
    },
    icon: {
      // minWidth: '50px',
      // minHeight: '50px',
      width: '75px',
      height: '75px',
      marginRight: theme.spacing.small,
    },
    closeIconWrapper: {
      paddingLeft: theme.spacing.small,
      color: '#C1C1C1',
      cursor: 'pointer',
    },
    mainWrapper: {
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
    },
  }),
);

const Notifications: FC = () => {
  const { notifications, setNotifications, setShowNotifications } =
    useNotifications();
  const classes = useStyles();

  const getColor = (variant?: string) => {
    if (variant === 'danger') {
      return '#BC3B1D';
    } else {
      return '#27AE60';
    }
  };

  // clean up notifications when the user navigates away
  // const cleanUp = () => {
  //   setShowNotifications(false);
  //   setNotifications([]);
  // };

  useEffect(() => {
    // remove notification if it's already in the array
    notifications.forEach(notification =>
      toast(
        <div className={classes.mainWrapper}>
          {notification?.variant === 'danger' ? (
            <ErrorIcon className={classes.icon} />
          ) : (
            <SuccessIcon className={classes.icon} />
          )}
          <div>
            <strong
              style={{
                color: getColor(notification?.variant),
                flexWrap: 'wrap',
              }}
            >
              {notification?.variant === 'danger' ? 'Error' : 'Success'}
              :&nbsp;
            </strong>
            {notification?.message}
          </div>
        </div>,
      ),
    );
    // return cleanup();
  }, [notifications, classes]);

  return (
    <div>
      <ToastContainer
        style={{ width: '700px' }}
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
