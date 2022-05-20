import { request } from '../../utils/request';

export class PolicyService {
  static policiesUrl = '/v1/policies';

  static listPolicies = (payload: any) => {
    return request('GET', this.policiesUrl, {
      cache: 'no-store',
    });
  };
  static getPolicyById = (id: string) => {
    return request('GET', `${this.policiesUrl}/${id}`, {
      cache: 'no-store',
    });
  };

  // TODO payload should be a ClusterId
  static listPolicyViolations = () => {
    return request('POST', `/v1/policyviolations`, {
      cache: 'no-store',
    });
  };

  static getPolicyViolationById = (id: string) => {
    return request('GET', `/v1/policyviolations/${id}`, {
      cache: 'no-store',
    });
  };
}
