import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { CallReceived, CallMade, Remove } from '@material-ui/icons';
import { Policy } from '../../../capi-server/capi_server.pb';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

import moment from 'moment';
import styled from 'styled-components';

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
      color: weaveTheme.colors.primary,
    },
    severityMedium: {
      color: '#8A460A',
    },
    severityHigh: {
      color: '#9F3119',
    },
    severityName: {
      textTransform: 'capitalize',
    },
  }),
);

const FlexStart = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
`;

interface RowProps {
  policy: Policy;
}

const SeverityComponent = (severity: string) => {
  const classes = useStyles();
  switch (severity.toLocaleLowerCase()) {
    case 'low':
      return (
        <FlexStart>
          {' '}
          <CallReceived
            className={`${classes.severityIcon} ${classes.severityLow}`}
          />
          <span className={classes.severityName}>{severity}</span>
        </FlexStart>
      );
    case 'high':
      return (
        <FlexStart>
          {' '}
          <CallMade
            className={`${classes.severityIcon} ${classes.severityHigh}`}
          />
          <span className={classes.severityName}>{severity}</span>
        </FlexStart>
      );
    case 'medium':
      return (
        <FlexStart>
          {' '}
          <Remove
            className={`${classes.severityIcon} ${classes.severityMedium}`}
          />
          <span className={classes.severityName}>{severity}</span>
        </FlexStart>
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
