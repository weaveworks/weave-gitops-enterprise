import {
  AutomationsTable,
  Button,
  Flex,
  Icon,
  IconType,
  useFeatureFlags,
  useListAutomations,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Routes } from '../../utils/nav';
import OpenedPullRequest from '../Clusters/OpenedPullRequest';
import ScopedExploreUI from '../Explorer/ScopedExploreUI';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

interface Size {
  size?: 'small';
}
const ActionsWrapper = styled(Flex)<Size>`
  & > .actionButton.btn {
    margin-right: ${({ theme }) => theme.spacing.small};
  }
`;

const WGApplicationsDashboard: FC = () => {
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
    <PageTemplate
      documentTitle="Applications"
      path={[
        {
          label: 'Applications',
        },
      ]}
    >
      <ContentWrapper loading={isLoading} errors={automations?.errors}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
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
        </div>

        {useQueryServiceBackend ? (
          <ScopedExploreUI category="automation" enableBatchSync />
        ) : (
          <AutomationsTable automations={automations?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
