import {
  Automation,
  GitRepository,
  Source,
} from '@weaveworks/weave-gitops/ui/lib/objects';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { GetTerraformObjectResponse } from '../../../api/terraform/terraform.pb';
import { GitopsClusterEnriched } from '../../../types/custom';
import { Resource } from '../Edit/EditButton';
import GitUrlParse from 'git-url-parse';

const yamlConverter = require('js-yaml');

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
          (resource as GetTerraformObjectResponse | Pipeline)?.yaml,
        )?.metadata?.annotations?.['templates.weave.works/create-request'];
      default:
        return '';
    }
  };

  return maybeParseJSON(getAnnotation(resource));
};

export const getRepositoryUrl = (repo: GitRepository) => {
  let repositoryUrl = repo?.obj?.spec?.url;
  let parsedUrl = GitUrlParse(repositoryUrl);
  if (parsedUrl?.protocol === 'ssh') {
    parsedUrl.git_suffix = true;
    repositoryUrl = parsedUrl.pathname.replace('//git@', 'https://');
  }
  return repositoryUrl;
};

export function getInitialGitRepo(
  initialUrl: string,
  gitRepos: GitRepository[],
) {
  if (!initialUrl) {
    return null;
  }
  for (var repo of gitRepos) {
    let repoUrl = repo?.obj?.spec?.url;
    if (repoUrl === initialUrl) {
      return repo;
    }
    let parsedRepoUrl = GitUrlParse(repoUrl);
    let parsedInitialUrl = GitUrlParse(initialUrl);
    console.log(parsedInitialUrl, parsedRepoUrl);
    if (parsedRepoUrl?.name === parsedInitialUrl?.name) {
      return repo;
    }
  }
}
