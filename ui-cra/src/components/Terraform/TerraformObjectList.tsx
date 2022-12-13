import styled from 'styled-components';
import { useListTerraformObjects } from '../../contexts/Terraform';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { TableWrapper } from '../Shared';
import TerraformListTable from './TerraformListTable';
type Props = {
  className?: string;
};

function TerraformObjectList({ className }: Props) {
  const { isLoading, data, error } = useListTerraformObjects();

  return (
    <PageTemplate
      documentTitle="Terraform"
      path={[
        {
          label: 'Terraform Objects',
          url: '/terraform',
        },
      ]}
    >
      <ContentWrapper
        errors={error ? [error] : data?.errors || []}
        loading={isLoading}
      >
        <TableWrapper>
          <TerraformListTable rows={data?.objects || []} />
        </TableWrapper>
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(TerraformObjectList).attrs({
  className: TerraformObjectList.name,
})``;
