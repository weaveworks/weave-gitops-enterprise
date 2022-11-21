import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { Box, Collapse } from '@material-ui/core';
import { ReactComponent as ErrorIcon } from '../../assets/img/error.svg';
import { ReactComponent as SuccessIcon } from '../../assets/img/success.svg';
import { ReactComponent as WarningIcon } from '../../assets/img/warning.svg';

const { xs, base } = theme.spacing;

const BoxWrapper = styled(Box)<{ severity: string }>`
  .MuiAlert-root {
    margin: 0 ${base} ${base} ${base};
    border-radius: ${xs};
  }
  .MuiAlert-action {
    display: inline;
    color: ${props => {
      if (props.severity === 'error') return '#d58572';
      else if (props.severity === 'warning') return theme.colors.feedbackLight;
      else if (props.severity === 'success') return theme.colors.success;
      else return 'transparent';
    }};
  }
  .MuiAlert-icon {
    .MuiSvgIcon-root {
      display: none;
    }
  }
  .MuiAlert-message {
    display: flex;
    justify-content: center;
    align-items: center;
    svg {
      margin-right: ${xs};
    }
  }
  .MuiAlert-standardError {
    background-color: #eecec7;
  }
  .MuiAlert-standardSuccess {
    background-color: #c9ebd7;
  }
  .MuiAlert-standardWarning {
    background-color: #fce6d2;
  }
`;

const Notifications: FC<{ notifications: NotificationData[] }> = ({
  notifications,
}) => {
  const { setNotifications } = useNotifications();

  const handleDelete = (n: NotificationData) =>
    setNotifications(
      notifications.filter(
        notif =>
          (n.message.text !== notif.message.text ||
            n.message.component !== notif.message.component) &&
          n.severity !== notif.severity,
      ),
    );

  const getIcon = (severity?: string) => {
    switch (severity) {
      case 'error':
        return <ErrorIcon />;
      case 'success':
        return <SuccessIcon />;
      case 'warning':
        return <WarningIcon />;
      default:
        return;
    }
  };

  const notificationAlert = (n: NotificationData, index: number) => {
    return (
      <BoxWrapper key={index} severity={n?.severity || ''}>
        <Collapse in={true}>
          <Alert severity={n?.severity} onClose={() => handleDelete(n)}>
            {getIcon(n?.severity)}
            {n?.message.text} {n?.message.component}
          </Alert>
        </Collapse>
      </BoxWrapper>
    );
  };

  return (
    <>
      {notifications.map((n, index) => {
        return (
          (n?.message.text || n?.message.component) &&
          notificationAlert(n, index)
        );
      })}
    </>
  );
};

export default Notifications;
