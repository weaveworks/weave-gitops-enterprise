import React, { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper, Title } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  useListFluxRuntimeObjects,
  FluxRuntime,
} from '@weaveworks/weave-gitops';

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading } = useListFluxRuntimeObjects();

  return (
    <PageTemplate documentTitle="WeGO · Flux Runtime">
      <SectionHeader
        path={[
          {
            label: 'Flux Runtime',
            url: '/flux_runtime',
            count: data?.deployments?.length,
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
