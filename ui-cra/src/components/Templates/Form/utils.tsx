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
  // the https url can be overridden with an annotation
  const httpsUrl =
    repo?.obj?.metadata?.annotations?.['weave.works/repo-https-url'];
  if (httpsUrl) {
    return httpsUrl;
  }
  let repositoryUrl = repo?.obj?.spec?.url;
  let parsedUrl = GitUrlParse(repositoryUrl);
  // flux does not support "git@github.com:org/repo.git" style urls
  // so we return the original url, the BE handler will fail and return
  // an error to the user
  if (parsedUrl?.protocol === 'file') {
    return repositoryUrl;
  }
  return parsedUrl.toString('https');
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
    if (parsedRepoUrl?.name === parsedInitialUrl?.name) {
      return repo;
    }
  }
}
