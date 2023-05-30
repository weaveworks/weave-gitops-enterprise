import styled from 'styled-components';
import { useListTerraformObjects } from '../../contexts/Terraform';
import { TableWrapper } from '../Shared';
import TerraformListTable from './TerraformListTable';
import { Page } from '@weaveworks/weave-gitops';
type Props = {
  className?: string;
};

function TerraformObjectList({ className }: Props) {
  const { isLoading, data, error } = useListTerraformObjects();

  return (
    <Page
      error={error ? [error] : data?.errors || []}
      loading={isLoading}
      path={[
        {
          label: 'Terraform Objects',
        },
      ]}
    >
      <TableWrapper>
        <TerraformListTable rows={data?.objects || []} />
      </TableWrapper>
    </Page>
  );
}

export default styled(TerraformObjectList).attrs({
  className: TerraformObjectList.name,
})``;
