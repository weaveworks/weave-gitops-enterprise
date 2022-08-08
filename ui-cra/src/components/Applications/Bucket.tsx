import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount, useSourcesCount } from './utils';
import { BucketDetail } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsBucket: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const sourcesCount = useSourcesCount();

  return (
    <PageTemplate documentTitle="WeGO · Bucket">
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
            count: sourcesCount,
          },
          {
            label: `${props.name}`,
          },
        ]}
      />
      <ContentWrapper>
        <BucketDetail {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsBucket;
