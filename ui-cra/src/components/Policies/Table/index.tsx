import {
  Checkbox,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableRow,
  TableSortLabel,
} from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import React, { FC, useEffect } from 'react';
import { Cluster } from '../../../types/kubernetes';
import { Pagination } from '../../Pagination';
import { ColumnHeaderTooltip } from '../../Shared';
import ClusterRow from './Row';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import useClusters from '../../../contexts/Clusters';
import useNotifications from '../../../contexts/Notifications';
import { useHistory } from 'react-router-dom';
import { Loader } from '../../Loader';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { Policy } from '../../../types/custom';
import PolicyRow from './Row';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

const useStyles = makeStyles(() =>
  createStyles({
    nameHeaderCell: {
      paddingLeft: weaveTheme.spacing.medium,
    },
    paper: {
      marginBottom: 10,
      marginTop: 10,
      overflowX: 'auto',
      width: '100%',
    },
    root: {
      width: '100%',
    },
    disabled: {
      opacity: 0.5,
    },
    table: {
      whiteSpace: 'nowrap',
    },
    tableHead: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    noMaxWidth: {
      maxWidth: 'none',
    },
    tablePagination: {
      height: '80px',
    },
  }),
);

interface Props {
  policies: Policy[] | null;
  // count: number | null;
}

export const PolicyTable: FC<Props> = ({ policies }) => {
  const classes = useStyles();

  return (
    <div className={`${classes.root}`} id="clusters-list">
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          <Table className={classes.table} size="small">
            {policies?.length === 0 ? (
              <caption>No clusters configured</caption>
            ) : null}
            <TableHead className={classes.tableHead}>
              <TableRow>
                <TableCell className={classes.nameHeaderCell} align="left">
                  <ColumnHeaderTooltip title="Name configured in management UI">
                    <span>Name</span>
                  </ColumnHeaderTooltip>
                </TableCell>
                <TableCell className={classes.nameHeaderCell} align="left">
                  <span>Category</span>
                </TableCell>
                <TableCell className={classes.nameHeaderCell} align="left">
                  <span>Severity</span>
                </TableCell>
                <TableCell className={classes.nameHeaderCell} align="left">
                  <span>Age</span>
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {policies?.map((policy: Policy, index: number) => {
                return <PolicyRow policy={policy} key={policy.name} />;
              })}
            </TableBody>
          </Table>
        </Paper>
      </ThemeProvider>
    </div>
  );
};
