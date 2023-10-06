import {
  Button,
  Icon,
  IconType,
  SourcesTable,
  useFeatureFlags,
  useListSources,
} from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import useNotifications from '../../contexts/Notifications';
import { formatError } from '../../utils/formatters';
import { Routes } from '../../utils/nav';

import OpenedPullRequest from '../Clusters/OpenedPullRequest';
import Explorer from '../Explorer/Explorer';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const WGApplicationsSources: FC = ({ className }: any) => {
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );
  const {
    data: sources,
    isLoading,
    error,
  } = useListSources('', '', {
    enabled: !useQueryServiceBackend,
    retry: false,
    refetchInterval: 5000,
  });
  const { setNotifications } = useNotifications();
  const history = useHistory();

  useEffect(() => {
    if (error) {
      setNotifications(formatError(error));
    }
  }, [error, setNotifications]);

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Sources',
        },
      ]}
    >
      <NotificationsWrapper errors={sources?.errors}>
        <Button
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={() => history.push(Routes.AddSource)}
        >
          ADD A SOURCE
        </Button>
        <OpenedPullRequest />

        <div className={className}>
          {useQueryServiceBackend ? (
            <Explorer enableBatchSync category="source" />
          ) : (
            <SourcesTable sources={sources?.result} />
          )}
        </div>
      </NotificationsWrapper>
    </Page>
  );
};

export default styled(WGApplicationsSources)``;
