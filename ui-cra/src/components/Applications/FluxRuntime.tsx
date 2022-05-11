import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  useListFluxRuntimeObjects,
  FluxRuntime,
} from '@weaveworks/weave-gitops';
import { useApplicationsCount } from './utils';

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading } = useListFluxRuntimeObjects();
  const applicationsCount = useApplicationsCount();

  return (
    <PageTemplate documentTitle="WeGO · Flux Runtime">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
          {
            label: 'Flux Runtime',
            url: '/flux_runtime',
          },
        ]}
      />
      <ContentWrapper>
        <Title>Flux Runtime</Title>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <FluxRuntime deployments={data?.deployments} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsFluxRuntime;
