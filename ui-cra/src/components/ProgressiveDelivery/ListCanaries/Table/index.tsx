import {
  FilterableTable,
  filterConfigForString,
  theme,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import { Link } from 'react-router-dom';
import styled, { ThemeProvider } from 'styled-components';
import { Canary } from '../../../../cluster-services/types.pb';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import CanaryStatus from '../../SharedComponent/CanaryStatus';
interface Props {
  canaries: Canary[];
}

const TableWrapper = styled.div`
  margin-top: ${theme.spacing.medium};
  div[class*='FilterDialog__SlideContainer'],
  div[class*='SearchField'] {
    overflow: hidden;
  }
  div[class*='FilterDialog'] {
    .Mui-checked {
      color: ${theme.colors.primary};
    }
  }
  tr {
    vertical-align:'center')};
  }
  max-width: calc(100vw - 220px);
`;

export const CanaryTable: FC<Props> = ({ canaries }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfigForString(canaries, 'clusterName'),
    ...filterConfigForString(canaries, 'namespace'),
  };

  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        <TableWrapper id="canaries-list">
          <FilterableTable
            key={canaries?.length}
            filters={initialFilterState}
            rows={canaries}
            fields={[
              {
                label: 'Name',
                value: (c: Canary) => (
                  <>
                    {!!c.status?.canaryWeight ? (
                      <>{c.name} </>
                    ) : (
                      <Link
                        to={`/applications/delivery/${c.targetDeployment?.uid}`}
                        className={classes.link}
                      >
                        {c.name}
                      </Link>
                    )}
                  </>
                ),
              },
              {
                label: 'Status',
                value: (c: Canary) => (
                  <div>
                    <CanaryStatus
                      status={c.status?.phase || ''}
                      canaryWeight={c.status?.canaryWeight || 0}
                    />
                  </div>
                ),
              },
              {
                label: 'Cluster',
                value: 'clusterName',
                textSearchable: true,
              },
              {
                label: 'Namespace',
                value: 'namespace',
              },
              {
                label: 'Target',
                value: (c: Canary) => c.targetReference?.name || '',
              },
              {
                label: 'Message',
                value: (c: Canary) =>
                  (c.status?.conditions && c.status?.conditions[0].message) ||
                  '--',
              },
              {
                label: 'Last Updated',
                value: (c: Canary) =>
                  (c.status?.conditions &&
                    moment(c.status?.conditions[0].lastUpdateTime).fromNow()) ||
                  '--',
              },
            ]}
          />
        </TableWrapper>
      </ThemeProvider>
    </div>
  );
};
