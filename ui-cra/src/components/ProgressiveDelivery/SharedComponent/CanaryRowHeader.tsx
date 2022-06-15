import { useCanaryStyle } from '../CanaryStyles';

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
    <div className={classes.rowHeaderWrapper}>
      <div className={classes.cardTitle}>{rowkey}:</div>
      <span className={classes.body1}>
        {value || '--'} {children}
      </span>
    </div>
  );
}

export default CanaryRowHeader;
