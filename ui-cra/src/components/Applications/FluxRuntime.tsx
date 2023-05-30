import {
  FluxRuntime,
  useListFluxCrds,
  useListFluxRuntimeObjects,
  Page,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { FC } from 'react';

const WGApplicationsFluxRuntime: FC = () => {
  const { data, isLoading, error } = useListFluxRuntimeObjects();
  const {
    data: crds,
    isLoading: crdsLoading,
    error: crdsError,
  } = useListFluxCrds();

  const errors = _.compact([
    ...(data?.errors || []),
    error,
    ...(crds?.errors || []),
    crdsError,
  ]);

  return (
    <Page
      error={errors}
      loading={isLoading || crdsLoading}
      path={[
        {
          label: 'Flux Runtime',
        },
      ]}
    >
      <FluxRuntime deployments={data?.deployments} crds={crds?.crds} />
    </Page>
  );
};

export default WGApplicationsFluxRuntime;
