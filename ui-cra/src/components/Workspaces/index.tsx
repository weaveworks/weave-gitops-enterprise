import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { WorkspacesTable } from './Table';
import { useListWorkspaces } from '../../contexts/Workspaces';
import { Routes } from '../../utils/nav';

const WorkspacesList = () => {
  const { data, isLoading } = useListWorkspaces({});
  return (
    <PageTemplate documentTitle="Workspaces" path={[{ label: 'Workspaces' }]}>
      <ContentWrapper loading={isLoading} errors={data?.errors}>
        {data?.workspaces && <WorkspacesTable workspaces={data.workspaces} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WorkspacesList;
