import { useGetWorkspaceDetails } from '../../../contexts/Workspaces';
import { Routes } from '../../../utils/nav';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import TabDetails from './TabDetails';
import WorkspaceHeaderSection from './WorkspaceHeaderSection';

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
      name: workspaceName,
    });

  return (
    <Page
      loading={isWorkspaceLoading}
      path={[
        { label: 'Workspaces', url: Routes.Workspaces },
        { label: workspaceDetails?.name || '' },
      ]}
    >
      <NotificationsWrapper>
        <WorkspaceHeaderSection
          name={workspaceDetails?.name}
          clusterName={workspaceDetails?.clusterName}
          namespaces={workspaceDetails?.namespaces}
        />
        <TabDetails clusterName={clusterName} workspaceName={workspaceName} />
      </NotificationsWrapper>
    </Page>
  );
};

export default WorkspaceDetails;
