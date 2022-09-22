import { request } from '../../utils/request';

export const AddApplicationRequest = ({ ...data }, token: string) => {
  return request('POST', `/v1/enterprise/automations`, {
    body: JSON.stringify(data),
    headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
  });
};

export const renderKustomization = (data: any) => {
  return request('POST', `/v1/enterprise/automations/render`, {
    body: JSON.stringify(data),
  });
};
