import { GitRepository } from '@weaveworks/weave-gitops';

export interface Pipeline {
  pipelineNamespace: string | undefined;
  pipelineName: string | undefined;
  applicationName: string;

  repo: string | null | GitRepository;
  provider: string;
  branchName: string;
  pullRequestTitle: string;
  commitMessage: string;
  pullRequestDescription: string;
}

export function getPipelineInitialData(
  callbackState: { state: { formData: Pipeline } } | null,
  random: string,
) {
  const defaultFormData = {
    repo: null,
    provider: '',
    branchName: `add-external-secret-branch-${random}`,
    pullRequestTitle: 'Add External Secret',
    commitMessage: 'Add External Secret',
    pullRequestDescription: 'This PR adds a new External Secret',

    pipelineName: '',
    pipelineNamespace: '',
    applicationName: '',
  };

  const initialFormData = {
    ...defaultFormData,
    ...callbackState?.state?.formData,
  };

  return { initialFormData };
}
