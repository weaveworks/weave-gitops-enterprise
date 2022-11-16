import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { TableWrapper } from '../Shared';

interface Props {
  sessions: any[];
}

export const GitOpsRunTable: FC<Props> = ({ sessions }) => {
  let initialFilterState = {
    ...filterConfig(
      sessions,
      'CLI Version',
      session =>
        session.obj.metadata.annotations['run.weave.works/cli-version'],
    ),
    ...filterConfig(
      sessions,
      'Port Forward',
      session =>
        session.obj.metadata.annotations['run.weave.works/port-forward'],
    ),
  };

  return (
    <TableWrapper id="gitopsRun-list">
      <DataTable
        key={sessions?.length}
        filters={initialFilterState}
        rows={sessions}
        fields={[
          { label: 'Name', value: 'name', textSearchable: true },
          {
            label: 'CLI Version',
            value: ({ obj }) =>
              obj.metadata.annotations['run.weave.works/cli-version'],
          },
          {
            label: 'Port Forward',
            value: ({ obj }) =>
              obj.metadata.annotations['run.weave.works/port-forward'],
          },
          {
            label: 'Command',
            value: ({ obj }) =>
              obj.metadata.annotations['run.weave.works/command'],
          },
        ]}
      />
    </TableWrapper>
  );
};
