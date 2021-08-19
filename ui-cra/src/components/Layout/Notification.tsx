import React, { FC, useState } from 'react';
import { Dialog, DialogContent, DialogTitle } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import { createStyles } from '@material-ui/styles';
import theme from 'weaveworks-ui-components/lib/theme';
import { CloseIconButton } from '../../assets/img/close-icon-button';
import useNotifications from '../../contexts/Notifications';

const useStyles = makeStyles(() =>
  createStyles({
    dialog: {
      backgroundColor: theme.colors.gray50,
      borderRadius: theme.spacing.xs,
      boxShadow: theme.boxShadow.light,
    },
    content: {
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      padding: theme.spacing.medium,
    },
  }),
);

export const NotificationDialog: FC = () => {
  const [open, setOpen] = useState<boolean>(true);
  const { notification, setNotification } = useNotifications();
  const classes = useStyles();
  const getColor = (variant?: string) => {
    if (variant === 'danger') {
      return '#BC3B1D';
    } else {
      return '#27AE60';
    }
  };
  const onClose = () => {
    setOpen(false);
    setNotification(null);
  };

  return (
    <Dialog open maxWidth="sm" fullWidth onClose={onClose}>
      <div id="notification-popup" className={classes.dialog}>
        <DialogTitle disableTypography>
          <CloseIconButton onClick={onClose} />
        </DialogTitle>
        <DialogContent className={classes.content}>
          <span style={{ color: getColor(notification?.variant) }}>
            {notification?.variant}:
          </span>
          <div style={{ padding: theme.spacing.small }}>
            {notification?.message}
          </div>
        </DialogContent>
      </div>
    </Dialog>
  );
};
