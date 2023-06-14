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

export default styled(DynamicTable)`
&.fadeIn {
  transform: scaleY(0);
  transformOrigin: top;
  display: block;
  max-height: 0,
  transition: transform 0.15s ease,
},
fadeOut: {
  transform: scaleY(1)
  transformOrigin: 'top',
  transition: 'transform 0.15s ease',
},
`;
