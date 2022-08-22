import { FC } from 'react';
import { ListError } from '../../cluster-services/cluster_services.pb';
import { theme } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem, ListItemText } from '@material-ui/core';
import { uniqBy, sortBy } from 'lodash';

const { base } = theme.spacing;

const useStyles = makeStyles(() =>
  createStyles({
    alertWrapper: {
      padding: base,
      margin: `0 ${base} ${base} ${base}`,
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

function errorInfo(item: ListError): string {
  const msg = `Cluster: ${item.clusterName}`;
  if (!item.namespace) {
    return msg;
  }
  return `${msg}, Namespace: ${item.namespace}`;
}

export const AlertListErrors: FC<{ errors?: ListError[] }> = ({ errors }) => {
  const classes = useStyles();

  if (!errors || !errors.length) {
    return null;
  }

  // still not ideal
  const filteredErrors = sortBy(
    uniqBy(errors, error => [error.clusterName, error.message].join()),
    [v => v.clusterName, v => v.namespace, v => v.message],
  );

  return (
    <Alert className={classes.alertWrapper} severity="error">
      <AlertTitle>There were errors while listing some resources:</AlertTitle>
      {filteredErrors?.map((item: ListError, index: number) => (
        <ListItem key={index} dense={true}>
          <ListItemText
            className={classes.listItems}
            primary={item.message}
            secondary={errorInfo(item)}
          />
        </ListItem>
      ))}
    </Alert>
  );
};
