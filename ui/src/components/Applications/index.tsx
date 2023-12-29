import {
  AutomationsTable,
  Button,
  Flex,
  formatURL,
  Icon,
  IconType,
  Link,
  useListAutomations,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import React, { FC } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { EnabledComponent, Object } from '../../api/query/query.pb';
import { useIsEnabledForComponent } from '../../hooks/query';
import { getKindRoute, Routes } from '../../utils/nav';
import OpenedPullRequest from '../Clusters/OpenedPullRequest';
import Explorer from '../Explorer/Explorer';
import {
  addFieldsWithIndex,
  defaultExplorerFields,
} from '../Explorer/ExplorerTable';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const WGApplicationsDashboard: FC = ({ className }: any) => {
  const isExplorerEnabled = useIsEnabledForComponent(
    EnabledComponent.applications,
  );

  const { data: automations, isLoading } = useListAutomations('', {
    enabled: !isExplorerEnabled,
    retry: false,
    refetchInterval: 5000,
  });

  const history = useHistory();

  const handleAddApplication = () => history.push(Routes.AddApplication);

  return (
    <Page
      loading={!isExplorerEnabled && isLoading}
      path={[
        {
          label: 'Applications',
        },
      ]}
    >
      <NotificationsWrapper errors={automations?.errors}>
        <Flex column alignItems="stretch" gap="24">
          <Flex gap="12">
            <Button
              id="add-application"
              className="actionButton btn"
              startIcon={<Icon type={IconType.AddIcon} size="base" />}
              onClick={handleAddApplication}
            >
              ADD AN APPLICATION
            </Button>
            <OpenedPullRequest />
          </Flex>
          <div className={className}>
            {isExplorerEnabled ? (
              <Explorer
                category="automation"
                enableBatchSync
                fields={addFieldsWithIndex(defaultExplorerFields, [
                  {
                    id: 'source',
                    label: 'Source',
                    index: 4,
                    value: (o: Object & { parsed: any }) => {
                      const sourceAddr =
                        o.kind === 'HelmRelease'
                          ? 'spec.chart.spec.sourceRef'
                          : 'spec.sourceRef';

                      const sourceName = _.get(o.parsed, `${sourceAddr}.name`);
                      const sourceKind = _.get(o.parsed, `${sourceAddr}.kind`);

                      if (!sourceName || !sourceKind) {
                        return '-';
                      }

                      const kind = getKindRoute(sourceKind || '');

                      if (!kind) {
                        return sourceName;
                      }

                      const url = formatURL(kind, {
                        name: sourceName,
                        namespace: o.namespace,
                        clusterName: o.cluster,
                      });

                      return <Link to={url}>{sourceName}</Link>;
                    },
                  },
                ])}
              />
            ) : (
              <AutomationsTable automations={automations?.result} />
            )}
          </div>
        </Flex>
      </NotificationsWrapper>
    </Page>
  );
};

export default styled(WGApplicationsDashboard)`
  tbody tr td:nth-child(6) {
    white-space: nowrap;
  }
`;
