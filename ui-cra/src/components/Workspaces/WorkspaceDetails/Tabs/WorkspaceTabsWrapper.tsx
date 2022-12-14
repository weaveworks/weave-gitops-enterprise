import React, { FC } from 'react';
import { Box, CircularProgress } from '@material-ui/core';
import { Flex } from '@weaveworks/weave-gitops';
import Alert from '@material-ui/lab/Alert';

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

  return (
    <div style={{ width: '100%' }}>
      {loading && (
        <Box marginTop={4}>
          <Flex wide center>
            <CircularProgress />
          </Flex>
        </Box>
      )}
      {errorMessage && (
        <Alert
          severity="error"
          //   className={classes.alertWrapper}
        >
          {errorMessage}
        </Alert>
      )}
      {!loading && children}
    </div>
  );
};

// alertWrapper: {
//     padding: base,
//     margin: `0 ${base} ${base} ${base}`,
//     borderRadius: '10px',
//   },

export default WorkspaceTabsWrapper;
