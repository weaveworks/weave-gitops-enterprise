import React from 'react';
import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Cluster } from '../../../types/kubernetes';

import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { Policy } from '../../../types/custom';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    actionButton: {
      fontSize: theme.typography.fontSize,
      margin: `${theme.spacing(0.5)}px ${theme.spacing(1)}px`,
    },
    icon: {
      color: weaveTheme.colors.neutral20,
    },
    nameHeaderCell: {
      paddingLeft: theme.spacing(4),
    },
    nameCell: {
      paddingLeft: theme.spacing(0.5),
      width: '300px',
    },
    iconTableCell: {
      width: 30,
    },
    noMaxWidth: {
      maxWidth: 'none',
    },
    normalRow: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    collapsibleRow: {
      '& > *': {
        paddingTop: 0,
        paddingBottom: 0,
      },
    },
  }),
);

interface RowProps {
  policy: Policy;
}

const PolicyRow = ({ policy }: RowProps) => {
  const classes = useStyles();
  const { name, category, severity, createdAt } = policy;
  return (
    <>
      <TableRow
        data-cluster-name={name}
        className={`details ${classes.collapsibleRow}`}
      >
        <TableCell>{name}</TableCell>
        <TableCell>{category}</TableCell>
        <TableCell>{severity}</TableCell>
        <TableCell>{createdAt}</TableCell>
      </TableRow>
    </>
  );
};

export default PolicyRow;
