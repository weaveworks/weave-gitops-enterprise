import { TableCell, TableRow } from '@material-ui/core';
import { Flex } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const RowTitle = styled.div`
  font-weight: 600;
  font-size: ${props => props.theme.fontSizes};
  color: ${props => props.theme.colors.neutral30};
`;

const RowBody = styled.div`
  font-weight: 400;
  font-size: ${props => props.theme.fontSizes};
  margin-left: ${props => props.theme.spacing.xs};
  color: ${props => props.theme.colors.neutral40};
`;

export interface SectionRowHeader {
  children?: any;
  rowkey: string;
  value?: string | JSX.Element | undefined;
  hidden?: boolean;
}
export const RowHeaders = ({ rows }: { rows: Array<SectionRowHeader> }) => {
  return (
    <>
      {rows.map(r => {
        return r.hidden === true ? null : (
          <RowHeader rowkey={r.rowkey} value={r.value} key={r.rowkey}>
            {r.children}
          </RowHeader>
        );
      })}
    </>
  );
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
    <Flex start gap="8" data-testid={rowkey}>
      <RowTitle>{rowkey} :</RowTitle>
      <RowBody>{children || value || '--'}</RowBody>
    </Flex>
  );
}

export default RowHeader;
