import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { ResourceRef } from '../../api/gitopssets/types.pb';

type Props = {
  className?: string;
  rows: ResourceRef[];
};

function GitOpsSetInventoryTable({ className, rows }: Props) {
  const filterState = {
    ...filterConfig(rows, 'version'),
  };
  return (
    <DataTable
      className={className}
      rows={rows}
      fields={[
        {
          value: (r: ResourceRef) => r.id as string,
          label: 'Id',
          textSearchable: true,
        },
        {
          value: (r: ResourceRef) => r.version as string,
          label: 'Version',
        },
      ]}
      filters={filterState}
    />
  );
}

export default styled(GitOpsSetInventoryTable).attrs({
  className: GitOpsSetInventoryTable.name,
})``;
