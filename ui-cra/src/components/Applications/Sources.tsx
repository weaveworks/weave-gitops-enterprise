import React, { FC } from 'react';
import styled from 'styled-components';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  SourcesTable,
  useListSources,
} from '@weaveworks/weave-gitops';

const SourcesTableWrapper = styled(SourcesTable)`
  div[class*='FilterDialog__SlideContainer'] {
    overflow: hidden;
  }
  max-width: calc(100vw - 220px - 76px);
`;

const WGApplicationsSources: FC = () => {
  const { data: sources, isLoading } = useListSources();

  return (
    <PageTemplate documentTitle="WeGO · Application Sources">
      <SectionHeader
        path={[
          {
            label: 'Sources',
            url: '/sources',
            count: sources?.length,
          },
        ]}
      />
      <ContentWrapper>
        <Title>Sources</Title>
        {isLoading ? <LoadingPage /> : <SourcesTableWrapper sources={sources} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsSources;
