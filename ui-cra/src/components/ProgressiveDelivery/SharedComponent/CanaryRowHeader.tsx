import { TableCell, TableRow } from '@material-ui/core';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';

const { normal } = theme.fontSizes;
const { small, xs } = theme.spacing;
const { neutral30, neutral40 } = theme.colors;

const RowHeader = styled.div`
  margin: ${small} 0;
  display: flex;
  justify-content: start;
  align-items: center;
`;
const RowTitle = styled.div`
  font-weight: 600;
  font-size: ${normal};
  color: ${neutral30};
`;

const RowBody = styled.div`
  font-weight: 400;
  font-size: ${normal};
  margin-left: ${xs};
  color: ${neutral40};
`;

export const generateRowHeaders = (
  rows: Array<{
    children?: any;
    rowkey: string;
    value?: string | JSX.Element | undefined;
  }>,
) => {
  return rows.map(r => {
    return !!r.children ? (
      <CanaryRowHeader rowkey={r.rowkey} value={undefined} key={r.rowkey}>
        {r.children}
      </CanaryRowHeader>
    ) : (
      <CanaryRowHeader rowkey={r.rowkey} value={r.value} key={r.rowkey} />
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

function CanaryRowHeader({
  children,
  rowkey,
  value,
}: {
  children?: any;
  rowkey: string;
  value: string | JSX.Element | undefined;
}) {
  return (
    <RowHeader data-testid={rowkey}>
      <RowTitle>{rowkey} :</RowTitle>
      <RowBody>{children || value || '--'}</RowBody>
    </RowHeader>
  );
}

export default CanaryRowHeader;
