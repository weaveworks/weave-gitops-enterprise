import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { SourcesTable, useListSources } from '@weaveworks/weave-gitops';

const WGApplicationsSources: FC = () => {
  const applicationsCount = useApplicationsCount();
  const { data: sources } = useListSources();

  return (
    <PageTemplate documentTitle="WeGO · Application Sources">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
        ]}
      />
      <ContentWrapper type="WG">
        <SourcesTable sources={sources} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
