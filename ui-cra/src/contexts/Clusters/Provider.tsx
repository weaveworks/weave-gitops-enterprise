import React, { FC, useCallback, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import fileDownload from 'js-file-download';
import { request, requestWithCountHeader } from '../../utils/request';
import { Clusters, DeleteClusterPRRequest } from './index';
import useNotifications from './../Notifications';
import { GitopsCluster } from '../../capi-server/capi_server.pb';

const CLUSTERS_POLL_INTERVAL = 5000;

const ClustersProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [clusters, setClusters] = useState<any[]>([]);
  const [count, setCount] = useState<number | null>(null);
  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const { notifications, setNotifications } = useNotifications();

  const clustersBaseUrl = '/v1/clusters';

  const fetchClusters = () =>
    requestWithCountHeader('GET', clustersBaseUrl, {
      cache: 'no-store',
    });

  const deleteCreatedClusters = useCallback(
    (data: DeleteClusterPRRequest, token: string) => {
      setLoading(true);
      return request('DELETE', '/v1/clusters', {
        body: JSON.stringify(data),
        headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
      }).finally(() => setLoading(false));
    },
    [],
  );

  const deleteConnectedClusters = useCallback(
    ({ ...data }) => {
      setLoading(true);
      request('DELETE', `/gitops/api/clusters/${[...data.clusters]}`)
        .then(() =>
          setNotifications([
            {
              message: 'Cluster successfully removed from the MCCP',
              variant: 'success',
            },
          ]),
        )
        .catch(err =>
          setNotifications([{ message: err.message, variant: 'danger' }]),
        )
        .finally(() => setLoading(false));
    },
    [setNotifications],
  );

  const getKubeconfig = useCallback(
    (clusterName: string, filename: string) => {
      request('GET', `v1/clusters/${clusterName}/kubeconfig`, {
        headers: {
          Accept: 'application/octet-stream',
        },
      })
        .then(res => fileDownload(res.message, filename))
        .catch(err =>
          setNotifications([{ message: err.message, variant: 'danger' }]),
        );
    },
    [setNotifications],
  );

  const { error, data } = useQuery<
    { data: { gitopsClusters: GitopsCluster[]; total: number } },
    Error
  >('clusters', () => fetchClusters(), {
    keepPreviousData: true,
    refetchInterval: CLUSTERS_POLL_INTERVAL,
  });

  useEffect(() => {
    if (data) {
      setClusters(data.data.gitopsClusters);
      setCount(data.data.total);
    }
    if (
      error &&
      notifications?.some(
        notification => error.message === notification.message,
      ) === false
    ) {
      setNotifications([
        ...notifications,
        { message: error.message, variant: 'danger' },
      ]);
    }
  }, [data, error, notifications, setNotifications]);

  return (
    <Clusters.Provider
      value={{
        clusters,
        count,
        loading,
        selectedClusters,
        setSelectedClusters,
        deleteCreatedClusters,
        deleteConnectedClusters,
        getKubeconfig,
      }}
    >
      {children}
    </Clusters.Provider>
  );
};

export default ClustersProvider;
