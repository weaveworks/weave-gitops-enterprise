import {
  SourcesTable,
  useFeatureFlags,
  useListSources,
} from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import useNotifications from '../../contexts/Notifications';
import { formatError } from '../../utils/formatters';
import ScopedExploreUI from '../Explorer/ScopedExploreUI';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

const WGApplicationsSources: FC = () => {
  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );
  const {
    data: sources,
    isLoading,
    error,
  } = useListSources('', '', { enabled: !useQueryServiceBackend });
  const { setNotifications } = useNotifications();

  useEffect(() => {
    if (error) {
      setNotifications(formatError(error));
    }
  }, [error, setNotifications]);

  return (
    <PageTemplate
      documentTitle="Application Sources"
      path={[
        {
          label: 'Sources',
        },
      ]}
    >
      <ContentWrapper errors={sources?.errors} loading={isLoading}>
        {useQueryServiceBackend ? (
          <ScopedExploreUI
            enableBatchSync
            scopedKinds={[
              'GitRepository',
              'Bucket',
              'HelmChart',
              'HelmRepository',
              'OCIRepository',
            ]}
          />
        ) : (
          <SourcesTable sources={sources?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
