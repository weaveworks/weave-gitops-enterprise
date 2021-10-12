import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { ApplicationAdd } from '@weaveworks/weave-gitops';
import { useApplicationsCount } from './utils';

const WGApplicationDetail: FC = () => {
  const applicationsCount = useApplicationsCount();

  const queryParams = new URLSearchParams(window.location.search);
  const name = queryParams.get('name');

  return (
    <PageTemplate documentTitle="WeGO · Application Detail">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          { label: 'Add' },
        ]}
      />
      <ContentWrapper type="WG">
        <ApplicationAdd />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationDetail;
