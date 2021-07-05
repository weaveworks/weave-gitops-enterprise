import { fromPairs, sortBy } from 'lodash';
import {
  Box,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Theme,
  Typography,
} from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import React, { FC } from 'react';
import styled from 'styled-components';
import { CAPICluster } from '../../types/kubernetes';

// styles

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    conditionNameCell: {
      width: 100,
    },
  }),
);

const StatusValue = styled.div`
  white-space: pre;
`;

// types

type Condition = { [key: string]: string };
type StatusRenderer = (
  key: string,
  status: CAPICluster['status'],
) => React.ReactChild;

// sorting/display hints

const conditionKeys = [
  'type',
  'status',
  'severity',
  'reason',
  'message',
  'lastTransitionTime',
];

const statusKeySortHint: { [key: string]: number } = fromPairs(
  ['phase', 'controlPlaneInitialized', 'infrastructureReady'].map(
    (value, index) => [value, index],
  ),
);

// renderers

const defaultRenderer: StatusRenderer = (key, status) => {
  return JSON.stringify(status[key], null, 2);
};

const conditionsRenderer: StatusRenderer = (key, status) => {
  if (!status.conditions) {
    // Not sure how we get here but...
    return <i>No conditions present</i>;
  }
  return (
    <Table size="small">
      <TableHead style={{ backgroundColor: 'unset' }}>
        <TableRow>
          {conditionKeys.map(key => (
            <TableCell key={key}>{key}</TableCell>
          ))}
        </TableRow>
      </TableHead>
      <TableBody>
        {status.conditions.map((cond: Condition, index: number) => {
          return (
            <TableRow key={index}>
              {conditionKeys.map(key => (
                <TableCell key={key} style={{ borderBottom: 'unset' }}>
                  {cond[key]}
                </TableCell>
              ))}
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
};

const statusRenderers: { [key: string]: StatusRenderer } = {
  conditions: conditionsRenderer,
};

export const CAPIClusterStatus: FC<{ status?: CAPICluster['status'] }> = ({
  status,
}) => {
  const classes = useStyles();
  if (!status) {
    return null;
  }

  // Note: sortBy pushes 'undefined' to end of lists.
  const sortedKeys = sortBy(Object.keys(status), key => statusKeySortHint[key]);

  return (
    <Box margin={2}>
      <Typography variant="h6" gutterBottom component="div">
        CAPI Status
      </Typography>

      <TableContainer component={Paper}>
        <Table size="small">
          <TableBody>
            {sortedKeys.map(key => {
              const renderer = statusRenderers[key] || defaultRenderer;
              return (
                <TableRow hover key={key}>
                  <TableCell
                    className={classes.conditionNameCell}
                    component="th"
                    scope="row"
                    style={{ borderBottom: 'unset' }}
                  >
                    {key}
                  </TableCell>
                  <TableCell style={{ borderBottom: 'unset' }}>
                    <StatusValue>{renderer(key, status)}</StatusValue>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};
