import {
  FluxRuntime,
  useListFluxRuntimeObjects,
  useListFluxCrds,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading, error } = useListFluxRuntimeObjects();
  const { data: crds, isLoading: crdsLoading, error: crdsError } = useListFluxCrds();

  const errors = _.compact([
    ...(data?.errors || []),
    error,
    ...(crds?.errors || []),
    crdsError,
  ]);

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
      <ContentWrapper errors={errors} loading={isLoading || crdsLoading}>
        <FluxRuntime deployments={data?.deployments} crds={crds?.crds} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsFluxRuntime;
