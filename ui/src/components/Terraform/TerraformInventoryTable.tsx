import { ResourceRef } from '../../api/terraform/types.pb';
import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

type Props = {
  className?: string;
  rows: ResourceRef[];
};

function TerraformInventoryTable({ className, rows }: Props) {
  const filterState = {
    ...filterConfig(rows, 'type'),
  };
  return (
    <DataTable
      className={className}
      rows={rows}
      fields={[
        {
          value: (r: ResourceRef) => r.name as string,
          label: 'Name',
          textSearchable: true,
        },
        {
          value: (r: ResourceRef) => r.type as string,
          label: 'Type',
        },
        {
          value: (r: ResourceRef) => r.identifier as string,
          label: 'Identifier',
        },
      ]}
      filters={filterState}
      emptyMessagePlaceholder='To see the inventory items on this Terraform object set the "spec.enableInventory" to true'
    />
  );
}

export default styled(TerraformInventoryTable).attrs({
  className: TerraformInventoryTable.name,
})``;
