import { Box, Grid } from '@material-ui/core';
import _ from 'lodash';
import styled from 'styled-components';

export type KeyValuePairs = [string, any][];

type Props = {
  className?: string;
  pairs: KeyValuePairs;
};

const Key = styled.div`
  font-weight: bold;
`;

const Value = styled.div`
  text-overflow: ellipsis;
  overflow: hidden !important;
`;

const Item = styled(Box)``;

function KeyValueTable({ className, pairs }: Props) {
  return (
    <div role="list" className={className}>
      <Grid container spacing={2}>
        {_.map(pairs, (a, i) => {
          const [key, value] = a;

          const label = key;

          return (
            <Grid item key={i}>
              <Item p={1}>
                <Key aria-label={label}>{label}</Key>
                <Value>
                  {value || <span style={{ marginLeft: 2 }}>-</span>}
                </Value>
              </Item>
            </Grid>
          );
        })}
      </Grid>
    </div>
  );
}

export default styled(KeyValueTable).attrs({ className: 'KeyValueTable' })`
  width: 100%;

  tr {
    height: 72px;
    border-bottom: none;
  }
`;
