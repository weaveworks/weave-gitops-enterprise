import { useListAutomations } from '@weaveworks/weave-gitops';
import { request } from '../../utils/request';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations();

  return automations ? automations?.result?.length : 0;
};
export const AddApplicationRequest = ({ ...data }, token: string) => {
  return request('POST', `/v1/enterprise/kustomizations`, {
    body: JSON.stringify(data),
    headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
  });
};
