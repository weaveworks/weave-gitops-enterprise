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
import { useHistory } from "react-router-dom";

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
  const { notifications, clear } = useNotifications();
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
      <Dialog
        open
        maxWidth="sm"
        onClose={clear}
        BackdropProps={{ style: { backgroundColor: "transparent" } }}
        style={{ opacity: 0.9 }}
      >
        {/* We should have separate dialogs not separate content areas  */}
        {notifications?.map((notification: NotificationData, index: number) => (
          <DialogContent className={classes.content} key={index}>
            <div className={classes.mainWrapper}>
              {notification?.variant === "danger" ? (
                <ErrorIcon className={classes.icon} />
              ) : (
                <SuccessIcon className={classes.icon} />
              )}
              <div>
                <text style={{ color: getColor(notification?.variant) }}>
                  {notification?.variant === "danger" ? "Error" : "Success"}
                  :&nbsp;
                </text>
                {notification?.message}
              </div>
            </div>
            <div className={classes.closeIconWrapper}>
              <Close onClick={clear} />
            </div>
          </DialogContent>
        ))}
      </Dialog>
    </ThemeProvider>
  );
};
