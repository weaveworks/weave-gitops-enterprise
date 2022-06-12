import React, { useEffect, useState } from 'react';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Refresh } from '@material-ui/icons';
import { createStyles, makeStyles } from '@material-ui/styles';
import Alert from '@material-ui/lab/Alert';
import styled from 'styled-components';

interface ILoadingError {
  fetchFn: () => Promise<any>;
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
  align-items: center;
  justify-content: center;
  height:100%
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
  retry: (fn: Promise<any>) => Promise<void>;
}
export const useRequest = (fetchFn: () => Promise<any>): RequestInfo => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');
  const [data, setData] = useState<any>();

  const fetchLoad = (fn: Promise<any>) => {
    setLoading(true);
    setError(false);
    return fn
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
  };

  useEffect(() => {
    fetchLoad(fetchFn());
    return () => {
      setData(null);
    };
  }, [fetchFn]);

  return { loading, error, errorMessage, data, retry: fetchLoad };
};

const LoadingError: React.FC<any> = ({ children, fetchFn }: ILoadingError) => {
  const classes = useStyles();

  // Use the useRequest hook to fetch the data from the server and show the loading spinner while the data is being fetched
  const { loading, error, errorMessage, data, retry } = useRequest(fetchFn);

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
              <span onClick={() => retry(fetchFn())} className={classes.retry}>
                <Refresh />
              </span>
            </FlexStart>
          </Alert>
        </div>
      )}
      {!loading && !error && children({ value: data })}
    </>
  );
};

// export const LoadingErrorRequestInfo: React.FC<any> = ({
//   children,
//   requestInfo: { error, errorMessage, loading, retry },
// }: any) => {
//   const classes = useStyles();
//   return (
//     <>
//       {loading && (
//         <FlexCenter>
//           <LoadingPage />
//         </FlexCenter>
//       )}
//       {!loading && error && (
//         <div>
//           <Alert severity="error">
//             <FlexStart>
//               {errorMessage}
//               <span onClick={retry} className={classes.retry}>
//                 <Refresh />
//               </span>
//             </FlexStart>
//           </Alert>
//         </div>
//       )}
//       {!loading && !error && children}
//     </>
//   );
// };
export default LoadingError;
