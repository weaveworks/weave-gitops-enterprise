import { FC } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem, ListItemText } from '@material-ui/core';
import { MultiRequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { uniqBy } from 'lodash';

const xs = theme.spacing.xs;
const base = theme.spacing.base;

const useStyles = makeStyles(() =>
  createStyles({
    alertWrapper: {
      padding: base,
      margin: `${xs} ${base}`,
      borderRadius: '10px',
    },
    warning: {
      backgroundColor: theme.colors.feedbackLight,
    },
    listItems: {
      display: 'list-item',
    },
  }),
);

export const AlertListErrors: FC<{ errors?: ListError[] }> = ({ errors }) => {
  const classes = useStyles();

  if (!errors || !errors.length) {
    return null;
  }
  const filteredErrors = uniqBy(errors, error => {
    [error.clusterName, error.message].join();
  }) as MultiRequestError[];

  return (
    <>
      <Alert className={classes.alertWrapper} severity="error">
        <AlertTitle>There were errors while listing some resources:</AlertTitle>
        {filteredErrors?.map((item: ListError, index: number) => (
          <ListItem key={index} dense={true}>
            <ListItemText
              className={classes.listItems}
              primary={item.message}
              secondary={`Cluster: ${item.clusterName} , Namespace: ${item.namespace}`}
            />
          </ListItem>
        ))}
      </Alert>
    </>
  );
};
