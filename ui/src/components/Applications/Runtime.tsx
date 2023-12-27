import {
  FluxRuntime,
  useListRuntimeCrds,
  useListRuntimeObjects,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { FC } from 'react';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const WGApplicationsRuntime: FC = () => {
  const { data, isLoading, error } = useListRuntimeObjects();
  const {
    data: crds,
    isLoading: crdsLoading,
    error: crdsError,
  } = useListRuntimeCrds();

  const errors = _.compact([
    ...(data?.errors || []),
    error,
    ...(crds?.errors || []),
    crdsError,
  ]);

  return (
    <Page
      loading={isLoading || crdsLoading}
      path={[
        {
          label: 'Runtime',
        },
      ]}
    >
      <NotificationsWrapper errors={errors}>
        <FluxRuntime deployments={data?.deployments} crds={crds?.crds} />
      </NotificationsWrapper>
    </Page>
  );
};

export default WGApplicationsRuntime;
