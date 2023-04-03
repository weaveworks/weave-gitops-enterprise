import { SourcesTable, useListSources } from '@weaveworks/weave-gitops';
import { FC, useEffect } from 'react';
import useNotifications from '../../contexts/Notifications';
import { formatError } from '../../utils/formatters';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

const WGApplicationsSources: FC = () => {
  const { data: sources, isLoading, error } = useListSources();
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
        {sources && <SourcesTable sources={sources?.result} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
