import {
  FluxRuntime,
  useListFluxCrds,
  useListFluxRuntimeObjects,
  Page,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';

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
      path={[
        {
          label: 'Flux Runtime',
        },
      ]}
    >
      <ContentWrapper errors={errors} loading={isLoading || crdsLoading}>
        <FluxRuntime deployments={data?.deployments} crds={crds?.crds} />
      </ContentWrapper>
    </Page>
  );
};

export default WGApplicationsFluxRuntime;
