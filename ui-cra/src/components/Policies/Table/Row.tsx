import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { CallReceived, CallMade, Remove } from '@material-ui/icons';
import { Policy } from '../../../capi-server/capi_server.pb';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

import moment from 'moment';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    normalRow: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    normalCell: {
      padding: theme.spacing(2),
    },
    severityIcon: {
      fontSize: '14px',
      marginRight: '4px',
    },
    severityLow: {
      color: '#DFD41B',
    },
    severityMedium: {
      color: '#FF7000',
    },
    severityHigh: {
      color: '#E2423B',
    },
  }),
);

interface RowProps {
  policy: Policy;
}

const SeverityComponent = (severity: string) => {
  const classes = useStyles();
  switch (severity.toLocaleLowerCase()) {
    case 'low':
      return (
        <div className="flex-start">
          <CallReceived
            className={`${classes.severityIcon} ${classes.severityLow}`}
          />
          <span>{severity}</span>
        </div>
      );
    case 'high':
      return (
        <div className="flex-start">
          <CallMade
            className={`${classes.severityIcon} ${classes.severityHigh}`}
          />
          <span>{severity}</span>
        </div>
      );
    case 'medium':
      return (
        <div className="flex-start">
          <Remove
            className={`${classes.severityIcon} ${classes.severityMedium}`}
          />
          <span>{severity}</span>
        </div>
      );
  }
};

const PolicyRow = ({ policy }: RowProps) => {
  const classes = useStyles();
  const { name, category, severity, createdAt } = policy;
  return (
    <>
      <TableRow data-cluster-name={name} className={` ${classes.normalRow}`}>
        <TableCell className={` ${classes.normalCell}`}>{name}</TableCell>
        <TableCell className={` ${classes.normalCell}`}>{category}</TableCell>
        <TableCell className={` ${classes.normalCell}`}>
          {SeverityComponent(severity || '')}
        </TableCell>
        <TableCell className={` ${classes.normalCell}`}>
          {moment(createdAt).fromNow()}
        </TableCell>
      </TableRow>
    </>
  );
};

export default PolicyRow;
