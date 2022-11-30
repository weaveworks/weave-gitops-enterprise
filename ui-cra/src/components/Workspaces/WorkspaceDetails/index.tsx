import { useGetWorkspaceDetails } from '../../../contexts/Workspaces';
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
