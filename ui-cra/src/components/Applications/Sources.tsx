import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  SourcesTable,
  useListSources,
} from '@weaveworks/weave-gitops';
import { useApplicationsCount } from './utils';

const WGApplicationsSources: FC = () => {
  const { data: sources, isLoading } = useListSources();
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
            count: sources?.length,
          },
        ]}
      />
      <ContentWrapper>
        {isLoading ? <LoadingPage /> : <SourcesTable sources={sources} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
