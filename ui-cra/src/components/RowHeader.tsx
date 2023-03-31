import { TableCell, TableRow } from '@material-ui/core';
import { Flex, theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const { medium } = theme.fontSizes;
const { xs } = theme.spacing;
const { neutral30, neutral40 } = theme.colors;

const RowTitle = styled.div`
  font-weight: 600;
  font-size: ${medium};
  color: ${neutral30};
`;

const RowBody = styled.div`
  font-weight: 400;
  font-size: ${medium};
  margin-left: ${xs};
  color: ${neutral40};
`;

export interface SectionRowHeader {
  children?: any;
  rowkey: string;
  value?: string | JSX.Element | undefined;
  hidden?: boolean;
}
export const generateRowHeaders = (rows: Array<SectionRowHeader>) => {
  return rows.map(r => {
    return r.hidden === true ? null : (
      <RowHeader rowkey={r.rowkey} value={r.value} key={r.rowkey}>
        {r.children}
      </RowHeader>
    );
  });
};

export function KeyValueRow({
  entryObj,
}: {
  entryObj: Array<any>;
}): JSX.Element {
  const [key, val] = entryObj;
  return (
    <TableRow
      key={key}
      data-testid={key}
      style={{
        height: '40px',
      }}
    >
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

function RowHeader({
  children,
  rowkey,
  value,
}: {
  children?: any;
  rowkey: string;
  value: string | JSX.Element | undefined;
}) {
  return (
    <Flex start style={{ margin: `${xs} 0` }} testId={rowkey}>
      <RowTitle>{rowkey} :</RowTitle>
      <RowBody>{children || value || '--'}</RowBody>
    </Flex>
  );
}

export default RowHeader;
