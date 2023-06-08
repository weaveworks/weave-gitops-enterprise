import { useContext } from 'react';
import { useQuery } from 'react-query';

import { CoreClientContext } from '@weaveworks/weave-gitops';
import {
  GetPolicyRequest,
  GetPolicyResponse,
  GetPolicyValidationRequest,
  GetPolicyValidationResponse,
  ListPoliciesRequest,
  ListPoliciesResponse,
  ListPolicyValidationsRequest,
  ListPolicyValidationsResponse,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { formatError } from '../../utils/formatters';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../../contexts/Notifications';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';

export const useCoreClientContext = () => useContext(CoreClientContext);

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
  const { api } = useCoreClientContext();
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
  const { api } = useCoreClientContext();

  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetPolicyValidationResponse, RequestError>(
    [GET_POLICY_VIOLATION_QUERY_KEY, req],
    () => api.GetPolicyValidation(req),
    { onError },
  );
}
