import { FC } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem } from '@material-ui/core';

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

  return (
    <>
      <Alert className={classes.alertWrapper} severity="error">
        <AlertTitle>
          There were errors while listing some resources:
        </AlertTitle>
        {errors?.map((item: ListError) => (
          <ListItem key={item.clusterName}>
            • error='{item.message}' cluster='{item.clusterName}' namespace='{item.namespace}'
          </ListItem>
        ))}
      </Alert>
    </>
  );
};
