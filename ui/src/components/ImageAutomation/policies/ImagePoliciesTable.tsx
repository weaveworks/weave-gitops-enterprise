import { ImgPolicy, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import {
  DataTable,
  filterConfig,
  formatURL,
  ImagePolicy,
  Kind,
  KubeStatusIndicator,
  Link,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { useListImageObjects } from '../../../contexts/ImageAutomation';
import { TableWrapper } from '../../Shared';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ImagePoliciesTable = () => {
  const { data, isLoading, error } = useListImageObjects(
    ImagePolicy,
    Kind.ImagePolicy,
  );
  const initialFilterState = {
    ...filterConfig(data?.objects, 'name'),
    ...filterConfig(data?.objects, 'imageRepositoryRef'),
  };
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="image-policy-list">
        <DataTable
          filters={initialFilterState}
          rows={data?.objects}
          fields={[
            {
              label: 'Name',
              value: ({ name, namespace, clusterName }) => (
                <Link
                  to={formatURL(V2Routes.ImagePolicyDetails, {
                    name: name,
                    namespace: namespace,
                    clusterName: clusterName,
                  })}
                >
                  {name}
                </Link>
              ),
              sortValue: ({ name }) => name || '',
              textSearchable: true,
              maxWidth: 600,
            },
            {
              label: 'Namespace',
              value: 'namespace',
            },
            {
              label: 'Cluster',
              value: 'clusterName',
            },
            {
              label: 'Status',
              value: (s: Source) => (
                <KubeStatusIndicator
                  short
                  conditions={s.conditions}
                  suspended={s.suspended}
                />
              ),
              defaultSort: true,
            },
            {
              label: 'Image Policy',
              value: ({ imagePolicy }: { imagePolicy: ImgPolicy }) =>
                imagePolicy?.type || '',
            },
            {
              label: 'Order/Range',
              value: ({ imagePolicy }: { imagePolicy: ImgPolicy }) =>
                imagePolicy?.value || '',
            },
            {
              label: 'Image Repository',
              value: ({ imageRepositoryRef, namespace, clusterName }) => (
                <Link
                  to={formatURL(V2Routes.ImageAutomationRepositoryDetails, {
                    name: imageRepositoryRef,
                    namespace: namespace,
                    clusterName: clusterName,
                  })}
                >
                  {imageRepositoryRef}
                </Link>
              ),
            },
          ]}
        />
      </TableWrapper>
    </LoadingWrapper>
  );
};

export default ImagePoliciesTable;
