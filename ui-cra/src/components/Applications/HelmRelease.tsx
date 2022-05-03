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
};

const WGApplicationsHelmRelease: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const { name, namespace, clusterName } = props;
  const { data, isLoading } = useGetHelmRelease(name, namespace, clusterName);
  const helmRelease = data?.helmRelease;

  if (isLoading) {
    return null;
  }

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
      <ContentWrapper>
        <HelmReleaseDetail helmRelease={helmRelease} {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRelease;
