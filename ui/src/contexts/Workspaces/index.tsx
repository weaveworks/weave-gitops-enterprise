import { useQuery } from 'react-query';
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
import { useEnterpriseClient } from '../API';
import useNotifications from './../../contexts/Notifications';

const LIST_WORKSPACES_QUERY_KEY = 'workspaces';

export function useListWorkspaces(req: ListWorkspacesRequest) {
  const { clustersService } = useEnterpriseClient();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<ListWorkspacesResponse, Error>(
    [LIST_WORKSPACES_QUERY_KEY, req],
    () => clustersService.ListWorkspaces(req),
    { onError },
  );
}

const GET_WORKSPACE_QUERY_KEY = 'workspace-details';

export function useGetWorkspaceDetails(req: GetWorkspaceRequest) {
  const { clustersService } = useEnterpriseClient();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceResponse, Error>(
    [GET_WORKSPACE_QUERY_KEY, req],
    () => clustersService.GetWorkspace(req),
    { onError },
  );
}

const GET_WORKSPACE_Roles_QUERY_KEY = 'workspace-roles';

export function useGetWorkspaceRoles(req: GetWorkspaceRequest) {
  const { clustersService } = useEnterpriseClient();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceRolesResponse, Error>(
    [GET_WORKSPACE_Roles_QUERY_KEY, req],
    () => clustersService.GetWorkspaceRoles(req),
    { onError },
  );
}

const GET_WORKSPACE_Role_Binding_QUERY_KEY = 'workspace-role-binding';

export function useGetWorkspaceRoleBinding(req: GetWorkspaceRequest) {
  const { clustersService } = useEnterpriseClient();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceRoleBindingsResponse, Error>(
    [GET_WORKSPACE_Role_Binding_QUERY_KEY, req],
    () => clustersService.GetWorkspaceRoleBindings(req),
    { onError },
  );
}

const GET_WORKSPACE_Service_Account_QUERY_KEY = 'workspace-service-account';

export function useGetWorkspaceServiceAccount(req: GetWorkspaceRequest) {
  const { clustersService } = useEnterpriseClient();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspaceServiceAccountsResponse, Error>(
    [GET_WORKSPACE_Service_Account_QUERY_KEY, req],
    () => clustersService.GetWorkspaceServiceAccounts(req),
    { onError },
  );
}

const GET_WORKSPACE_Policies_QUERY_KEY = 'workspace-policies';

export function useGetWorkspacePolicies(req: GetWorkspaceRequest) {
  const { clustersService } = useEnterpriseClient();
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
  return useQuery<GetWorkspacePoliciesResponse, Error>(
    [GET_WORKSPACE_Policies_QUERY_KEY, req],
    () => clustersService.GetWorkspacePolicies(req),
    { onError },
  );
}
