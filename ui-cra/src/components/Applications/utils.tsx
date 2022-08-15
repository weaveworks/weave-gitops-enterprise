import { useListAutomations, useListSources } from '@weaveworks/weave-gitops';
import { request } from '../../utils/request';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations();

  return automations?.result?.length || 0;
};

export const useSourcesCount = (): number => {
  const { data: sources } = useListSources();

  return sources?.result?.length || 0;
};

export const AddApplicationRequest = ({ ...data }, token: string) => {
  return request('POST', `/v1/enterprise/kustomizations`, {
    body: JSON.stringify(data),
    headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
  });
};
