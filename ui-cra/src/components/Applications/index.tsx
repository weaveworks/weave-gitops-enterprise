import React, { FC } from 'react';
import styled from 'styled-components';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  AutomationsTable,
  LoadingPage,
  useListAutomations,
} from '@weaveworks/weave-gitops';

const AutomationsTableWrapper = styled(AutomationsTable)`
  div[class*='FilterDialog__SlideContainer'] {
    overflow: hidden;
  }
  max-width: calc(100vw - 220px - 76px);
`;

const WGApplicationsDashboard: FC = () => {
  const { data: automations, isLoading } = useListAutomations();
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Applications">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
        ]}
      />
      <ContentWrapper>
        <Title>Applications</Title>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <AutomationsTableWrapper automations={automations} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
