import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import useNotifications from '../../contexts/Notifications';
import * as React from 'react';
import { QueryClient, useQuery, useQueryClient } from 'react-query';
import { formatError } from '../../utils/formatters';
import {
  GitOpsSets,
  ListGitOpsSetsResponse,
} from '../../api/gitopssets/gitopssets.pb';

const GitOpsSetsContext = React.createContext<typeof GitOpsSets>(
  {} as typeof GitOpsSets,
);

interface Props {
  api: typeof GitOpsSets;
  children?: any;
}

export function GitOpsSetsProvider({ api, children }: Props) {
  return (
    <GitOpsSetsContext.Provider value={api}>
      {children}
    </GitOpsSetsContext.Provider>
  );
}

function useGitOpsSets() {
  return React.useContext(GitOpsSetsContext);
}

const GITOPSSETS_KEY = 'gitopssets';

export function useListGitOpsSets() {
  const { setNotifications } = useNotifications();

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  return useQuery<ListGitOpsSetsResponse, Error>(
    [GITOPSSETS_KEY],
    () => GitOpsSets.ListGitOpsSets({}),
    {
      keepPreviousData: true,
      onError,
    },
  );
}

interface DetailParams {
  name: string;
  namespace: string;
  clusterName: string;
}

function invalidate(
  qc: QueryClient,
  { name, namespace, clusterName }: DetailParams,
) {
  return qc.invalidateQueries([GITOPSSETS_KEY, clusterName, namespace, name]);
}

// export function useSyncGitOpsSet(params: DetailParams) {
//   const gs = useGitOpsSets();
//   const qc = useQueryClient();

//   return () =>
//     gs.SyncGitOpsSet(params).then(res => {
//       invalidate(qc, params);

//       return res;
//     });
// }

export function useToggleSuspendGitOpsSet(params: DetailParams) {
  const gs = useGitOpsSets();
  const qc = useQueryClient();

  return (suspend: boolean) =>
    gs.ToggleSuspendGitOpsSet({ ...params, suspend }).then(res => {
      return invalidate(qc, params).then(() => res);
    });
}
