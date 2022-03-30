import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { CallReceived, CallMade, Remove } from '@material-ui/icons';

import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { Policy } from '../../../types/custom';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    
    icon: {
      color: weaveTheme.colors.neutral20,
    },
    iconTableCell: {
      width: 30,
    },
    noMaxWidth: {
      maxWidth: 'none',
    },
    normalRow: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
      padding:'16px'
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
      <TableRow data-cluster-name={name} className={`details ${classes.normalRow}`}>
        <TableCell>{name}</TableCell>
        <TableCell>{category}</TableCell>
        <TableCell>{SeverityComponent(severity || '')}</TableCell>
        <TableCell>{moment(createdAt).fromNow()}</TableCell>
      </TableRow>
    </>
  );
};

export default PolicyRow;
