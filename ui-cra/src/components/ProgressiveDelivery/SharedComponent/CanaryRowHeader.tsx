import React from 'react';
import { useCanaryStyle } from '../CanaryStyles';

function CanaryRowHeader({
  rowkey,
  value,
}: {
  rowkey: string;
  value: string | undefined;
}) {
  const classes = useCanaryStyle();
  return (
    <div className={classes.rowHeaderWrapper}>
      <div className={classes.cardTitle}>{rowkey}:</div>
      <span className={classes.body1}>{value || '--'}</span>
    </div>
  );
}

export default CanaryRowHeader;
