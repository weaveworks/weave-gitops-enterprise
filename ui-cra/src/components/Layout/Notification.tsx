import React, { FC } from "react";
import { Dialog, DialogContent } from "@material-ui/core";
import {
  makeStyles,
  createMuiTheme,
  ThemeProvider,
} from "@material-ui/core/styles";
import { createStyles } from "@material-ui/styles";
import { muiTheme } from "../../muiTheme";
import theme from "weaveworks-ui-components/lib/theme";
import { ReactComponent as ErrorIcon } from "../../assets/img/error-icon.svg";
import { ReactComponent as SuccessIcon } from "../../assets/img/success-icon.svg";
import useNotifications, {
  NotificationData,
} from "../../contexts/Notifications";
import { Close } from "@material-ui/icons";

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
    content: {
      display: "flex",
      boxShadow: theme.boxShadow.light,
    },
    icon: {
      minWidth: "50px",
      minHeight: "50px",
      width: "75px",
      height: "75px",
      marginRight: theme.spacing.small,
    },
    closeIconWrapper: {
      paddingLeft: theme.spacing.small,
      color: "#C1C1C1",
      cursor: "pointer",
    },
    mainWrapper: {
      display: "flex",
      justifyContent: "center",
      alignItems: "center",
    },
  })
);

export const NotificationDialog: FC = () => {
  const { notifications, setShowNotifications } = useNotifications();
  const classes = useStyles();
  const getColor = (variant?: string) => {
    if (variant === "danger") {
      return "#BC3B1D";
    } else {
      return "#27AE60";
    }
  };

  return (
    <ThemeProvider theme={localMuiTheme}>
      {notifications?.map((notification: NotificationData, index: number) => (
        <Dialog
          key={index}
          open
          maxWidth="sm"
          onClose={() => setShowNotifications(false)}
          BackdropProps={{ style: { backgroundColor: "transparent" } }}
          style={{ opacity: 0.9 }}
        >
          <DialogContent className={classes.content}>
            <div className={classes.mainWrapper}>
              {notification?.variant === "danger" ? (
                <ErrorIcon className={classes.icon} />
              ) : (
                <SuccessIcon className={classes.icon} />
              )}
              <div>
                <strong style={{ color: getColor(notification?.variant) }}>
                  {notification?.variant === "danger" ? "Error" : "Success"}
                  :&nbsp;
                </strong>
                {notification?.message}
              </div>
            </div>
            <div className={classes.closeIconWrapper}>
              <Close onClick={() => setShowNotifications(false)} />
            </div>
          </DialogContent>
        </Dialog>
      ))}
    </ThemeProvider>
  );
};
