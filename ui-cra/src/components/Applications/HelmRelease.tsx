import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import { HelmReleaseDetail, useGetHelmRelease } from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  clusterName: string;
  namespace: string;
}

const WGApplicationsBucket: FC<Props> = ({ name, clusterName }) => {
  const applicationsCount = useApplicationsCount();
  const { data } = useGetHelmRelease(name);
  const helmRelease = data?.helmRelease;

  return (
    <PageTemplate documentTitle="WeGO · Helm Release">
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
        <HelmReleaseDetail helmRelease={helmRelease} clusterName={clusterName} name={name} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsBucket;
