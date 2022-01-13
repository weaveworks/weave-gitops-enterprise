import React, { useCallback } from 'react';
import { TableCell, TableRow } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Template } from '../../../types/custom';
import { useHistory } from 'react-router-dom';
import {
  theme as weaveTheme,
  Button,
  Icon,
  IconType,
} from '@weaveworks/weave-gitops';

const useStyles = makeStyles(() =>
  createStyles({
    icon: {
      color: '#ccc',
    },
    normalRow: {
      borderBottom: `1px solid #d8d8d8`,
    },
    error: {
      color: weaveTheme.colors.alert,
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
