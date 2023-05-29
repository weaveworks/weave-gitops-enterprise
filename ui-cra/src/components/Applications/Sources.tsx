import {
  Page,
  SourcesTable,
  useFeatureFlags,
  useListSources,
} from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import styled from 'styled-components';
import useNotifications from '../../contexts/Notifications';
import { formatError } from '../../utils/formatters';
import Explorer from '../Explorer/Explorer';
import { ContentWrapper } from '../Layout/ContentWrapper';

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

  useEffect(() => {
    if (error) {
      setNotifications(formatError(error));
    }
  }, [error, setNotifications]);

  return (
    <Page
      path={[
        {
          label: 'Sources',
        },
      ]}
    >
      <ContentWrapper errors={sources?.errors} loading={isLoading}>
        <div className={className}>
          {useQueryServiceBackend ? (
            <Explorer enableBatchSync category="source" />
          ) : (
            <SourcesTable sources={sources?.result} />
          )}
        </div>
      </ContentWrapper>
    </Page>
  );
};

export default styled(WGApplicationsSources)`
  width: 100%;
  overflow: auto;
`;
