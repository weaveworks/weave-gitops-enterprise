import { GitRepository } from '@weaveworks/weave-gitops';
import { getPullRequestUrl } from '../OpenedPullRequest';
import { GetConfigResponse } from '../../../cluster-services/cluster_services.pb';

describe('getPullRequestUrl', () => {
  it('should generate the correct url', () => {
    const urls = {
      // https with .git
      'https://github.com/org/repo.git': 'https://github.com/org/repo/pulls',
      'https://gitlab.com/org/repo.git':
        'https://gitlab.com/org/repo/-/merge_requests',
      // https without .git
      'https://github.com/org/repo': 'https://github.com/org/repo/pulls',
      'https://gitlab.com/org/repo':
        'https://gitlab.com/org/repo/-/merge_requests',
      // ssh with .git
      'ssh://git@github.com/org/repo.git': 'https://github.com/org/repo/pulls',
      'ssh://git@gitlab.com/org/repo.git':
        'https://gitlab.com/org/repo/-/merge_requests',
      // ssh without .git
      'ssh://git@github.com/org/repo': 'https://github.com/org/repo/pulls',
      'ssh://git@gitlab.com/org/repo':
        'https://gitlab.com/org/repo/-/merge_requests',
    };

    Object.entries(urls).forEach(([url, expected]) => {
      const config: GetConfigResponse = {
        gitHostTypes: {
          'gitlab.com': 'gitlab',
        },
      };
      expect(
        getPullRequestUrl({ obj: { spec: { url } } } as GitRepository, config),
      ).toEqual(expected);
    });
  });
});
