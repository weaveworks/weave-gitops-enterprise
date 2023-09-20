import { TableWrapper } from '../../../Shared';
import CanaryStatus from '../../SharedComponent/CanaryStatus';
import { UnstructuredObject } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { DataTable, Link } from '@weaveworks/weave-gitops';
export const ManagedObjectsTable = ({
  objects,
}: {
  objects: UnstructuredObject[];
}) => {
  return (
    <TableWrapper id="objects-list">
      <DataTable
        rows={objects}
        fields={[
          {
            label: 'Name',
            value: 'name',
          },
          {
            label: 'Type',
            value: object => `${object.groupVersionKind.kind}`,
          },
          {
            label: 'Namespace',
            value: 'namespace',
          },
          {
            label: 'Status',
            value: ({ conditions }: UnstructuredObject) =>
              !!conditions && conditions.length > 0 ? (
                <CanaryStatus status={conditions[0].type || ''} />
              ) : (
                '--'
              ),
          },
          {
            label: 'Message',
            value: ({ conditions }: UnstructuredObject) =>
              !!conditions && conditions.length > 0
                ? `${conditions[0].message}`
                : '--',
          },
          {
            label: 'Images',
            value: ({ images }: UnstructuredObject) => (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                }}
              >
                {images?.map((image, indx) => (
                  <Link newTab href={`https://${image}`} key={indx}>
                    {image}
                  </Link>
                ))}
              </div>
            ),
          },
        ]}
      />
    </TableWrapper>
  );
};
