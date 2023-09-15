import {
  Canary,
  CanaryAnalysis,
  CanaryStatus as Status,
  CanaryTargetDeployment,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import {
  DataTable,
  filterConfig,
  formatURL,
  Link,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import moment from 'moment';
import React, { FC } from 'react';
import styled from 'styled-components';
import { ReactComponent as ABIcon } from '../../../../assets/img/ab.svg';
import { ReactComponent as BlueGreenIcon } from '../../../../assets/img/blue-green.svg';
import { ReactComponent as CanaryIcon } from '../../../../assets/img/canary.svg';
import { ReactComponent as MirroringIcon } from '../../../../assets/img/mirroring.svg';
import { Routes } from '../../../../utils/nav';
import { TableWrapper } from '../../../Shared';
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
      return <ABIcon title="A/B Testing" />;
    case DeploymentStrategy.BlueGreen:
      return <BlueGreenIcon title="Blue/Green" />;
    case DeploymentStrategy.Mirroring:
      return <MirroringIcon title="Blue/Green Mirroring" />;
    case DeploymentStrategy.Canary:
      return <CanaryIcon title="Canary Release" />;
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
    <Link newTab href={`https://${img}`}>
      {img}
    </Link>
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
  const initialFilterState = {
    ...filterConfig(canaries, 'namespace'),
    ...filterConfig(canaries, 'clusterName'),
  };

  return (
    <TableWrapper id="canaries-list">
      <DataTable
        filters={initialFilterState}
        rows={canaries}
        fields={[
          {
            label: 'Name',
            value: (c: Canary) => (
              <CanaryLink
                to={formatURL(Routes.CanaryDetails, {
                  clusterName: c.clusterName,
                  namespace: c.namespace,
                  name: c.name,
                })}
              >
                {c.name}
                <span
                  style={{
                    marginLeft: 8,
                  }}
                >
                  {getDeploymentStrategyIcon(c.deploymentStrategy || '')}
                </span>
              </CanaryLink>
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
              (c.status?.conditions &&
                _.get(c, ['status', 'conditions', 0, 'message'])) ||
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
                  _.get(c, ['status', 'conditions', 0, 'lastUpdateTime']),
                ).fromNow()) ||
              '--',
            defaultSort: true,
            sortValue: (c: Canary) => {
              const t =
                c.status?.conditions &&
                new Date(
                  _.get(c, ['status', 'conditions', 0, 'lastUpdateTime']) || '',
                ).getTime();
              return Number(t) * -1;
            },
          },
        ]}
      />
    </TableWrapper>
  );
};

const CanaryLink = styled(Link)`
  color: #00b3ec;
  font-weight: 600;
  display: flex;
  justify-content: start;
  align-items: center;
`;
