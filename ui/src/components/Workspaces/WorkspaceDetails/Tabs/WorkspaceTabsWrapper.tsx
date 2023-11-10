import { Box, CircularProgress } from '@material-ui/core';
import { AlertListErrors, Flex } from '@weaveworks/weave-gitops';
import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { FC } from 'react';

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
  return (
    <Flex wide>
      {loading && (
        <Flex wide center>
          <Box margin={4}>
            <Flex wide center>
              <CircularProgress size={'2rem'} />
            </Flex>
          </Box>
        </Flex>
      )}
      {(errors?.length || errorMessage) && (
        <AlertListErrors errors={errors || [{ message: errorMessage }]} />
      )}
      {!loading && !errorMessage && children}
    </Flex>
  );
};

export default LoadingWrapper;
