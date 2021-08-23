import React, { FC, useState } from "react";
import { Dialog, DialogContent, DialogTitle } from "@material-ui/core";
import {
  makeStyles,
  createMuiTheme,
  ThemeProvider,
} from "@material-ui/core/styles";
import { createStyles } from "@material-ui/styles";
import { muiTheme } from "../../muiTheme";
import theme from "weaveworks-ui-components/lib/theme";
import { CloseIconButton } from "../../assets/img/close-icon-button";
import { ReactComponent as ErrorIcon } from "../../assets/img/error-icon.svg";
import { ReactComponent as SuccessIcon } from "../../assets/img/success-icon.svg";
import useNotifications from "../../contexts/Notifications";

const localMuiTheme = createMuiTheme({
  ...muiTheme,
  overrides: {
    ...muiTheme.overrides,
    MuiPaper: {
      ...muiTheme.overrides?.MuiPaper,
      rounded: {
        ...muiTheme.overrides?.MuiPaper?.rounded,
        borderRadius: theme.spacing.small,
      },
    },
  },
});

const useStyles = makeStyles(() =>
  createStyles({
    title: {
      display: "flex",
      justifyContent: "right",
    },
    content: {
      display: "flex",
      backgroundColor: theme.colors.gray50,
      boxShadow: theme.boxShadow.light,
      justifyContent: "center",
      alignItems: "center",
    },
  })
);

export const NotificationDialog: FC = () => {
  const [open, setOpen] = useState<boolean>(true);
  const { notification, setNotification } = useNotifications();
  const classes = useStyles();
  const getColor = (variant?: string) => {
    if (variant === "danger") {
      return "#BC3B1D";
    } else {
      return "#27AE60";
    }
  };
  const onClose = () => {
    setOpen(false);
    setNotification(null);
  };

  return (
    <ThemeProvider theme={localMuiTheme}>
      <Dialog
        open={open}
        // maxWidth="sm"
        onClose={onClose}
        BackdropProps={{ style: { backgroundColor: "transparent" } }}
        style={{ opacity: 0.8 }}
      >
        {/* <div id="notification-popup" className={classes.dialog}> */}
        <DialogTitle className={classes.title}>
          <CloseIconButton onClick={onClose} />
        </DialogTitle>
        <DialogContent className={classes.content}>
          <div style={{ width: "50px", height: "50px" }}>
            {notification?.variant === "danger" ? (
              <ErrorIcon />
            ) : (
              <SuccessIcon />
            )}
          </div>
          <span style={{ color: getColor(notification?.variant) }}>
            {notification?.variant === "danger" ? "Error" : "Success"}
            :&nbsp;
          </span>
          {notification?.message}
        </DialogContent>
      </Dialog>
    </ThemeProvider>
  );
};
