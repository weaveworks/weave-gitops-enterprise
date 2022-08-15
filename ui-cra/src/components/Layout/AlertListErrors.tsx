import { FC } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem } from '@material-ui/core';
import { MultiRequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import _ from 'lodash';

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
  const filteredErrors = _.uniqBy(errors, error => {
    [error.clusterName, error.message].join();
  }) as MultiRequestError[];

  return (
    <>
      <Alert className={classes.alertWrapper} severity="error">
        <AlertTitle>There were errors while listing some resources:</AlertTitle>
        {filteredErrors?.map((item: ListError) => (
          <ListItem key={item.clusterName}>
            • error='{item.message}' cluster='{item.clusterName}' namespace='
            {item.namespace}'
          </ListItem>
        ))}
      </Alert>
    </>
  );
};
