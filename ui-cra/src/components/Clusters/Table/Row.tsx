import React from 'react';
import {
  Checkbox,
  Collapse,
  IconButton,
  TableCell,
  TableRow,
  Theme,
} from '@material-ui/core';
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp';
import { createStyles, makeStyles } from '@material-ui/styles';
import Octicon, { Icon, Tools } from '@primer/octicons-react';
import { Cluster } from '../../../types/kubernetes';
import { EKS, ExistingInfra, GKE, Kind } from '../../../utils/icons';
import { ClusterNameLink, Tooltip } from '../../Shared';
import { CAPIClusterStatus } from '../CAPIClusterStatus';
import {
  getClusterStatus,
  ReadyStatus,
  Status,
  statusSummary,
} from '../Status';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

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
  index: number;
  cluster: Cluster;
  onEdit: (cluster: Cluster) => void;
  selected: boolean;
  onCheckboxClick: (event: any, name: string) => void;
}

const getClusterTypeIcon = (clusterType?: string): Icon | null => {
  if (clusterType === 'aws') {
    return EKS;
  } else if (clusterType === 'existingInfra') {
    return ExistingInfra;
  } else if (clusterType === 'gke') {
    return GKE;
  } else if (clusterType === 'kind') {
    return Kind;
  }
  return null;
};

const ClusterRow = ({
  index,
  cluster,
  onEdit,
  selected,
  onCheckboxClick,
}: RowProps) => {
  const classes = useStyles();
  const { name, status: clusterStatus, type: clusterType, updatedAt } = cluster;
  const status = getClusterStatus(clusterStatus);
  const icon = getClusterTypeIcon(clusterType);
  const [open, setOpen] = React.useState<boolean>(false);
  const labelId = `enhanced-table-checkbox-${index}`;

  return (
    <>
      <TableRow
        className={`summary ${classes.normalRow}`}
        data-cluster-name={name}
        key={name}
      >
        <TableCell padding="checkbox">
          <Checkbox
            checked={selected}
            inputProps={{ 'aria-labelledby': labelId }}
            style={{
              color: weaveTheme.colors.primary,
            }}
            onClick={event => onCheckboxClick(event, name)}
          />
        </TableCell>
        <TableCell className={classes.nameCell} align="left">
          <IconButton
            aria-label="expand row"
            size="small"
            disabled={!cluster.capiCluster}
            onClick={() => setOpen(!open)}
          >
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
          <ClusterNameLink cluster={cluster} />
        </TableCell>
        <TableCell
          title={`Cluster type: ${clusterType ?? 'Unknown'}`}
          className={classes.iconTableCell}
          align="left"
        >
          {icon && (
            <Octicon
              className={classes.icon}
              icon={icon}
              size={24}
              verticalAlign="middle"
            />
          )}
        </TableCell>
        <TableCell align="left">
          {/* Using div instead of forwardRefs in ReadyStatus */}
          <Tooltip
            disabled={status !== Status.lastSeen}
            title={statusSummary(status, updatedAt)}
            classes={{ tooltip: classes.noMaxWidth }}
          >
            <div>
              <ReadyStatus
                updatedAt={updatedAt}
                status={status}
                pullRequest={cluster.pullRequest}
                onClick={
                  cluster.capiCluster && cluster.pullRequest?.type !== 'delete'
                    ? () => setOpen(!open)
                    : undefined
                }
              />
            </div>
          </Tooltip>
        </TableCell>
        <TableCell
          title="Edit cluster"
          className={classes.iconTableCell}
          align="left"
        >
          <IconButton onClick={() => onEdit(cluster)}>
            <Octicon
              className={classes.icon}
              icon={Tools}
              size={16}
              verticalAlign="middle"
            />
          </IconButton>
        </TableCell>
      </TableRow>
      <TableRow
        data-cluster-name={name}
        className={`details ${classes.collapsibleRow}`}
      >
        <TableCell colSpan={8}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <CAPIClusterStatus
              clusterName={cluster.name}
              status={cluster.capiCluster?.status}
            />
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
};

export default ClusterRow;
