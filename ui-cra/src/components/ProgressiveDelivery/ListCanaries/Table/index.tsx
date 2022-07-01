import {
  Canary,
  CanaryAnalysis,
  CanaryStatus as Status,
  CanaryTargetDeployment
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { FilterableTable, filterConfig, theme } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import moment from 'moment';
import React, { FC } from 'react';
import { Link } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import { ReactComponent as ABIcon } from '../../../../assets/img/ab.svg';
import { ReactComponent as BlueGreenIcon } from '../../../../assets/img/blue-green.svg';
import { ReactComponent as CanaryIcon } from '../../../../assets/img/canary.svg';
import { ReactComponent as MirroringIcon } from '../../../../assets/img/mirroring.svg';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import { TableWrapper } from '../../CanaryStyles';
import CanaryStatus from '../../SharedComponent/CanaryStatus';
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

function toLink(img: string) {
  return (
    <a target="_blank" rel="noreferrer" href={`https://${img}`}>
      {img}
    </a>
  );
}

function formatPromoted(target?: CanaryTargetDeployment): any {
  if (!target) {
    return '';
  }

  // Most of the time we will only have one container and therefore one image.
  // If that is the case, we don't need to list container:image key value pairs.
  if (_.keys(target.promotedImageVersions).length === 1) {
    const img = _.first(_.values(target.promotedImageVersions)) || '';
    return toLink(img);
  }

  const out: any[] = [];
  _.each(target.promotedImageVersions, (img, container) => {
    out.push(
      <React.Fragment key={container}>
        <span>
          {container}: {toLink(img)}
        </span>{' '}
      </React.Fragment>,
    );
  });

  // Remove trainling comma
  return out;
}

export const CanaryTable: FC<Props> = ({ canaries }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfig(canaries, 'namespace'),
    ...filterConfig(canaries, 'clusterName'),
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
                      to={`/applications/delivery/${c.targetDeployment?.uid}?clusterName=${c.clusterName}&namespace=${c.namespace}&name=${c.name}`}
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
                  label: 'Promoted',
                  value: (c: Canary) => formatPromoted(c.targetDeployment),
                },
                {
                  label: 'Last Updated',
                  value: (c: Canary) =>
                    (c.status?.conditions &&
                      moment(
                        c.status?.conditions[0].lastUpdateTime,
                      ).fromNow()) ||
                    '--',
                  sortValue: (c: Canary) =>
                    c.status?.conditions![0].lastUpdateTime,
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
