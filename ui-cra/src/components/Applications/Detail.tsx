import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { WGContentWrapper } from '../Layout/ContentWrapper';
import { ApplicationDetail } from '@weaveworks/weave-gitops';
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
          { label: `${name}` },
        ]}
      />
      <WGContentWrapper>
        <ApplicationDetail name={name || ''} />
      </WGContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationDetail;
