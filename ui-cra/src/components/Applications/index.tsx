import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { AutomationsTable, useListAutomations } from '@weaveworks/weave-gitops';

const WGApplicationsDashboard: FC = () => {
  const { data: automations } = useListAutomations();
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
      <ContentWrapper type="WG">
        <AutomationsTable automations={automations} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
