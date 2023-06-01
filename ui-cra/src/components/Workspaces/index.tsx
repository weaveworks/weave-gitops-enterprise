import { WorkspacesTable } from './Table';
import { useListWorkspaces } from '../../contexts/Workspaces';
import { Page } from '@weaveworks/weave-gitops';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

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
