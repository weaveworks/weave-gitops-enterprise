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
import { formatError } from '../../utils/formatters';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../../contexts/Notifications';

const LIST_POLICIES_QUERY_KEY = 'list-policy';

export function useListListPolicies(req: ListPoliciesRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListPoliciesResponse, Error>(
    [LIST_POLICIES_QUERY_KEY, req],
    () => api.ListPolicies(req),
    { onError },
  );
}
const GET_POLICY_QUERY_KEY = 'get-policy-details';

export function useGetPolicyDetails(req: GetPolicyRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetPolicyResponse, Error>(
    [GET_POLICY_QUERY_KEY, req],
    () => api.GetPolicy(req),
    { onError },
  );
}

const LIST_POLICY_VIOLATION_QUERY_KEY = 'list-policy-violations';

export function useListPolicyValidations(req: ListPolicyValidationsRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<ListPolicyValidationsResponse, Error>(
    [LIST_POLICY_VIOLATION_QUERY_KEY, req],
    () => api.ListPolicyValidations(req),
    { onError },
  );
}

const GET_POLICY_VIOLATION_QUERY_KEY = 'get-policy-violation-details';

export function useGetPolicyValidationDetails(req: GetPolicyValidationRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetPolicyValidationResponse, Error>(
    [GET_POLICY_VIOLATION_QUERY_KEY, req],
    () => api.GetPolicyValidation(req),
    { onError },
  );
}
