import { FC } from 'react';
import {
  DataTable,
  filterConfig,
  formatURL,
} from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../Shared';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { Link } from 'react-router-dom';
import { Routes } from '../../../utils/nav';
import { usePolicyStyle } from '../../Policies/PolicyStyles';

interface Props {
  workspaces: Workspace[];
}

export const WorkspacesTable: FC<Props> = ({ workspaces }) => {
  const classes = usePolicyStyle();

  let initialFilterState = {
    ...filterConfig(workspaces, 'clusterName'),
    ...filterConfig(workspaces, 'name'),
  };

  return (
    <TableWrapper id="workspaces-list" style={{minHeight: 'calc(100vh - 233px)'}}>
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
                className={classes.link}
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
            value: ({namespaces}) => namespaces.join(', '),
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