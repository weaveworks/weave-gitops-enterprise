import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  HelmReleaseDetail,
  LoadingPage,
  useGetHelmRelease,
} from '@weaveworks/weave-gitops';

type Props = {
  name: string;
  clusterName: string;
  namespace: string;
};

const WGApplicationsHelmRelease: FC<Props> = props => {
  const applicationsCount = useApplicationsCount();
  const { name, namespace, clusterName } = props;
  const { data, isLoading, error } = useGetHelmRelease(
    name,
    namespace,
    clusterName,
  );
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
          {
            label: `${name}`,
          },
        ]}
      />
      <ContentWrapper>
        <Title>{name}</Title>
        {error && <h3>{error.message}</h3>}
        {isLoading && <LoadingPage />}
        {!error && !isLoading && (
          <HelmReleaseDetail helmRelease={helmRelease} {...props} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRelease;
