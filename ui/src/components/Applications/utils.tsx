import { useListSources } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useMemo } from 'react';
import { GitopsCluster } from '../../cluster-services/cluster_services.pb';

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

export const useClustersWithSources = (
  allowSelectCluster: boolean,
): GitopsCluster[] | undefined => {
  const { data } = useListSources();
  const clusters = useMemo(() => {
    return _.uniq(data?.result?.map(s => s.clusterName))
      .sort()
      .map(toCluster);
  }, [data]);
  return allowSelectCluster ? clusters : undefined;
};

export const useIsClusterWithSources = (clusterName: string): boolean => {
  const clusters = useClustersWithSources(true);
  return clusters?.some((c: GitopsCluster) => c.name === clusterName) || false;
};
