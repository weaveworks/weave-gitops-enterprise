import { TableCell, TableRow } from '@material-ui/core';
import { Policy } from '../../../cluster-services/cluster_services.pb';
import { Link } from 'react-router-dom';
import Severity from '../Severity';
import moment from 'moment';
import { usePolicyStyle } from '../PolicyStyles';


interface RowProps {
  policy: Policy;
}

const PolicyRow = ({ policy }: RowProps) => {
  const classes = usePolicyStyle();
  const { name, category, severity, createdAt, id } = policy;
  return (
    <>
      <TableRow data-cluster-name={name} className={classes.tableHead}>
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
