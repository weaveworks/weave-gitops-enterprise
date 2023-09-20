import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import { TableWrapper } from '../../Shared';
import {
  DataTable,
  filterConfig,
  formatURL,
  Link,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';

interface Props {
  workspaces: Workspace[];
}

export const WorkspacesTable: FC<Props> = ({ workspaces }) => {
  const initialFilterState = {
    ...filterConfig(workspaces, 'clusterName'),
    ...filterConfig(workspaces, 'name'),
  };

  return (
    <TableWrapper
      id="workspaces-list"
      style={{ minHeight: 'calc(100vh - 233px)' }}
    >
      <DataTable
        key={workspaces?.length}
        filters={initialFilterState}
        rows={workspaces}
        fields={[
          {
            label: 'Name',
            value: (w: Workspace) => (
              <Link
                to={formatURL(Routes.WorkspaceDetails, {
                  clusterName: w.clusterName,
                  workspaceName: w.name,
                })}
                data-workspace-name={w.name}
              >
                {w.name}
              </Link>
            ),
            textSearchable: true,
            sortValue: ({ name }) => name,
            maxWidth: 650,
          },
          {
            label: 'Namespaces',
            value: ({ namespaces }) => namespaces.join(', '),
          },
          {
            label: 'Cluster',
            value: 'clusterName',
          },
        ]}
      />
    </TableWrapper>
  );
};
