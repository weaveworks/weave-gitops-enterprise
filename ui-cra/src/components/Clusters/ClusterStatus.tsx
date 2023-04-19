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

const conditionsRenderer: StatusRenderer = (key, conditions) => {
  if (!conditions) {
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
};

const statusRenderers: { [key: string]: StatusRenderer } = {
  conditions: conditionsRenderer,
};

export const ClusterStatus: FC<{
  clusterName: string;
  conditions?: GitopsCluster['conditions'];
  status?: CAPICluster['status'];
}> = ({ status, conditions }) => {
  const classes = useStyles();

  if (!status && !conditions) {
    return null;
  }

  // Note: sortBy pushes 'undefined' to end of lists.
  const sortedKeys = (
    conditions?: GitopsCluster['conditions'],
    status?: CAPICluster['status'],
  ) => {
    if (status) {
      return sortBy(Object.keys(status), key => statusKeySortHint[key]);
    } else if (conditions) {
      return sortBy(Object.keys(conditions), key => statusKeySortHint[key]);
    }
    return [];
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom component="div">
        {conditions && 'Status'}
        {status && 'CAPI status'}
      </Typography>
      <Table size="small">
        <TableBody>
          {conditions &&
            sortedKeys(conditions).map(key => {
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
                    <StatusValue>
                      {conditions && renderer(key, conditions)}
                    </StatusValue>
                  </TableCell>
                </TableRow>
              );
            })}
          {status &&
            sortedKeys(status).map(key => {
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
                    <StatusValue>
                      {status && renderer(key, status.conditions)}
                    </StatusValue>
                  </TableCell>
                </TableRow>
              );
            })}
        </TableBody>
      </Table>
    </Box>
  );
};
