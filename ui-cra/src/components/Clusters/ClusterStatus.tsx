import { fromPairs, sortBy } from 'lodash';
import {
  Box,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import React, { FC } from 'react';
import styled from 'styled-components';
import { Icon, IconType, theme as weaveTheme } from '@weaveworks/weave-gitops';
import { CAPICluster } from '../../types/custom';
import { GitopsCluster } from '../../cluster-services/cluster_services.pb';

// styles

const useStyles = makeStyles(() =>
  createStyles({
    conditionNameCell: {
      width: 100,
    },
    section: {
      marginTop: `${weaveTheme.spacing.medium}`,
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
  const keyStatus = JSON.stringify(status[key], null, 2);
  if (
    [
      'controlPlaneReady',
      'controlPlaneInitialized',
      'infrastructureReady',
    ].includes(key)
  ) {
    return keyStatus === 'true' ? (
      <Icon type={IconType.SuccessIcon} size="base" />
    ) : (
      <Icon type={IconType.FailedIcon} size="base" />
    );
  }
  return keyStatus;
};

const conditionsRenderer = (conditions: Condition[]) => (
  <Table size="small">
    <TableHead style={{ backgroundColor: 'unset' }}>
      <TableRow>
        {conditionKeys.map(key => (
          <TableCell key={key}>{key}</TableCell>
        ))}
      </TableRow>
    </TableHead>
    <TableBody>
      {conditions.map((cond: Condition, index: number) => (
        <TableRow key={index}>
          {conditionKeys.map(key => (
            <TableCell key={key}>{cond[key]}</TableCell>
          ))}
        </TableRow>
      ))}
    </TableBody>
  </Table>
);

const capiConditionsRenderer: StatusRenderer = (_, status) => {
  if (!status.conditions) {
    // Not sure how we get here but...
    return <i>No conditions present</i>;
  }
  return conditionsRenderer(status.conditions);
};

const statusRenderers: { [key: string]: StatusRenderer } = {
  conditions: capiConditionsRenderer,
};

export const ClusterStatus: FC<{
  clusterName: string;
  status?: CAPICluster['status'];
  conditions?: GitopsCluster['conditions'];
}> = ({ status, conditions }) => {
  const classes = useStyles();

  if (!status && !conditions) {
    return null;
  }

  // Note: sortBy pushes 'undefined' to end of lists.
  const sortedKeys = sortBy(
    status && Object.keys(status),
    key => statusKeySortHint[key],
  );

  return (
    <>
      {status && (
        <Box>
          <Typography variant="h6" gutterBottom component="div">
            CAPI Status
          </Typography>
          <Table size="small">
            <TableBody>
              {sortedKeys.map(key => {
                const renderer = statusRenderers[key] || defaultRenderer;
                return (
                  <TableRow key={key}>
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
        </Box>
      )}
      {conditions && (
        <Box>
          <Typography variant="h6" gutterBottom component="div">
            Status
          </Typography>
          <Table size="small">
            <TableBody>
              <TableRow key="conditions">
                <TableCell style={{ borderBottom: 'unset' }}>
                  <StatusValue>{conditionsRenderer(conditions)}</StatusValue>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Box>
      )}
    </>
  );
};
