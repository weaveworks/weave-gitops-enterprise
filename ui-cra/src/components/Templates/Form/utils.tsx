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
import URI from 'urijs';
import { GetConfigResponse } from '../../../cluster-services/cluster_services.pb';

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

export function getRepositoryUrl(repo: GitRepository) {
  // the https url can be overridden with an annotation
  const httpsUrl =
    repo?.obj?.metadata?.annotations?.['weave.works/repo-https-url'];
  if (httpsUrl) {
    return httpsUrl;
  }
  let uri = URI(repo?.obj?.spec?.url);
  if (uri.hostname() === 'ssh.dev.azure.com') {
    uri = azureSshToHttps(uri.toString());
  }
  return uri.protocol('https').port('').userinfo('').toString();
}

function azureSshToHttps(sshUrl: string) {
  const parts = sshUrl.split('/');
  const organization = parts[4];
  const project = parts[5];
  const repository = parts[6];

  const httpsUrl = `https://dev.azure.com/${organization}/${project}/_git/${repository}`;

  return URI(httpsUrl);
}

export function getProvider(repo: GitRepository, config: GetConfigResponse) {
  const url = getRepositoryUrl(repo);
  const domain = URI(url).hostname();
  return config?.gitHostTypes?.[domain] || 'github';
}

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

  // FIXME: return getDefaultGitRepo(gitRepos, config);
  return null;
}

export function getDefaultGitRepo(
  gitRepos: GitRepository[],
  config: GetConfigResponse,
) {
  const annoRepo = gitRepos.find(
    repo =>
      repo?.obj?.metadata?.annotations?.['weave.works/repo-role'] === 'default',
  );
  if (annoRepo) {
    return annoRepo;
  }

  const mainRepo = gitRepos.find(
    repo =>
      repo.clusterName === config.managementClusterName &&
      repo?.obj?.metadata?.name === 'flux-system' &&
      repo?.obj?.metadata?.namespace === 'flux-system',
  );

  if (mainRepo) {
    return mainRepo;
  }

  return gitRepos[0];
}
