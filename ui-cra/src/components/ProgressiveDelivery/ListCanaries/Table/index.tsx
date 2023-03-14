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
  theme,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import moment from 'moment';
import React, { FC } from 'react';

import { Routes } from '../../../../utils/nav';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import { TableWrapper } from '../../../Shared';
import SVGIcon, { IconType } from '../../../WeGoSVGIcon';
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
      return <SVGIcon icon={IconType.AB} title="A/B Testing" />;
    case DeploymentStrategy.BlueGreen:
      return <SVGIcon icon={IconType.BlueGreen} title="Blue/Green" />;
    case DeploymentStrategy.Mirroring:
      return <SVGIcon icon={IconType.Mirroring} title="Blue/Green Mirroring" />;
    case DeploymentStrategy.Canary:
      return <SVGIcon icon={IconType.Canary} title="Canary Release" />;
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
  const classes = usePolicyStyle();

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
              <Link
                to={formatURL(Routes.CanaryDetails, {
                  clusterName: c.clusterName,
                  namespace: c.namespace,
                  name: c.name,
                })}
                className={classes.canaryLink}
              >
                {c.name}
                <span
                  style={{
                    marginLeft: theme.spacing.xs,
                  }}
                >
                  {getDeploymentStrategyIcon(c.deploymentStrategy || '')}
                </span>
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
              (c.status?.conditions && c.status?.conditions[0].message) || '--',
          },
          {
            label: 'Promoted',
            value: (c: Canary) => formatPromoted(c.targetDeployment),
          },
          {
            label: 'Last Updated',
            value: (c: Canary) =>
              (c.status?.conditions &&
                moment(c.status?.conditions[0].lastUpdateTime).fromNow()) ||
              '--',
            defaultSort: true,
            sortValue: (c: Canary) => {
              const t =
                c.status?.conditions &&
                new Date(
                  c.status?.conditions[0].lastUpdateTime || '',
                ).getTime();
              return Number(t) * -1;
            },
          },
        ]}
      />
    </TableWrapper>
  );
};
