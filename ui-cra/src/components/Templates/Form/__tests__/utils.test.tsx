import { GitRepository } from '@weaveworks/weave-gitops';
import { getRepositoryUrl } from '../utils';

describe('getRepositoryUrl', () => {
  it('should return nil on a git@github.com: style url as flux does not support these', () => {
    const url = 'git@github.com:org/repo.git';
    expect(getRepositoryUrl({ obj: { spec: { url } } } as GitRepository)).toBe(
      url,
    );
  });

  it('should normalize ssh/https urls to https preserving .git if present', () => {
    const urls = {
      // https with .git
      'https://github.com/org/repo.git': 'https://github.com/org/repo.git',
      // https without .git
      'https://github.com/org/repo': 'https://github.com/org/repo',
      // ssh with .git
      'ssh://git@github.com/org/repo.git': 'https://github.com/org/repo.git',
      // ssh without .git
      'ssh://git@github.com/org/repo': 'https://github.com/org/repo',
    };

    Object.entries(urls).forEach(([url, expected]) => {
      expect(
        getRepositoryUrl({ obj: { spec: { url } } } as GitRepository),
      ).toEqual(expected);
    });
  });

  it('should allow you to override the https URL with an annotation present', () => {
    const repo = {
      obj: {
        metadata: {
          annotations: {
            'weave.works/https-url': 'https://github.com/org/repo',
          },
        },
        spec: {
          url: 'ssh://git@internal.cluster/org/repo',
        },
      },
    } as GitRepository;

    expect(getRepositoryUrl(repo)).toEqual('https://github.com/org/repo');
  });
});
