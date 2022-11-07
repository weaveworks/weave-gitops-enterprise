import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import useNotifications, {
  NotificationData,
} from '../../contexts/Notifications';
import { Box, Collapse } from '@material-ui/core';

const { base } = theme.spacing;

const AlertWrapper = styled(Alert)`
  padding: ${base};
  margin: 0 ${base} ${base} ${base};
  border-radius: 10px;
  div[class*='MuiAlert-action'] {
    display: inline;
  }
`;

const Notifications: FC = () => {
  const { notifications, setNotifications } = useNotifications();

  const handleOpen = (n: NotificationData) => {
    const newNotif = notifications.filter(
      notif =>
        (n.message.text !== notif.message.text ||
          n.message.component !== notif.message.component) &&
        n.severity !== notif.severity,
    );
    setNotifications(newNotif);
  };

  const notificationAlert = (n: NotificationData, index: number) => {
    return (
      <Box key={index}>
        <Collapse in={true}>
          <AlertWrapper severity={n?.severity} onClose={() => handleOpen(n)}>
            {n?.message.text} {n?.message.component}
          </AlertWrapper>
        </Collapse>
      </Box>
    );
  };

  return (
    <>
      {notifications
        .filter(n => n.display !== 'bottom')
        .map((n, index) => {
          return (
            (n?.message.text || n?.message.component) &&
            notificationAlert(n, index)
          );
        })}
    </>
  );
};

export default Notifications;
