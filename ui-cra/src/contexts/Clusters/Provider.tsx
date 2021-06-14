import React, { FC, useCallback, useState } from 'react';
import { Cluster } from '../../types/kubernetes';
import { requestWithHeaders } from '../../utils/request';
import { useInterval } from '../../utils/use-interval';
import { Clusters } from './index';

const CLUSTERS_POLL_INTERVAL = 5000;

const ClustersProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [disabled, setDisabled] = useState<boolean>(false);
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [order, setOrder] = React.useState<string>('asc');
  const [orderBy, setOrderBy] = React.useState<string>('ClusterStatus');
  const [pageParams, setPageParams] = useState<{
    page: number;
    perPage: number;
  }>({
    page: 0,
    perPage: 10,
  });
  const [count, setCount] = useState<number | null>(null);
  const [error, setError] = React.useState<string | null>(null);

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
    // abort any inflight requests
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
        setError(null);
      })
      .catch(err => {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      })
      .finally(() => {
        setLoading(false);
        setDisabled(false);
        setAbortController(null);
      });
  }, [abortController, clustersParameters]);

  const addCluster = useCallback((data: any) => {
    console.log('addCluster has been called');
  }, []);

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
        error,
        count,
        loading,
        handleRequestSort,
        handleSetPageParams,
        order,
        orderBy,
        addCluster,
      }}
    >
      {/* TODO: Create loader */}
      {loading && !clusters ? 'loader' : children}
    </Clusters.Provider>
  );
};

export default ClustersProvider;
