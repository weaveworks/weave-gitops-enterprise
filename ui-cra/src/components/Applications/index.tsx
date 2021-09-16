import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { WGContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { Applications } from '@weaveworks/weave-gitops';

const WGApplicationsDashboard: FC = () => {
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
      <WGContentWrapper>
        <Applications />
      </WGContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
