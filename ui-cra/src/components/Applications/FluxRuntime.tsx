import {
  FluxRuntime,
  useListFluxRuntimeObjects,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading } = useListFluxRuntimeObjects();

  return (
    <PageTemplate
      documentTitle="Flux Runtime"
      path={[
        {
          label: 'Flux Runtime',
          url: '/flux_runtime',
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
