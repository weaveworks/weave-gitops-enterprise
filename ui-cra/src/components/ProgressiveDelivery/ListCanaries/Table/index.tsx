import {
  FilterableTable,
  filterConfigForStatus,
  theme,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import { Link } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import { Canary } from '../../../../cluster-services/types.pb';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import CanaryStatus from '../../SharedComponent/CanaryStatus';
import { ReactComponent as CanaryIcon } from '../../../../assets/img/canary.svg';
import { ReactComponent as ABIcon } from '../../../../assets/img/ab.svg';
import { ReactComponent as BlueGreenIcon } from '../../../../assets/img/blue-green.svg';
import { ReactComponent as MirroringIcon } from '../../../../assets/img/mirroring.svg';
import { TableWrapper } from '../../CanaryStyles';
interface Props {
  canaries: Canary[];
}

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
        {canaries.length > 0 ? (
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
                          to={`/applications/delivery/${c.clusterName}/${c.namespace}/${c.name}`}
                          className={classes.link}
                        >
                          {c.name}
                          {'  '}
                          {getDeploymentStrategyIcon(
                            c.deploymentStrategy || '',
                          )}
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
                      moment(
                        c.status?.conditions[0].lastUpdateTime,
                      ).fromNow()) ||
                    '--',
                },
              ]}
            />
          </TableWrapper>
        ) : (
          <p>No data to display</p>
        )}
      </ThemeProvider>
    </div>
  );
};
