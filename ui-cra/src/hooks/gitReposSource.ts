import { useQuery } from 'react-query';
import { ListGitRepositoriesResponse } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';

import { coreClient } from '@weaveworks/weave-gitops';
export function useListGitRepos() {
  return useQuery<ListGitRepositoriesResponse, Error>('gitrepos', () =>
    coreClient.ListGitRepositories({ namespace: '' }),
  );
}
