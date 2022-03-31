import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { CallReceived, CallMade, Remove } from '@material-ui/icons';

import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { Policy } from '../../../types/custom';
import moment from 'moment';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    normalRow: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    normalCell: {
      padding: theme.spacing(2),
    },
  }),
);

interface RowProps {
  policy: Policy;
}

const SeverityComponent = (severity: string) => {
  switch (severity.toLocaleLowerCase()) {
    case 'low':
      return (
        <div className="flex-center">
          <CallReceived className="severity-icon severity-low" />
          <span>{severity}</span>
        </div>
      );
    case 'high':
      return (
        <div className="flex-center">
          <CallMade className="severity-icon severity-high" />
          <span>{severity}</span>
        </div>
      );
    case 'medium':
      return (
        <div className="flex-center">
          <Remove className="severity-icon severity-medium" />
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
          {moment(Date.parse(createdAt.replace('EET', ''))).fromNow()}
        </TableCell>
      </TableRow>
    </>
  );
};

export default PolicyRow;
