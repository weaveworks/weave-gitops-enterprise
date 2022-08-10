import { useQuery } from 'react-query';
import {
  Core,
  ListGitRepositoriesResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';

export function useListGitRepos() {
  return useQuery<ListGitRepositoriesResponse, Error>('gitrepos', () =>
    Core.ListGitRepositories({ namespace: '' }),
  );
}
