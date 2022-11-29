import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { WorkspacesTable } from './Table';
import { useListListWorkspaces } from '../../contexts/Workspaces';
import { Routes } from '../../utils/nav';

const WPList = {
  workspaces: [
    {
      name: 'bar-tenant',
      clusterName: 'management',
      namespaces: ['bar-ns', 'foobar-ns'],
    },
    {
      name: 'foo-tenant',
      clusterName: 'management',
      namespaces: ['foo-ns'],
    },
  ],
  total: 2,
  nextPageToken: 'eyJDb250aW51ZVRva2VucyI6eyJtYW5hZ2VtZW50Ijp7IiI6IiJ9fX0K',
  errors: [],
};

const WorkspacesList = () => {
  return (
    <PageTemplate
      documentTitle="Workspaces"
      path={[{ label: 'Workspaces', url: Routes.Workspaces }]}
    >
      <ContentWrapper>
        {WPList?.workspaces && <WorkspacesTable workspaces={WPList.workspaces} />}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WorkspacesList;
