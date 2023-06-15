import { Size, Table, TableBody } from '@material-ui/core';
import styled from 'styled-components';
import { KeyValueRow } from '../../RowHeader';

const DynamicTable = ({
  obj,
  tableSize,
  classes,
}: {
  obj: Object;
  tableSize?: Size;
  classes?: string;
}) => {
  return (
    <Table size={tableSize || 'small'} className={classes}>
      <TableBody>
        {Object.entries(obj).map((entry, index) => (
          <KeyValueRow entryObj={entry} key={index}></KeyValueRow>
        ))}
      </TableBody>
    </Table>
  );
};

export default styled(DynamicTable)``;
