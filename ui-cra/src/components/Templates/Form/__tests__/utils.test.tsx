import { GitRepository } from '@weaveworks/weave-gitops';
import { getInitialGitRepo, getRepositoryUrl } from '../utils';

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
            'weave.works/repo-https-url': 'https://github.com/org/repo',
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

describe('getInitialGitRepo', () => {
  const gitRepos = [
    {
      obj: {
        spec: {
          url: 'https://github.com/org/repo.git',
        },
      },
    },
    {
      obj: {
        spec: {
          url: 'ssh://git@github.com/org/repo',
        },
      },
    },
    {
      obj: {
        metadata: {
          annotations: {
            'weave.works/repo-rule': 'default',
          },
        },
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    },
    {
      obj: {
        metadata: {
          name: 'flux-system',
          namespace: 'flux-system',
        },
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    },
  ] as GitRepository[];

  it('should return the repo containing the initial url if they are the same', () => {
    const initialUrl = 'https://github.com/org/repo.git';
    expect(getInitialGitRepo(initialUrl, gitRepos)).toStrictEqual({
      obj: {
        spec: {
          url: 'https://github.com/org/repo.git',
        },
      },
      createPRRepo: true,
    });
  });

  it('should return the repo containing the initial url if the initial url is in https format and the gitrepo url is in ssh format', () => {
    const initialUrl = 'https://github.com/org/repo';
    expect(getInitialGitRepo(initialUrl, gitRepos)).toStrictEqual({
      obj: {
        spec: {
          url: 'ssh://git@github.com/org/repo',
        },
      },
      createPRRepo: true,
    });
  });

  it('should return the repo containing the annotation if the initial url repo isnt found and there is a repo with anno', () => {
    const initialUrl = 'https://github.com/test/anno';
    expect(getInitialGitRepo(initialUrl, gitRepos)).toStrictEqual({
      obj: {
        metadata: {
          annotations: {
            'weave.works/repo-rule': 'default',
          },
        },
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    });
  });

  it('should return the repo containing the flux-system combination if the initial url repo isnt found and there is no repo with anno', () => {
    const initialUrl = 'https://github.com/test/fs';
    const gitRepos = [
      {
        obj: {
          spec: {
            url: 'ssh://git@github.com/org/repo',
          },
        },
      },
      {
        obj: {
          metadata: {
            name: 'flux-system',
            namespace: 'flux-system',
          },
          spec: {
            url: 'https://github.com/test/repo.git',
          },
        },
      },
    ] as GitRepository[];
    expect(getInitialGitRepo(initialUrl, gitRepos)).toStrictEqual({
      obj: {
        metadata: {
          name: 'flux-system',
          namespace: 'flux-system',
        },
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    });
  });

  it('should return the repo containing the annotation if there is no initialUrl', () => {
    const initialUrl = '';
    expect(getInitialGitRepo(initialUrl, gitRepos)).toStrictEqual({
      obj: {
        metadata: {
          annotations: {
            'weave.works/repo-rule': 'default',
          },
        },
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    });
  });
});
