import React, { FC, useCallback, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { request, requestWithCountHeader } from '../../utils/request';
import {
  Clusters,
  DeleteClusterPRRequest,
  GitopsClusterEnriched,
} from './index';
import useNotifications from './../Notifications';
import fileDownload from 'js-file-download';

const CLUSTERS_POLL_INTERVAL = 5000;

const ClustersProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [clusters, setClusters] = useState<GitopsClusterEnriched[]>([]);
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

  const getKubeconfig = useCallback(
    (clusterName: string, filename: string) => {
      request('GET', `v1/clusters/${clusterName}/kubeconfig`, {
        headers: {
          Accept: 'application/octet-stream',
        },
      })
        .then(res => fileDownload(res.message, filename))
        .catch(err =>
          setNotifications([
            { message: { text: err.message }, variant: 'danger' },
          ]),
        );
    },
    [setNotifications],
  );

  const { error, data, isLoading } = useQuery<
    { data: { gitopsClusters: GitopsClusterEnriched[]; total: number } },
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
        notification => error.message === notification.message.text,
      ) === false
    ) {
      setNotifications([
        ...notifications,
        { message: { text: error.message }, variant: 'danger' },
      ]);
    }
  }, [data, error, notifications, setNotifications]);

  return (
    <Clusters.Provider
      value={{
        clusters,
        isLoading,
        count,
        loading,
        selectedClusters,
        setSelectedClusters,
        deleteCreatedClusters,
        getKubeconfig,
      }}
    >
      {children}
    </Clusters.Provider>
  );
};

export default ClustersProvider;
