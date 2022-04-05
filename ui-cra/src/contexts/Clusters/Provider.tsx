import React, { FC, useCallback, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { Cluster } from '../../types/kubernetes';
import { request, requestWithCountHeader } from '../../utils/request';
import { Clusters, DeleteClusterPRRequest } from './index';
import useNotifications from './../Notifications';
import fileDownload from 'js-file-download';

const CLUSTERS_POLL_INTERVAL = 5000;

const ClustersProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [disabled, setDisabled] = useState<boolean>(false);
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [order, setOrder] = useState<string>('asc');
  const [orderBy, setOrderBy] = useState<string>('ClusterStatus');
  const [pageParams, setPageParams] = useState<{
    page: number;
    perPage: number;
  }>({
    page: 0,
    perPage: 10,
  });
  const [count, setCount] = useState<number | null>(null);
  const [selectedClusters, setSelectedClusters] = useState<string[]>([]);
  const { notifications, setNotifications } = useNotifications();

  const handleRequestSort = (property: string) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
    setDisabled(true);
  };

  const handleSetPageParams = (page: number, perPage: number) => {
    setPageParams({ page, perPage });
    setDisabled(true);
  };

  const clustersBaseUrl = '/gitops/api/clusters';

  const fetchClusters = (page: number, perPage: number) =>
    requestWithCountHeader(
      'GET',
      clustersBaseUrl +
        `?sortBy=${orderBy}&order=${order.toUpperCase()}&page=${
          page + 1
        }&per_page=${perPage}`,
      {
        cache: 'no-store',
      },
    ).finally(() => setDisabled(false));

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

  const { error, data } = useQuery(
    ['clusters', pageParams.page, pageParams.perPage],
    () => fetchClusters(pageParams.page, pageParams.perPage),
    {
      keepPreviousData: true,
      refetchInterval: CLUSTERS_POLL_INTERVAL,
    },
  );

  useEffect(() => {
    if (data) {
      setClusters(data.data.clusters);
      setCount(data.data.total);
    }
  }, [data]);

  return (
    <Clusters.Provider
      value={{
        clusters,
        disabled,
        count,
        loading,
        handleRequestSort,
        handleSetPageParams,
        order,
        orderBy,
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
