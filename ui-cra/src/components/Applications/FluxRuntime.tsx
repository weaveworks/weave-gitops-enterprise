import {
  coreClient,
  FluxRuntime,
  useListFluxRuntimeObjects,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { FC } from 'react';
import { useQuery } from 'react-query';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

export function useListFluxCrds(clusterName = 'management') {
  // Importing the client directly rather than pulling from context.
  // There was an issue with importing from weave-gitops-main conflicting with the official released package.
  // Downside here is duplication and the hook/component is not testable.
  // This should be a temporary fix until the useListFluxCrds hook is exported in an official release.
  return useQuery('flux_crds', () => coreClient.ListFluxCrds({ clusterName }), {
    retry: false,
    refetchInterval: 5000,
  });
}

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading, error } = useListFluxRuntimeObjects();
  const {
    data: crds,
    isLoading: crdsLoading,
    error: crdError,
  } = useListFluxCrds();

  const errors = _.compact([
    ...(data?.errors || []),
    error,
    ...(crds?.errors || []),
    crdError,
  ]);

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
      <ContentWrapper errors={errors as any} loading={isLoading || crdsLoading}>
        <FluxRuntime deployments={data?.deployments} crds={crds?.crds} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsFluxRuntime;
