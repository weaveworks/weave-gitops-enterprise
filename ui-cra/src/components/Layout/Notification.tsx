import React, { FC, useState } from "react";
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
import useNotifications from "../../contexts/Notifications";
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
        maxWidth="md"
        onClose={onClose}
        BackdropProps={{ style: { backgroundColor: "transparent" } }}
        style={{ opacity: 0.8 }}
      >
        <DialogContent className={classes.content}>
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
            }}
          >
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
          <div
            style={{
              paddingLeft: theme.spacing.small,
              color: theme.colors.gray600,
              cursor: "pointer",
            }}
          >
            <Close onClick={onClose} />
          </div>
        </DialogContent>
      </Dialog>
    </ThemeProvider>
  );
};
