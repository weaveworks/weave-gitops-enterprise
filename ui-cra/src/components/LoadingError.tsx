import React, { useEffect, useState } from 'react';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Refresh } from '@material-ui/icons';
import { ErrorOutline } from '@material-ui/icons';
import { createStyles, makeStyles } from '@material-ui/styles';

export interface ILoadingError {
  fetchFn: () => Promise<any>;
  children?: any;
}


const useStyles = makeStyles(() =>
  createStyles({
  alertDanger: {
    borderColor: '#f5c2c7',
    color:'rgb(97, 26, 21)',
    backgroundColor: 'rgb(253, 236, 234)',
  },
  alert: {
    border: `1px solid transparent`,
    borderRadius: '0.25rem',
    marginBottom: '1rem',
    padding: '1rem',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'start',
  },
    retry :{
    marginLeft: '4px',
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    },
    
  alertIcon :{
    marginRight: '8px',
    color:'#f44336',
  }
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
          <div className={`${classes.alertDanger} ${classes.alert}`} role="alert">
          <ErrorOutline className={classes.alertIcon} />
            {errorMessage}
            <span onClick={() => fetchLoad(fetchFn())} className={classes.retry}>
              <Refresh />
            </span>
          </div>
        </div>
      )}
      {!loading && !error && children({ value: data })}
    </>
  );
};

export default LoadingError;
