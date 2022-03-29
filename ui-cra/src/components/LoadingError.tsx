import React, { useEffect, useState } from 'react';
import { LoadingPage } from '@weaveworks/weave-gitops';

export interface ILoadingError {
  fetchFn: () => Promise<any>;
  children?: any;
}

const LoadingError: React.FC<any> = ({ children, fetchFn }: ILoadingError) => {

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [data, setData] = useState<any>();

  const fetchLoad = (fn: Promise<any>) => {
    setLoading(loading => (loading = true));
    setError(err => (err = false));
    return fn
      .then(res => {
        setData(res);
      })
      .catch(err => {
        setError(true);
      })
      .finally(() => {
        setLoading(loading => (loading = false));
      });
  };

  useEffect(() => {
    setLoading(loading => (loading = true));
    setError(false);
    fetchLoad(fetchFn());

    return () => {
      setData(null);
    };
  }, [fetchFn]);

  return (
    <>
      {loading && (
        <div className="flex-center">
          <LoadingPage />
        </div>
      )}
      {!loading && error && (
        <div>
          <div className="alert alert-danger text-center" role="alert">
            Something Went Wrong,
            <span
              onClick={() => fetchLoad(fetchFn())}
              className="ml-medium pointer"
            >
              Retry
            </span>
          </div>
        </div>
      )}
      {!loading && !error && children({ value: data })}
    </>
  );
};

export default LoadingError;
