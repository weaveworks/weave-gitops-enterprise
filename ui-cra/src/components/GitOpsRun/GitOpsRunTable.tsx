import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { TableWrapper } from '../Shared';

interface Props {
  sessions: any[];
}

export const GitOpsRunTable: FC<Props> = ({ sessions }) => {
  let initialFilterState = {
    ...filterConfig(sessions, 'cliVersion'),
    ...filterConfig(sessions, 'portForward'),
  };

  return (
    <TableWrapper id="policy-list">
      <DataTable
        key={sessions?.length}
        filters={initialFilterState}
        rows={sessions}
        fields={[
          { label: 'Name', value: 'name' },
          { label: 'CLI Version', value: 'cliVersion' },
          { label: 'Port Forward', value: 'portForward' },
          { label: 'Command', value: 'command' },
        ]}
      />
    </TableWrapper>
  );
};
