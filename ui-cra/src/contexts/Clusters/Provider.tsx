import React, { FC, useCallback, useState } from 'react';
import { Cluster } from '../../types/kubernetes';
import { request, requestWithHeaders } from '../../utils/request';
import { useInterval } from '../../utils/use-interval';
import { Clusters, DeleteClusterPRRequest } from './index';
import useNotifications from './../Notifications';
import fileDownload from 'js-file-download';

const CLUSTERS_POLL_INTERVAL = 5000;

const ClustersProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [disabled, setDisabled] = useState<boolean>(false);
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);
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
  const { setNotification } = useNotifications();
  const [creatingPR, setCreatingPR] = useState<boolean>(false);
  const { notification } = useNotifications();

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
  const clustersParameters = `?sortBy=${orderBy}&order=${order.toUpperCase()}&page=${
    pageParams.page + 1
  }&per_page=${pageParams.perPage}`;

  const fetchClusters = useCallback(() => {
    abortController?.abort();

    const newAbortController = new AbortController();
    setAbortController(newAbortController);
    setLoading(true);
    requestWithHeaders('GET', clustersBaseUrl + clustersParameters, {
      cache: 'no-store',
      signal: newAbortController.signal,
    })
      .then(res => {
        setCount(res.total);
        setClusters(res.data.clusters);
      })
      .catch(err => {
        if (
          err.name !== 'AbortError' &&
          err.message !== notification?.message
        ) {
          setNotification({ message: err.message, variant: 'danger' });
        }
      })
      .finally(() => {
        setLoading(false);
        setDisabled(false);
        setAbortController(null);
      });
  }, [abortController, clustersParameters, setNotification]);

  const deleteCreatedClusters = useCallback(
    (data: DeleteClusterPRRequest) => {
      setCreatingPR(true);
      request('DELETE', '/v1/clusters', {
        body: JSON.stringify(data),
      })
        .then(res => {
          setNotification({
            message: `PR created successfully`,
            variant: 'success',
          });
        })
        .catch(err =>
          setNotification({ message: err.message, variant: 'danger' }),
        )
        .finally(() => {
          setCreatingPR(false);
          setSelectedClusters([]);
        });
    },
    [setNotification],
  );

  const deleteConnectedClusters = useCallback(
    ({ ...data }) => {
      setLoading(true);
      request('DELETE', `/gitops/api/clusters/${[...data.clusters]}`)
        .then(() =>
          setNotification({
            message: 'Cluster successfully removed from the MCCP',
            variant: 'success',
          }),
        )
        .catch(err =>
          setNotification({ message: err.message, variant: 'danger' }),
        )
        .finally(() => setLoading(false));
    },
    [setNotification],
  );

  const getKubeconfig = useCallback(
    (clusterName: string, filename: string) => {
      setLoading(true);
      request('GET', `v1/clusters/${clusterName}/kubeconfig`, {
        headers: {
          Accept: 'application/octet-stream',
        },
      })
        .then(res => fileDownload(res.message, filename))
        .catch(err =>
          setNotification({ message: err.message, variant: 'danger' }),
        )
        .finally(() => setLoading(false));
    },
    [setNotification],
  );

  useInterval(() => fetchClusters(), CLUSTERS_POLL_INTERVAL, true, [
    order,
    orderBy,
    pageParams.page,
    pageParams.perPage,
  ]);

  return (
    <Clusters.Provider
      value={{
        clusters,
        disabled,
        count,
        loading,
        creatingPR,
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
