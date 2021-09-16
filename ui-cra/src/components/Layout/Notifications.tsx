import React, { FC, useEffect } from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { createStyles } from '@material-ui/styles';
import theme from 'weaveworks-ui-components/lib/theme';
import { ReactComponent as ErrorIcon } from '../../assets/img/error-icon.svg';
import { ReactComponent as SuccessIcon } from '../../assets/img/success-icon.svg';
import useNotifications from '../../contexts/Notifications';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import styled from 'styled-components';

const ToastContainerWrapper = styled.div`
  .Toastify__toast-container {
    width: auto;
  }
  .Toastify__toast {
  }
  .Toastify__toast-body {
  }
  .Toastify__progress-bar {
  }
`;

const useStyles = makeStyles(() =>
  createStyles({
    content: {
      display: 'flex',
      boxShadow: theme.boxShadow.light,
    },
    icon: {
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
  const { notifications } = useNotifications();
  const classes = useStyles();

  const getColor = (variant?: string) => {
    if (variant === 'danger') {
      return '#BC3B1D';
    } else {
      return '#27AE60';
    }
  };

  useEffect(() => {
    notifications.forEach(notification =>
      toast(
        <div className={classes.mainWrapper}>
          <div>
            {notification?.variant === 'danger' ? (
              <ErrorIcon className={classes.icon} />
            ) : (
              <SuccessIcon className={classes.icon} />
            )}
          </div>
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
        {
          toastId: notification?.message,
        },
      ),
    );
  }, [notifications, classes]);

  return (
    <ToastContainerWrapper>
      <ToastContainer
        position="bottom-center"
        autoClose={10000}
        hideProgressBar
        newestOnTop={true}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
      />
    </ToastContainerWrapper>
  );
};

export default Notifications;
