import { useListAutomations, useListSources } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useMemo } from 'react';
import { GitopsCluster } from '../../cluster-services/cluster_services.pb';
import { request } from '../../utils/request';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations(undefined, {});
  return automations?.result?.length || 0;
};

export const useSourcesCount = (): number => {
  const { data: sources } = useListSources(undefined, undefined, {});
  return sources?.result?.length || 0;
};

const toCluster = (clusterName: string): GitopsCluster => {
  const [firstBit, secondBit] = clusterName.split('/');
  const [namespace, name, controlPlane] = secondBit
    ? [firstBit, secondBit, false]
    : ['', firstBit, true];
  return {
    name,
    namespace,
    controlPlane,
  };
};
export const UseClustersWithSources = (): GitopsCluster[] => {
  const { data } = useListSources();
  const clusters = useMemo(() => {
    return _.uniq(data?.result?.map(s => s.clusterName))
      .sort()
      .map(toCluster);
  }, [data]);
  return clusters;
};

export const useIsClusterWithSources = (clusterName: string): boolean => {
  const clusters = UseClustersWithSources();
  return clusters.some((c: GitopsCluster) => c.name === clusterName);
};

export const AddApplicationRequest = ({ ...data }, token: string) => {
  return request('POST', `/v1/enterprise/automations`, {
    body: JSON.stringify(data),
    headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
  });
};
