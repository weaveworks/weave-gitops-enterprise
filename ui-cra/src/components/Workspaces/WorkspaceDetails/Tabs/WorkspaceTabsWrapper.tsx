import React, { FC } from 'react';
import { Box, CircularProgress } from '@material-ui/core';
import { Flex } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import { useWorkspaceStyle } from '../../WorkspaceStyles';

interface Props {
  loading: boolean;
  errorMessage?: string;
  children: any;
}
const LoadingWrapper: FC<Props> = ({
  children,
  errorMessage,
  loading,
}) => {
  const classes = useWorkspaceStyle();

  return (
    <div className={classes.fullWidth}>
      {loading && (
        <Box margin={4}>
          <Flex wide center>
            <CircularProgress size={'2rem'}/>
          </Flex>
        </Box>
      )}
      {errorMessage && (
        <Alert severity="error" className={classes.alertWrapper}>
          {errorMessage}
        </Alert>
      )}
      {!loading && !errorMessage && children}
    </div>
  );
};

export default LoadingWrapper;
