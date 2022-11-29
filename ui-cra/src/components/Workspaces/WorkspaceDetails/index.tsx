import {
  useGetWorkspaceDetails,
  useGetWorkspacePolicies,
  useGetWorkspaceRoleBinding,
  useGetWorkspaceRoles,
  useGetWorkspaceServiceAccount,
} from '../../../contexts/Workspaces';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import HeaderSection from './HeaderSection';

const WorkspaceDetails = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const { data: workspaceDetails, isLoading: isWorkspaceLoading } =
    useGetWorkspaceDetails({
      clusterName,
      workspaceName,
    });

  const { data: roles, isLoading: isRolesLoading } = useGetWorkspaceRoles({
    clusterName,
    workspaceName,
  });

  const { data: listRoleBindings, isLoading: isListRoleBindingssLoading } =
    useGetWorkspaceRoleBinding({
      clusterName,
      workspaceName,
    });

  const { data: serviceAccounts, isLoading: isServiceAccountsLoading } =
    useGetWorkspaceServiceAccount({
      clusterName,
      workspaceName,
    });
  const { data: workspacePolicies, isLoading: isWorkspacePoliciesLoading } =
    useGetWorkspacePolicies({
      clusterName,
      workspaceName,
    });

  return (
    <>
      <PageTemplate
        documentTitle="Workspaces"
        path={[
          { label: 'Workspaces', url: Routes.Workspaces },
          { label: workspaceDetails?.name || '' },
        ]}
      >
        <ContentWrapper loading={isWorkspaceLoading}>
          <HeaderSection
            name={workspaceDetails?.name}
            clusterName={workspaceDetails?.clusterName}
            namespaces={workspaceDetails?.namespaces}
          />
        </ContentWrapper>
      </PageTemplate>
    </>
  );
};

export default WorkspaceDetails;
