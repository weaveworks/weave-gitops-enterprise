import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Policy } from '../../../capi-server/capi_server.pb';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import Severity from '../Severity';
import moment from 'moment';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    normalRow: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    normalCell: {
      padding: theme.spacing(2),
    },
    link: {
      color: weaveTheme.colors.primary,
      fontWeight: 600,
    },
    severityName: {
      textTransform: 'capitalize',
    },
  }),
);

interface RowProps {
  policy: Policy;
}

const PolicyRow = ({ policy }: RowProps) => {
  const classes = useStyles();
  const { name, category, severity, createdAt, id } = policy;
  return (
    <>
      <TableRow data-cluster-name={name} className={classes.normalRow}>
        <TableCell className={classes.normalCell}>
          <Link to={`/policies/${id}`} className={classes.link}>
            {name}
          </Link>
        </TableCell>
        <TableCell className={classes.normalCell}>{category}</TableCell>
        <TableCell className={classes.normalCell}>
          <Severity severity={severity || ''} />
        </TableCell>
        <TableCell className={classes.normalCell}>
          {moment(createdAt).fromNow()}
        </TableCell>
      </TableRow>
    </>
  );
};

export default PolicyRow;
