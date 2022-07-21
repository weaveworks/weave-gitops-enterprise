import { TableCell, TableRow } from '@material-ui/core';
import { useCanaryStyle } from '../CanaryStyles';

export function KeyValueRow({ entryObj }: { entryObj: Array<any> }): JSX.Element {
  const [key, val] = entryObj;
  return (
    <TableRow key={key} data-testid={key}>
      <TableCell
        style={{
          textTransform: 'capitalize',
          width: '30%',
        }}
      >
        {key.replace(/([a-z])([A-Z])/g, '$1 $2')}
      </TableCell>
      <TableCell
        style={{
          width: '70%',
        }}
      >
        {val}
      </TableCell>
    </TableRow>
  );
}

function CanaryRowHeader({
  children,
  rowkey,
  value,
}: {
  children?: any;
  rowkey: string;
  value: string | undefined;
}) {
  const classes = useCanaryStyle();
  return (
    <div className={classes.rowHeaderWrapper} data-testid={rowkey}>
      <div className={classes.cardTitle}>{rowkey}:</div>
      <span className={classes.body1}>
        {value || '--'} {children}
      </span>
    </div>
  );
}

export default CanaryRowHeader;
