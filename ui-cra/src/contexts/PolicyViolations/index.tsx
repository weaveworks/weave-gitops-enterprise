import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  GetPolicyValidationRequest,
  GetPolicyValidationResponse,
  ListPolicyValidationsRequest,
  ListPolicyValidationsResponse,
} from '../../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../EnterpriseClient';


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

const COUNT_POLICY_VIOLATION_QUERY_KEY = 'count-policy-violations';

export function useCountPolicyValidations(req: ListPolicyValidationsRequest) {
  const { api } = useContext(EnterpriseClientContext);

  const { data} = useQuery<ListPolicyValidationsResponse, Error>(
    [COUNT_POLICY_VIOLATION_QUERY_KEY, req],
    () => api.ListPolicyValidations(req),
  );
  return data?.total;
}
