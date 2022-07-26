import { CustomDataTable, TableWrapper } from '../../CanaryStyles';
export const ManagedObjectsTable = ({ objects }: { objects: any[] }) => {
  return (
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
  );
};
