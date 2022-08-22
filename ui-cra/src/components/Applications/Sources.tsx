import { Alert } from '@material-ui/lab';
import { SourcesTable, useListSources } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { useApplicationsCount } from './utils';

const WGApplicationsSources: FC = () => {
  const { data: sources, isLoading, error } = useListSources();
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Application Sources">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          {
            label: 'Sources',
            url: '/sources',
            count: sources?.result?.length,
          },
        ]}
      />
      <ContentWrapper errors={sources?.errors} loading={isLoading}>
        {error && <Alert severity="error">{error.message}</Alert>}
        {sources && <SourcesTable sources={sources?.result} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
