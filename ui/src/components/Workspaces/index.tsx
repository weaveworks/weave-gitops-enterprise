import { useListWorkspaces } from '../../contexts/Workspaces';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { WorkspacesTable } from './Table';

const WorkspacesList = () => {
  const { data, isLoading } = useListWorkspaces({});
  return (
    <Page loading={isLoading} path={[{ label: 'Workspaces' }]}>
      <NotificationsWrapper errors={data?.errors}>
        {data?.workspaces && <WorkspacesTable workspaces={data.workspaces} />}
      </NotificationsWrapper>
    </Page>
  );
};

export default WorkspacesList;
