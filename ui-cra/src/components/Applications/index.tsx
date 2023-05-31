import {
  AutomationsTable,
  Button,
  Icon,
  IconType,
  Page,
  useFeatureFlags,
  useListAutomations,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Routes } from '../../utils/nav';
import OpenedPullRequest from '../Clusters/OpenedPullRequest';
import Explorer from '../Explorer/Explorer';
import { ActionsWrapper } from '../Clusters';
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
        <ActionsWrapper>
          <Button
            id="add-application"
            className="actionButton btn"
            startIcon={<Icon type={IconType.AddIcon} size="base" />}
            onClick={handleAddApplication}
          >
            ADD AN APPLICATION
          </Button>
          <OpenedPullRequest />
        </ActionsWrapper>
        <div className={className}>
          {useQueryServiceBackend ? (
            <Explorer category="automation" enableBatchSync />
          ) : (
            <AutomationsTable automations={automations?.result} />
          )}
        </div>
      </NotificationsWrapper>
    </Page>
  );
};

export default styled(WGApplicationsDashboard)`
  width: 100%;
  overflow: auto;
`;
