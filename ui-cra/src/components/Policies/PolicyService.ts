import { request } from '../../utils/request';

export class PolicyService {
  static policiesUrl = '/v1/policies';

  static getPolicyList = () => {
    return request('GET', this.policiesUrl, {
      cache: 'no-store',
    });
  };
}
