import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  SourcesTable,
  useListSources,
} from '@weaveworks/weave-gitops';
import { useApplicationsCount } from './utils';
import { Alert } from '@material-ui/lab';

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
      <ContentWrapper errors={sources?.errors}>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <SourcesTable sources={sources?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
