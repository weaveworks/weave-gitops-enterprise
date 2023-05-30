import { useGetWorkspaceDetails } from '../../../contexts/Workspaces';
import { Routes } from '../../../utils/nav';
import WorkspaceHeaderSection from './WorkspaceHeaderSection';
import TabDetails from './TabDetails';
import { Page } from '@weaveworks/weave-gitops';

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

  return (
    <Page
      loading={isWorkspaceLoading}
      path={[
        { label: 'Workspaces', url: Routes.Workspaces },
        { label: workspaceDetails?.name || '' },
      ]}
    >
      <WorkspaceHeaderSection
        name={workspaceDetails?.name}
        clusterName={workspaceDetails?.clusterName}
        namespaces={workspaceDetails?.namespaces}
      />
      <TabDetails clusterName={clusterName} workspaceName={workspaceName} />
    </Page>
  );
};

export default WorkspaceDetails;
