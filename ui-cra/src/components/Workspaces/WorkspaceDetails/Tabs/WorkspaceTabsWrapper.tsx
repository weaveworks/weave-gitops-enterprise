import React, { FC } from 'react';
import { Box, CircularProgress } from '@material-ui/core';
import { Flex } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';
import { useWorkspaceStyle } from '../../WorkspaceStyles';
import { AlertListErrors } from '../../../Layout/AlertListErrors';
import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';

interface Props {
  loading: boolean;
  errorMessage?: string;
  children: any;
  errors?: ListError[];
}
const LoadingWrapper: FC<Props> = ({
  children,
  errors,
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
      {errors && <AlertListErrors errors={errors} />}
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
