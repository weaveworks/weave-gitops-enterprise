import { Box, Collapse } from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { FC } from 'react';
import styled from 'styled-components';
import { ReactComponent as ErrorIcon } from '../../assets/img/error.svg';
import { ReactComponent as SuccessIcon } from '../../assets/img/success.svg';
import { ReactComponent as WarningIcon } from '../../assets/img/warning.svg';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';

const BoxWrapper = styled(Box)<{ severity: string }>`
  div[class*='MuiAlert-root'] {
    width: auto;
    margin-bottom: ${props => props.theme.spacing.base};
    border-radius: ${props => props.theme.spacing.xs};
  }
  div[class*='MuiAlert-action'] {
    display: inline;
    color: ${props => {
      if (props.severity === 'error') return props.theme.colors.alertLight;
      else if (props.severity === 'warning')
        return props.theme.colors.feedbackLight;
      else if (props.severity === 'success')
        return props.theme.colors.successLight;
      else return 'transparent';
    }};
    svg {
      fill: ${props => {
        if (props.severity === 'error') return props.theme.colors.alertMedium;
        else if (props.severity === 'warning')
          return props.theme.colors.feedbackMedium;
        else if (props.severity === 'success')
          return props.theme.colors.successMedium;
        else return 'transparent';
      }};
    }
  }
  div[class*='MuiAlert-icon'] {
    svg[class*='MuiSvgIcon-root'] {
      display: none;
    }
  }
  div[class*='MuiAlert-message'] {
    display: flex;
    justify-content: center;
    align-items: center;
    svg {
      margin-right: ${props => props.theme.spacing.xs};
    }
  }
  div[class*='MuiAlert-standardError'] {
    background-color: ${props => props.theme.colors.alertLight};
  }
  div[class*='MuiAlert-standardSuccess'] {
    background-color: ${props => props.theme.colors.successLight};
  }
  div[class*='MuiAlert-standardWarning'] {
    background-color: ${props => props.theme.colors.alertLight};
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
