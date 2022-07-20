import { filterConfig, theme, FilterableTable } from '@weaveworks/weave-gitops';
import styled, { ThemeProvider } from 'styled-components';
import { UnstructuredObject } from '@weaveworks/progressive-delivery';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { TableWrapper } from '../CanaryStyles';

export const ManagedObjectsTable = ({ objects }: { objects: UnstructuredObject[] }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfig(objects, 'component'),
  };
  const CustomFilterableTable = styled(FilterableTable)`
  thead > tr {
    background: ${theme.colors.neutral10};
  }
`;

  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        {objects.length > 0 ? (
          <TableWrapper id="objects-list">
            <CustomFilterableTable
              filters={initialFilterState}
              rows={objects}
              fields={[
                {
                  label: 'Name',
                  value: 'name',
                },
                {
                  label: 'Type',
                  value: (object) => (
                    `${object.groupVersionKind.version}/${object.groupVersionKind.kind}`
                  ),
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
        ) : (
          <p>No data to display</p>
        )}
      </ThemeProvider>
    </div>
  );
};
