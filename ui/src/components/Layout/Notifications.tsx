import { Box, Collapse } from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { Button, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { ErrorIcon, SuccessIcon, WarningIcon } from '../RemoteSVGIcon';

interface Props {
  isClearable?: boolean;
  notifications: NotificationData[];
}

const BoxWrapper = styled(Box)<{ severity: string }>`
  div[class*='MuiAlert-root'] {
    width: auto;
    margin-bottom: ${props => props.theme.spacing.base};
    border-radius: ${props => props.theme.spacing.xs};
  }
  div[class*='MuiAlert-action'], div[class*='MuiAlert-message'] {
    display: inline;
    svg {
      fill: ${props => {
        if (props.severity === 'error') return props.theme.colors.alertDark;
        else if (props.severity === 'warning')
          return props.theme.colors.feedbackOriginal;
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
    background-color: ${props => props.theme.colors.feedbackLight};
  }
`;

const Notifications: FC<Props> = ({ notifications, isClearable = true }) => {
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
        return <Icon type={IconType.ErrorIcon} size="medium" />;
      case 'success':
        return <Icon type={IconType.SuccessIcon} size="medium" />;
      case 'warning':
        return isClearable ? (
          <Icon
            type={IconType.RemoveCircleIcon}
            size="medium"
            color="#ff9800"
          />
        ) : (
          <Icon type={IconType.WarningIcon} size="medium" color="#ff9800" />
        );
      default:
        return;
    }
  };

  const notificationAlert = (n: NotificationData, index: number) => {
    return (
      <BoxWrapper key={index} severity={n?.severity || ''}>
        <Collapse in={true}>
          <Alert
            severity={n?.severity}
            action={
              isClearable && (
                <span onClick={() => handleDelete(n)}>
                  <Icon type={IconType.ClearIcon} size="medium" />
                </span>
              )
            }
          >
            {getIcon(n?.severity)}
            <Text color="black">{n?.message.text}</Text> {n?.message.component}
          </Alert>
        </Collapse>
      </BoxWrapper>
    );
  };

  return (
    <Box style={{ width: '100%' }}>
      {notifications.map((n, index) => {
        return (
          (n?.message.text || n?.message.component) &&
          notificationAlert(n, index)
        );
      })}
    </Box>
  );
};

export default Notifications;
