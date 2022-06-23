import {
  FilterableTable,
  filterConfig,
  theme,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import { Link } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import CanaryStatus from '../../SharedComponent/CanaryStatus';
import { ReactComponent as CanaryIcon } from '../../../../assets/img/canary.svg';
import { ReactComponent as ABIcon } from '../../../../assets/img/ab.svg';
import { ReactComponent as BlueGreenIcon } from '../../../../assets/img/blue-green.svg';
import { ReactComponent as MirroringIcon } from '../../../../assets/img/mirroring.svg';
import { TableWrapper } from '../../CanaryStyles';
import {
  Canary,
  CanaryAnalysis,
  CanaryStatus as Status,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
interface Props {
  canaries: Canary[];
}
enum DeploymentStrategy {
  Canary = 'canary',
  AB = 'ab-testing',
  BlueGreen = 'blue-green',
  Mirroring = 'blue-green-mirror',
  NoAnalysis = 'no-analysis',
}

export const getDeploymentStrategyIcon = (strategy: string) => {
  switch (strategy.toLocaleLowerCase()) {
    case DeploymentStrategy.AB:
      return <ABIcon />;
    case DeploymentStrategy.BlueGreen:
      return <BlueGreenIcon />;
    case DeploymentStrategy.Mirroring:
      return <MirroringIcon />;
    case DeploymentStrategy.Canary:
      return <CanaryIcon />;
    default:
      return;
  }
};

export function getProgressValue(
  deploymentStrategy: string,
  status: Status | undefined,
  analysis: CanaryAnalysis | undefined,
): { current: number; total: number } {
  switch (deploymentStrategy) {
    case DeploymentStrategy.Canary:
      return {
        current: (status?.canaryWeight || 0) / (analysis?.stepWeight || 0),
        total: (analysis?.maxWeight || 0) / (analysis?.stepWeight || 0),
      };

    default:
      return {
        current: status?.iterations || 0,
        total: analysis?.iterations || 0,
      };
  }
}

export const CanaryTable: FC<Props> = ({ canaries }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfig(canaries, 'name'),
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
                    <Link
                      to={`/delivery?clusterName=${c.clusterName}&namespace=${c.namespace}&name=${c.name}`}
                      className={classes.link}
                    >
                      {c.name}
                      {'  '}
                      {getDeploymentStrategyIcon(c.deploymentStrategy || '')}
                    </Link>
                  ),
                },
                {
                  label: 'Status',
                  value: (c: Canary) => (
                    <div>
                      <CanaryStatus
                        status={c.status?.phase || ''}
                        value={getProgressValue(
                          c.deploymentStrategy || '',
                          c.status,
                          c.analysis,
                        )}
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