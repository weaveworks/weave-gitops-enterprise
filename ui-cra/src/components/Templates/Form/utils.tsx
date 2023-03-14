import {
  Automation,
  GitRepository,
  Source,
} from '@weaveworks/weave-gitops/ui/lib/objects';
import GitUrlParse from 'git-url-parse';
import { GitopsClusterEnriched } from '../../../types/custom';
import { Resource } from '../Edit/EditButton';

import yamlConverter from 'js-yaml';
import _ from 'lodash';

export const maybeParseJSON = (data: string) => {
  try {
    return JSON.parse(data);
  } catch (e) {
    // FIXME: show a warning to a user or something
    return undefined;
  }
};

export const getCreateRequestAnnotation = (resource: Resource) => {
  const getAnnotation = (resource: Resource) => {
    switch (resource.type) {
      case 'GitopsCluster':
        return (resource as GitopsClusterEnriched)?.annotations?.[
          'templates.weave.works/create-request'
        ];
      case 'GitRepository':
      case 'Bucket':
      case 'HelmRepository':
      case 'HelmChart':
      case 'Kustomization':
      case 'HelmRelease':
      case 'OCIRepository':
        return (resource as Automation | Source)?.obj?.metadata?.annotations?.[
          'templates.weave.works/create-request'
        ];
      case 'Terraform':
      case 'Pipeline':
        return yamlConverter.load(
          _.get(resource, [
            'yaml',
            'metadata',
            'annotations',
            'templates.weave.works/create-request',
          ]),
        );
      default:
        return '';
    }
  };

  return maybeParseJSON(getAnnotation(resource));
};

export const getRepositoryUrl = (repo: GitRepository) => {
  // the https url can be overridden with an annotation
  const httpsUrl =
    repo?.obj?.metadata?.annotations?.['weave.works/repo-https-url'];
  if (httpsUrl) {
    return httpsUrl;
  }
  let repositoryUrl = repo?.obj?.spec?.url;
  let parsedUrl = GitUrlParse(repositoryUrl);
  if (parsedUrl?.protocol === 'ssh') {
    repositoryUrl = parsedUrl.href.replace('ssh://git@', 'https://');
  }
  // flux does not support "git@github.com:org/repo.git" style urls
  // so we return the original url, the BE handler will fail and return
  // an error to the user
  return repositoryUrl;
};

export function getInitialGitRepo(
  initialUrl: string | null,
  gitRepos: GitRepository[],
) {
  // if there is a repository url in the create-pr annotation, go through the gitrepos and compare it to their links;
  // if no result, parse it and check for the protocol; if ssh, convert it to https and try again to compare it to the gitrepos links
  // createPRRepo signals that this refers to a pre-existing resource
  if (initialUrl) {
    for (var repo of gitRepos) {
      let repoUrl = repo?.obj?.spec?.url;
      if (repoUrl === initialUrl) {
        return { ...repo, createPRRepo: true };
      }
      let parsedRepolUrl = GitUrlParse(repoUrl);
      if (parsedRepolUrl?.protocol === 'ssh') {
        if (
          initialUrl === parsedRepolUrl.href.replace('ssh://git@', 'https://')
        ) {
          return { ...repo, createPRRepo: true };
        }
      }
    }
  }

  const annoRepo = gitRepos.find(
    repo =>
      repo?.obj?.metadata?.annotations?.['weave.works/repo-role'] === 'default',
  );
  if (annoRepo) {
    return annoRepo;
  }

  const mainRepo = gitRepos.find(
    repo =>
      repo?.obj?.metadata?.name === 'flux-system' &&
      repo?.obj?.metadata?.namespace === 'flux-system',
  );
  if (mainRepo) {
    return mainRepo;
  }

  return gitRepos[0];
}
