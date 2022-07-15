import { Template } from '../cluster-services/cluster_services.pb';

export const isClusterTemplate = (template: Template | null): boolean => {
  return template?.templateKind === 'CAPITemplate';
};

export const textForCreateButton = (template: Template | null): string => {
  return isClusterTemplate(template) ? 'CREATE CLUSTER' : 'CREATE RESOURCES';
};

export const textForCreateHeader = (template: Template | null): string => {
  return isClusterTemplate(template)
    ? 'Create new cluster'
    : 'Create new resources';
};

export const defaultCommitValues = (
  template: Template | null,
  hash: string,
): { [key: string]: string } => {
  return isClusterTemplate(template)
    ? {
        branchName: `create-clusters-branch-${hash}`,
        pullRequestTitle: 'Creates capi cluster',
        commitMessage: 'Creates capi cluster',
        pullRequestDescription: 'This PR creates a new cluster',
      }
    : {
        branchName: `create-resouce-branch-${hash}`,
        pullRequestTitle: 'Creates resources',
        commitMessage: 'Creates resources',
        pullRequestDescription: 'This PR creates a new resource',
      };
};
