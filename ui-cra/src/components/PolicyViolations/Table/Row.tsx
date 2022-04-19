import { TableCell, TableRow } from '@material-ui/core';
import { PolicyValidation } from '../../../capi-server/capi_server.pb';
import { Link } from 'react-router-dom';
import Severity from '../../Policies/Severity';
import moment from 'moment';
import { PolicyStyles } from '../../Policies/PolicyStyles';

interface RowProps {
  violation: PolicyValidation;
}

const PolicyViolationRow = ({ violation }: RowProps) => {
  const classes = PolicyStyles.useStyles();
  const { severity, createdAt, id, entity, message, namespace } = violation;
  return (
    <>
      <TableRow className={classes.normalRow}>
        <TableCell className={classes.normalCell}>
          <Link to={`/policies/${id}`} className={classes.link}>
            {message}
          </Link>
        </TableCell>
        <TableCell className={classes.normalCell}>
          <Severity severity={severity || ''} />
        </TableCell>
        <TableCell className={classes.normalCell}>
          {moment(createdAt).fromNow()}
        </TableCell>
        <TableCell className={classes.normalCell}>
          {`${namespace}/${entity}`}
        </TableCell>
      </TableRow>
    </>
  );
};

export default PolicyViolationRow;
