import { FC } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem, ListItemText } from '@material-ui/core';

const useStyles = makeStyles(() =>
  createStyles({
    alertWrapper: {
      marginTop: theme.spacing.medium,
      marginRight: theme.spacing.small,
      marginBottom: 0,
      marginLeft: theme.spacing.small,
      paddingRight: theme.spacing.medium,
      paddingLeft: theme.spacing.medium,
      borderRadius: theme.spacing.xs,
    },
    warning: {
      backgroundColor: theme.colors.feedbackLight,
    },
  }),
);

export const AlertListErrors: FC<{ errors?: ListError[] }> = ({ errors }) => {
  const classes = useStyles();

  if (!errors || !errors.length) {
    return null;
  }
  const err = `Something went wrong, User "wego-admin"`;

  return (
    <>
      <Alert className={classes.alertWrapper} severity="error">
        <AlertTitle>There were errors while listing some resources:</AlertTitle>
        {[...errors,...errors]?.map((item: ListError, index: number) => (
          <ListItem key={index} dense={true}>
            <ListItemText
              style={{display:'list-item'}}
              primary={item.message || err}
              secondary={`${item.clusterName} / ${item.namespace}`}
            />
          </ListItem>
        ))}
      </Alert>
    </>
  );
};
