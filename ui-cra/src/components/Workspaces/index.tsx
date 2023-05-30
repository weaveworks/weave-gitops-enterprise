import { WorkspacesTable } from './Table';
import { useListWorkspaces } from '../../contexts/Workspaces';
import { Page } from '@weaveworks/weave-gitops';

const WorkspacesList = () => {
  const { data, isLoading } = useListWorkspaces({});
  return (
    <Page
      loading={isLoading}
      error={data?.errors}
      path={[{ label: 'Workspaces' }]}
    >
      {data?.workspaces && <WorkspacesTable workspaces={data.workspaces} />}
    </Page>
  );
};

export default WorkspacesList;
