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
const WorkspaceTabsWrapper: FC<Props> = ({
  children,
  errorMessage,
  loading,
}) => {
  const classes = useWorkspaceStyle();

  return (
    <div className={classes.fullWidth}>
      {loading && (
        <Box marginTop={4}>
          <Flex wide center>
            <CircularProgress />
          </Flex>
        </Box>
      )}
      {errorMessage && (
        <Alert severity="error" className={classes.alertWrapper}>
          {errorMessage}
        </Alert>
      )}
      {!loading && children}
    </div>
  );
};

export default WorkspaceTabsWrapper;
