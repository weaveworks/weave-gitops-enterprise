import React, { useEffect, useState } from 'react';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Refresh } from '@material-ui/icons';
import { createStyles, makeStyles } from '@material-ui/styles';
import Alert from '@material-ui/lab/Alert';

export interface ILoadingError {
  fetchFn: () => Promise<any>;
  children?: any;
}


const useStyles = makeStyles(() =>
  createStyles({
    retry :{
    marginLeft: '4px',
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    },
  })
  );



const LoadingError: React.FC<any> = ({ children, fetchFn }: ILoadingError) => {
  const classes = useStyles();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [data, setData] = useState<any>();

  const fetchLoad = (fn: Promise<any>) => {
    setLoading(loading => (loading = true));
    setError(err => (err = false));
    return fn
      .then(res => {
        setData(res);
      })
      .catch(err => {
        setErrorMessage(err.message || 'Something Went wrong');
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
        <div className="flex-center" >
          <LoadingPage />
        </div>
      )}
      {!loading && error && (
        <div>

      <Alert severity="error" > 
        <div className="flex-start">
        {errorMessage}
              <span onClick={() => fetchLoad(fetchFn())} className={classes.retry}>
                <Refresh />
              </span>   
          </div>
        </Alert>
   
     
        </div>
      )}
      {!loading && !error && children({ value: data })}
    </>
  );
};

export default LoadingError;
