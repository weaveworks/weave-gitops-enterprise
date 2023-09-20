import { GetConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { getPullRequestUrl } from '../OpenedPullRequest';
import { GitRepository } from '@weaveworks/weave-gitops';

describe('getPullRequestUrl', () => {
  it('should generate the correct url', () => {
    const urls = {
      // https with .git
      'https://github.com/org/repo.git': 'https://github.com/org/repo/pulls',
      'https://gitlab.com/org/repo.git':
        'https://gitlab.com/org/repo/-/merge_requests',
      'https://dev.azure.com/org/proj/_git/repo.git':
        'https://dev.azure.com/org/proj/_git/repo/pullrequests?_a=active',
      'https://example.com/scm/proj/repo.git':
        'https://example.com/projects/proj/repos/repo/pull-requests',

      // https without .git
      'https://github.com/org/repo': 'https://github.com/org/repo/pulls',
      'https://gitlab.com/org/repo':
        'https://gitlab.com/org/repo/-/merge_requests',
      'https://dev.azure.com/org/proj/_git/repo':
        'https://dev.azure.com/org/proj/_git/repo/pullrequests?_a=active',
      'https://example.com/scm/proj/repo':
        'https://example.com/projects/proj/repos/repo/pull-requests',
      // ssh with .git
      'ssh://git@github.com/org/repo.git': 'https://github.com/org/repo/pulls',
      'ssh://git@gitlab.com/org/repo.git':
        'https://gitlab.com/org/repo/-/merge_requests',
      'ssh://git@ssh.dev.azure.com/v3/org/proj/repo.git':
        'https://dev.azure.com/org/proj/_git/repo/pullrequests?_a=active',
      // ssh without .git
      'ssh://git@github.com/org/repo': 'https://github.com/org/repo/pulls',
      'ssh://git@gitlab.com/org/repo':
        'https://gitlab.com/org/repo/-/merge_requests',
      'ssh://git@ssh.dev.azure.com/v3/org/proj/repo':
        'https://dev.azure.com/org/proj/_git/repo/pullrequests?_a=active',
    };

    Object.entries(urls).forEach(([url, expected]) => {
      const config: GetConfigResponse = {
        gitHostTypes: {
          'gitlab.com': 'gitlab',
          'dev.azure.com': 'azure-devops',
          'ssh.dev.azure.com': 'azure-devops',
          'example.com': 'bitbucket-server',
        },
      };
      expect(
        getPullRequestUrl({ obj: { spec: { url } } } as GitRepository, config),
      ).toEqual(expected);
    });
  });
});
