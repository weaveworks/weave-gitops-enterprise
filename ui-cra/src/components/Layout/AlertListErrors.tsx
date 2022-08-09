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

  return (
    <>
      {!!(errors && errors.length) && (
        <Alert className={classes.alertWrapper} severity="error">
          <AlertTitle>
            There was a problem retrieving results from some clusters:
          </AlertTitle>
          {errors?.map((item: ListError) => (
            <ListItem key={item.clusterName}>
              - Cluster {item.clusterName} {item.message}
            </ListItem>
          ))}
        </Alert>
      )}
    </>
  );
};
