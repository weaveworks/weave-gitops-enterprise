import React, { useCallback } from 'react';
import { TableCell, TableRow } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { useHistory } from 'react-router-dom';
import { theme, Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { Template } from '../../../cluster-services/cluster_services.pb';

const useStyles = makeStyles(() =>
  createStyles({
    icon: {
      color: theme.colors.neutral20,
    },
    normalRow: {
      borderBottom: `1px solid ${theme.colors.neutral20}`,
    },
    error: {
      color: theme.colors.alert,
    },
  }),
);

interface RowProps {
  index: number;
  template: Template;
}

const TemplateRow = ({ index, template }: RowProps) => {
  const classes = useStyles();
  const { name, provider, description, error } = template;
  const history = useHistory();

  const handleAddCluster = useCallback(
    () => history.push(`/clusters/templates/${template.name}/create`),
    [history, template],
  );

  return (
    <TableRow
      className={`summary ${classes.normalRow}`}
      data-template-name={name}
      key={name}
    >
      <TableCell>{name}</TableCell>
      <TableCell>{provider}</TableCell>
      <TableCell>
        {description}
        <span className={classes.error}>{error}</span>
      </TableCell>
      <TableCell>
        <Button
          id="create-cluster"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={handleAddCluster}
          disabled={Boolean(error)}
        >
          CREATE CLUSTER WITH THIS TEMPLATE
        </Button>
      </TableCell>
    </TableRow>
  );
};

export default TemplateRow;
