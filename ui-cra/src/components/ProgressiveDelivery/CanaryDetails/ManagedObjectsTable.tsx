import { theme, DataTable } from '@weaveworks/weave-gitops';
import styled, { ThemeProvider } from 'styled-components';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { TableWrapper } from '../CanaryStyles';

const CustomDataTable = styled(DataTable)`
  thead > tr {
    background: ${theme.colors.neutral10};
  }
`;

export const ManagedObjectsTable = ({ objects }: { objects: any[] }) => {
  const classes = usePolicyStyle();

  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        <TableWrapper id="objects-list">
          <CustomDataTable
            rows={objects}
            fields={[
              {
                label: 'Name',
                value: 'name',
              },
              {
                label: 'Type',
                value: object =>
                  `${object.groupVersionKind.version}/${object.groupVersionKind.kind}`,
              },
              {
                label: 'Namespace',
                value: 'namespace',
              },
              {
                label: 'Status',
                value: 'status',
              },
              {
                label: 'Images',
                value: 'images',
              },
            ]}
          />
        </TableWrapper>
      </ThemeProvider>
    </div>
  );
};
