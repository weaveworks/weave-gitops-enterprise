import React, { useEffect, useState, useCallback } from 'react';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Refresh } from '@material-ui/icons';
import { createStyles, makeStyles } from '@material-ui/styles';
import Alert from '@material-ui/lab/Alert';
import styled from 'styled-components';

export interface ILoadingError {
  requestInfo: RequestInfo;
  children?: any;
}

const useStyles = makeStyles(() =>
  createStyles({
    retry: {
      marginLeft: '4px',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
  }),
);

const FlexCenter = styled.div`
  display: flex;
  lign-items: center;
  justify-content: center;
`;

const FlexStart = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
`;

export interface RequestInfo {
  loading: boolean;
  error: boolean;
  errorMessage: string;
  data: any;
  retry: () => Promise<any>;
}

export const useRequest = (fetchFn: () => Promise<any>): RequestInfo => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [data, setData] = useState<any>();

  const fetchLoad = useCallback(() => {
    setLoading(true);
    setError(false);
    return fetchFn()
      .then(res => {
        setData(res);
      })
      .catch(err => {
        setErrorMessage(err.message || 'Something Went wrong');
        setError(true);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [fetchFn]);

  useEffect(() => {
    setLoading(true);
    setError(false);
    fetchLoad();

    return () => {
      setData(null);
    };
  }, [fetchLoad]);

  return { retry: fetchLoad, error, errorMessage, data, loading };
};

const LoadingError: React.FC<any> = ({
  children,
  requestInfo: { error, errorMessage, loading, retry },
}: ILoadingError) => {
  const classes = useStyles();
  return (
    <>
      {loading && (
        <FlexCenter>
          <LoadingPage />
        </FlexCenter>
      )}
      {!loading && error && (
        <div>
          <Alert severity="error">
            <FlexStart>
              {errorMessage}
              <span onClick={retry} className={classes.retry}>
                <Refresh />
              </span>
            </FlexStart>
          </Alert>
        </div>
      )}
      {!loading && !error && children}
    </>
  );
};

export default LoadingError;
