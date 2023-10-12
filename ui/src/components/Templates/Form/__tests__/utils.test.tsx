import { renderHook } from '@testing-library/react-hooks';
import { GitRepository } from '@weaveworks/weave-gitops';
import { useGetInitialGitRepo, getRepositoryUrl } from '../utils';

describe('getRepositoryUrl', () => {
  it("should return something, but we don't care what it is as git@github.com: style url as flux does not support these", () => {
    const url = 'git@github.com:org/repo.git';

    expect(
      getRepositoryUrl({ obj: { spec: { url } } } as GitRepository),
    ).toBeTruthy();
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
            'weave.works/repo-role': 'default',
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
    const { result } = renderHook(() =>
      useGetInitialGitRepo(initialUrl, gitRepos),
    );
    expect(result.current).toStrictEqual({
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
    const { result } = renderHook(() =>
      useGetInitialGitRepo(initialUrl, gitRepos),
    );
    expect(result.current).toStrictEqual({
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
    const { result } = renderHook(() =>
      useGetInitialGitRepo(initialUrl, gitRepos),
    );
    expect(result.current).toStrictEqual({
      obj: {
        metadata: {
          annotations: {
            'weave.works/repo-role': 'default',
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
    const { result } = renderHook(() =>
      useGetInitialGitRepo(initialUrl, gitRepos),
    );
    expect(result.current).toStrictEqual({
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
    const { result } = renderHook(() => useGetInitialGitRepo('', gitRepos));
    expect(result.current).toStrictEqual({
      obj: {
        metadata: {
          annotations: {
            'weave.works/repo-role': 'default',
          },
        },
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    });
  });

  it('should return the first repo if nothing matches', () => {
    const repos = [
      {
        obj: {
          spec: {
            url: 'https://github.com/test/repo.git',
          },
        },
      },
    ] as GitRepository[];
    const { result } = renderHook(() => useGetInitialGitRepo('', repos));
    expect(result.current).toStrictEqual({
      obj: {
        spec: {
          url: 'https://github.com/test/repo.git',
        },
      },
    });
  });
});
