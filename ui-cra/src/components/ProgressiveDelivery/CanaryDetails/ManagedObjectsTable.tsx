import { theme, DataTable } from '@weaveworks/weave-gitops';
import { ThemeProvider } from 'styled-components';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { TableWrapper } from '../CanaryStyles';

export const ManagedObjectsTable = ({ objects }: { objects: any[] }) => {
  const classes = usePolicyStyle();

  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        {objects.length > 0 ? (
          <TableWrapper id="objects-list">
            <DataTable
              key={objects?.length}
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
