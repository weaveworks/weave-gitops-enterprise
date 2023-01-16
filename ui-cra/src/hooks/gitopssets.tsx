import { useState } from 'react';
import { useQuery } from 'react-query';
import useNotifications from '../contexts/Notifications';
import {
  ListGitOpsSetsResponse,
  GitOpsSets,
} from '../api/gitopssets/gitopssets.pb';
import { GitOpsSet } from '../api/gitopssets/types.pb';

const useGitOpsSets = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const { setNotifications } = useNotifications();

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  const { isLoading, data } = useQuery<ListGitOpsSetsResponse, Error>(
    'gitopssets',
    () => GitOpsSets.ListGitOpsSets({}),
    {
      keepPreviousData: true,
      onError,
    },
  );

  const gitopssets = data?.gitopssets as GitOpsSet[] | undefined;

  return {
    isLoading,
    gitopssets,
    loading,
  };
};

export default useGitOpsSets;
