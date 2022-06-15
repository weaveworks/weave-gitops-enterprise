import {
  FilterableTable,
  filterConfigForStatus,
  theme,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import { Link } from 'react-router-dom';
import styled, { ThemeProvider } from 'styled-components';
import { Canary } from '../../../../cluster-services/types.pb';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import CanaryStatus from '../../SharedComponent/CanaryStatus';
import { ReactComponent as CanaryIcon } from '../../../../assets/img/canary.svg';
import { ReactComponent as ABIcon } from '../../../../assets/img/ab.svg';
import { ReactComponent as BlueGreenIcon } from '../../../../assets/img/blue-green.svg';
import { ReactComponent as MirroringIcon } from '../../../../assets/img/mirroring.svg';
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

export const getDeploymentStrategyIcon = (strategy: string) => {
  switch (strategy.toLocaleLowerCase()) {
    case 'a/b':
      return <ABIcon />;
    case 'blue/green':
      return <BlueGreenIcon />;
    case 'blue/green mirroring':
      return <MirroringIcon />;
    default:
      return <CanaryIcon />;
  }
};

export const CanaryTable: FC<Props> = ({ canaries }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfigForStatus(canaries),
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
                        to={`/applications/delivery/${c.name}/${c.namespace}/${c.clusterName}`}
                        className={classes.link}
                      >
                        {c.name}{'  '}
                        {getDeploymentStrategyIcon(c.deploymentStrategy || '')}
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
