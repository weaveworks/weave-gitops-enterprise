import React, { FC, useCallback, useContext, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { request } from '../../utils/request';
import { Clusters, DeleteClusterPRRequest } from './index';
import useNotifications from './../Notifications';
import fileDownload from 'js-file-download';
import { EnterpriseClientContext } from '../EnterpriseClient';
import { ListGitopsClustersResponse } from '../../cluster-services/cluster_services.pb';
import { GitopsClusterEnriched } from '../../types/custom';

const CLUSTERS_POLL_INTERVAL = 5000;

const ClustersProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [clusters, setClusters] = useState<GitopsClusterEnriched[]>([]);
  const [count, setCount] = useState<number | null>(null);
  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const { notifications, setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const getCluster = (clusterName: string) =>
    clusters?.find(cluster => cluster.name === clusterName) || null;

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
      request('GET', `/v1/clusters/${clusterName}/kubeconfig`, {
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
    ListGitopsClustersResponse,
    Error
  >('clusters', () => api.ListGitopsClusters({}), {
    keepPreviousData: true,
    refetchInterval: CLUSTERS_POLL_INTERVAL,
  });

  useEffect(() => {
    if (data) {
      setClusters(data.gitopsClusters as GitopsClusterEnriched[]);
      setCount(data.total as number);
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
        getDashboardAnnotations,
        getCluster,
      }}
    >
      {children}
    </Clusters.Provider>
  );
};

export default ClustersProvider;
