import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  GetPolicyRequest,
  GetPolicyResponse,
  GetPolicyValidationRequest,
  GetPolicyValidationResponse,
  ListPoliciesRequest,
  ListPoliciesResponse,
  ListPolicyValidationsRequest,
  ListPolicyValidationsResponse,
} from '../../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../EnterpriseClient';

const LIST_POLICIES_QUERY_KEY = 'list-policy';

export function useListListPolicies(req: ListPoliciesRequest) {
  const { api } = useContext(EnterpriseClientContext);

  const allData = useQuery<ListPoliciesResponse, Error>(
    [LIST_POLICIES_QUERY_KEY, req],
    () => api.ListPolicies(req),
  );
  const updateModes = allData?.data?.policies?.map(policy => {
    let audit = '';
    let enforce = '';
    if (policy.modes?.includes('audit')) audit = 'audit';
    if (policy.modes?.includes('admission')) enforce = 'enforce';
    return policy = {...policy, 'audit': audit, 'enforce': enforce}
  });
  if(allData?.data?.policies) allData.data.policies = updateModes
  return allData
}
const GET_POLICY_QUERY_KEY = 'get-policy-details';

export function useGetPolicyDetails(req: GetPolicyRequest) {
  const { api } = useContext(EnterpriseClientContext);

  return useQuery<GetPolicyResponse, Error>([GET_POLICY_QUERY_KEY, req], () =>
    api.GetPolicy(req),
  );
}

const LIST_POLICY_VIOLATION_QUERY_KEY = 'list-policy-violations';

export function useListPolicyValidations(req: ListPolicyValidationsRequest) {
  const { api } = useContext(EnterpriseClientContext);

  return useQuery<ListPolicyValidationsResponse, Error>(
    [LIST_POLICY_VIOLATION_QUERY_KEY, req],
    () => api.ListPolicyValidations(req),
  );
}

const GET_POLICY_VIOLATION_QUERY_KEY = 'get-policy-violation-details';

export function useGetPolicyValidationDetails(req: GetPolicyValidationRequest) {
  const { api } = useContext(EnterpriseClientContext);

  return useQuery<GetPolicyValidationResponse, Error>(
    [GET_POLICY_VIOLATION_QUERY_KEY, req],
    () => api.GetPolicyValidation(req),
  );
}
