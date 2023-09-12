import {
  AutomationsTable,
  Button,
  Flex,
  Icon,
  IconType,
  Link,
  V2Routes,
  formatURL,
  useFeatureFlags,
  useListAutomations,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import React, { FC } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { Routes } from '../../utils/nav';
import OpenedPullRequest from '../Clusters/OpenedPullRequest';
import Explorer from '../Explorer/Explorer';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const WGApplicationsDashboard: FC = ({ className }: any) => {
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );

  const { data: automations, isLoading } = useListAutomations('', {
    enabled: !useQueryServiceBackend,
    retry: false,
    refetchInterval: 5000,
  });

  const history = useHistory();

  const handleAddApplication = () => history.push(Routes.AddApplication);

  return (
    <Page
      loading={isLoading}
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
            {useQueryServiceBackend ? (
              <Explorer
                category="automation"
                enableBatchSync
                extraColumns={[
                  {
                    label: 'Source',
                    index: 4,
                    value: (o: Object & { parsed: any }) => {
                      const sourceAddr =
                        o.kind === 'HelmRelease'
                          ? 'spec.chart.spec.sourceRef.name'
                          : 'spec.sourceRef.name';

                      const url = formatURL(V2Routes.Sources, {
                        name: o.name,
                        namespace: o.namespace,
                        clusterName: o.cluster,
                      });

                      const sourceName = _.get(o.parsed, sourceAddr);

                      if (!sourceName) {
                        return '-';
                      }

                      return (
                        <Link to={url}>
                          {o.namespace}/{sourceName}
                        </Link>
                      );
                    },
                  },
                ]}
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

export default styled(WGApplicationsDashboard)``;
