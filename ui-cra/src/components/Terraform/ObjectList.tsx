import { ThemeProvider } from '@material-ui/core';
import { DataTable } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useListTerraformObjects } from '../../contexts/Terraform';
import { localEEMuiTheme } from '../../muiTheme';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { TableWrapper } from '../Shared';

type Props = {
  className?: string;
};

function ObjectList({ className }: Props) {
  const { isLoading, data, error } = useListTerraformObjects();

  if (error) {
    console.error(error);
  }

  console.log(data?.errors);

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Terraform">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Terraform Objects',
              url: '/terraform',
              count: data?.objects?.length,
            },
          ]}
        />

        <ContentWrapper loading={isLoading}>
          <TableWrapper>
            <DataTable
              fields={[{ value: 'name', label: 'Name' }]}
              rows={data?.objects}
            />
          </TableWrapper>
        </ContentWrapper>
      </PageTemplate>
      <div className={className}></div>
    </ThemeProvider>
  );
}

export default styled(ObjectList).attrs({ className: ObjectList.name })``;
