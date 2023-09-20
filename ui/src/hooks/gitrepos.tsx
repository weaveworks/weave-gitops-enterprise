import { GitRepository, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Kind, useListSources } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import React from 'react';


export const getGitRepos = (sources: Source[] | undefined) =>
  _.orderBy(
    _.uniqBy(
      _.filter(
        sources,
        (item): item is GitRepository => item.type === Kind.GitRepository,
      ),
      repo => repo?.obj?.spec?.url,
    ),
    ['name'],
    ['asc'],
  );

export const useGitRepos = () => {
  const { data, error, isLoading } = useListSources('', '', { retry: false });
  const gitRepos = React.useMemo(
    () => getGitRepos(data?.result),
    [data?.result],
  );

  return { gitRepos, error, isLoading };
};
