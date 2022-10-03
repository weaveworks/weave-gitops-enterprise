import {
  FluxRuntime,
  useListFluxCrds,
  useListFluxRuntimeObjects,
} from '@weaveworks/weave-gitops-main';
import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading } = useListFluxRuntimeObjects();
  const { data: crds, isLoading: crdsLoading } = useListFluxCrds();
  
  return (
    <PageTemplate
      documentTitle="Flux Runtime"
      path={[
        {
          label: 'Flux Runtime',
          url: '/flux_runtime',
          count: data?.deployments?.length,
        },
      ]}
    >
      <ContentWrapper errors={data?.errors} loading={isLoading}>
        <FluxRuntime deployments={data?.deployments} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsFluxRuntime;
