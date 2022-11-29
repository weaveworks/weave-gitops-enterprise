import { FC } from 'react';
import {
  DataTable,
  filterConfig,
} from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../Shared';

interface Props {
  workspaces: any[];
}

export const WorkspacesTable: FC<Props> = ({ workspaces }) => {
  // const classes = usePolicyStyle();

  let initialFilterState = {
    ...filterConfig(workspaces, 'clusterName'),
    ...filterConfig(workspaces, 'name'),
  };

  return (
    <TableWrapper id="policy-list">
      <DataTable
        key={workspaces?.length}
        filters={initialFilterState}
        rows={workspaces}
        fields={[
          {
            label: 'Name',
            value: 'name',
            textSearchable: true,
            sortValue: ({ name }) => name,
            maxWidth: 650,
          },
          {
            label: 'Namespaces',
            value: 'namespaces',
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
