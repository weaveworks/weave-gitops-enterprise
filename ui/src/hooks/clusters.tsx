import { ReactQueryOptions } from '@weaveworks/weave-gitops/ui/lib/types';
import fileDownload from 'js-file-download';
import { useCallback, useContext, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import {
  GetKubeconfigRequest,
  ListGitopsClustersResponse,
} from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import useNotifications from '../contexts/Notifications';
import {
  GitopsClusterEnriched,
  DeleteClustersPRRequestEnriched,
} from '../types/custom';
import { rawRequest } from '../utils/request';

const CLUSTERS_POLL_INTERVAL = 5000;

const useClusters = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const { notifications, setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const onError = (error: Error) => {
    if (
      error &&
      notifications?.some(
        notification => error.message === notification.message.text,
      ) === false
    ) {
      setNotifications([
        ...notifications,
        { message: { text: error?.message }, severity: 'error' },
      ]);
    }
  };

  const { isLoading, data } = useQuery<ListGitopsClustersResponse, Error>(
    'clusters',
    () => api.ListGitopsClusters({}),
    {
      keepPreviousData: true,
      refetchInterval: CLUSTERS_POLL_INTERVAL,
      onError,
    },
  );

  const clusters = useMemo(
    () => data?.gitopsClusters || [],
    [data],
  ) as GitopsClusterEnriched[];
  const count: number | null = data?.gitopsClusters?.length || null;

  const getCluster = (clusterName: string, namespace: string) =>
    clusters?.find(
      cluster =>
        cluster.name === clusterName && namespace === cluster.namespace,
    ) || null;

  const getDashboardAnnotations = useCallback(
    (cluster: GitopsClusterEnriched) => {
      if (cluster?.annotations) {
        const annotations = Object.entries(cluster?.annotations);
        const dashboardAnnotations: { [key: string]: string } = {};
        for (const [key, value] of annotations) {
          if (key.includes('metadata.weave.works/dashboard.')) {
            const dashboardProvider = key.split(
              'metadata.weave.works/dashboard.',
            )[1];
            dashboardAnnotations[dashboardProvider] = value;
          }
        }
        return dashboardAnnotations;
      }
      return {};
    },
    [],
  );

  const deleteCreatedClusters = useCallback(
    (data: DeleteClustersPRRequestEnriched, token: string | null) => {
      setLoading(true);
      return api
        .CreateDeletionPullRequest(data, {
          headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
        })
        .finally(() => setLoading(false));
    },
    [api],
  );

  return {
    clusters,
    isLoading,
    count,
    loading,
    deleteCreatedClusters,
    getDashboardAnnotations,
    getCluster,
  };
};

export const useListCluster = (
  opts: ReactQueryOptions<ListGitopsClustersResponse, Error> = {
    keepPreviousData: true,
  },
) => {
  const { api } = useContext(EnterpriseClientContext);
  const { notifications, setNotifications } = useNotifications();

  const onError = (error: Error) => {
    if (
      error &&
      notifications?.some(
        notification => error.message === notification.message.text,
      ) === false
    ) {
      setNotifications([
        ...notifications,
        { message: { text: error?.message }, severity: 'error' },
      ]);
    }
  };
  return useQuery<ListGitopsClustersResponse, Error>(
    'clusters',
    () => api.ListGitopsClusters({}),
    {
      ...opts,
      onError,
    },
  );
};

export function useGetKubeconfig() {
  const { setNotifications } = useNotifications();
  const getKubeconfig = useCallback(
    (req: Required<GetKubeconfigRequest>, filename: string) => {
      return rawRequest(
        'GET',
        `/v1/namespaces/${req.namespace}/clusters/${req.name}/kubeconfig`,
        {
          headers: {
            Accept: 'application/octet-stream',
          },
        },
      )
        .then(res => fileDownload(res.message, filename))
        .catch(err =>
          setNotifications([
            { message: { text: err?.message }, severity: 'error' },
          ]),
        );
    },
    [setNotifications],
  );
  return getKubeconfig;
}

export default useClusters;
