import {
  ListWorkspacesRequest,
  ListWorkspacesResponse,
  GetWorkspaceRequest,
  GetWorkspaceResponse,
  GetWorkspaceRolesResponse,
  GetWorkspaceRoleBindingsResponse,
  GetWorkspaceServiceAccountsResponse,
  GetWorkspacePoliciesResponse,
} from '../../cluster-services/cluster_services.pb';
import { formatError } from '../../utils/formatters';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../../contexts/Notifications';
import { useContext } from 'react';
import { useQuery } from 'react-query';

const LIST_WORKSPACES_QUERY_KEY = 'workspaces';

export function useListWorkspaces(req: ListWorkspacesRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<ListWorkspacesResponse, Error>(
    [LIST_WORKSPACES_QUERY_KEY, req],
    () => api.ListWorkspaces(req),
    { onError },
  );
}

const GET_WORKSPACE_QUERY_KEY = 'workspace-details';

export function useGetWorkspaceDetails(req: GetWorkspaceRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceResponse, Error>(
    [GET_WORKSPACE_QUERY_KEY, req],
    () => api.GetWorkspace(req),
    { onError },
  );
}

const GET_WORKSPACE_Roles_QUERY_KEY = 'workspace-roles';

export function useGetWorkspaceRoles(req: GetWorkspaceRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceRolesResponse, Error>(
    [GET_WORKSPACE_Roles_QUERY_KEY, req],
    () => api.GetWorkspaceRoles(req),
    { onError },
  );
}

const GET_WORKSPACE_Role_Binding_QUERY_KEY = 'workspace-role-binding';

export function useGetWorkspaceRoleBinding(req: GetWorkspaceRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceRoleBindingsResponse, Error>(
    [GET_WORKSPACE_Role_Binding_QUERY_KEY, req],
    () => api.GetWorkspaceRoleBindings(req),
    { onError },
  );
}

const GET_WORKSPACE_Service_Account_QUERY_KEY = 'workspace-service-account';

export function useGetWorkspaceServiceAccount(req: GetWorkspaceRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceServiceAccountsResponse, Error>(
    [GET_WORKSPACE_Service_Account_QUERY_KEY, req],
    () => api.GetWorkspaceServiceAccounts(req),
    { onError },
  );
}

const GET_WORKSPACE_Policies_QUERY_KEY = 'workspace-policies';

export function useGetWorkspacePolicies(req: GetWorkspaceRequest) {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspacePoliciesResponse, Error>(
    [GET_WORKSPACE_Policies_QUERY_KEY, req],
    () => api.GetWorkspacePolicies(req),
    { onError },
  );
}
