import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { WGContentWrapper } from '../Layout/ContentWrapper';
import { Applications } from '@weaveworks/weave-gitops';

// TODO: Uncomment once hook is available from WeGO
// import { useApplications } from '@weaveworks/weave-gitops';

const WGApplicationsDashboard: FC = () => {
  // const { applications } = useApplications();
  // const applicationsCount = applications.length;

  return (
    <PageTemplate documentTitle="WeGO · Applications">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: 1,
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
