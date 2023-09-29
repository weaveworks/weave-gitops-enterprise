import { Source } from '@weaveworks/weave-gitops/ui/lib/types';
import {
  DataTable,
  filterConfig,
  formatURL,
  ImageRepository,
  Kind,
  KubeStatusIndicator,
  showInterval,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import { useListImageObjects } from '../../../contexts/ImageAutomation';
import { TableWrapper } from '../../Shared';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ImageRepositoriesTable = () => {
  const { data, isLoading, error } = useListImageObjects(
    ImageRepository,
    Kind.ImageRepository,
  );
  const initialFilterState = {
    ...filterConfig(data?.objects, 'name'),
  };
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      {data?.objects && (
        <TableWrapper id="image-repository-list">
          <DataTable
            filters={initialFilterState}
            rows={data?.objects}
            fields={[
              {
                label: 'Name',
                value: ({ name, namespace, clusterName }) => (
                  <Link
                    to={formatURL(V2Routes.ImageAutomationRepositoryDetails, {
                      name: name,
                      namespace: namespace,
                      clusterName: clusterName,
                    })}
                  >
                    {name}
                  </Link>
                ),
                textSearchable: true,
                sortValue: ({ name }) => name || '',
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
                    conditions={s.conditions || []}
                    suspended={s.suspended}
                  />
                ),
                defaultSort: true,
              },
              {
                label: 'Interval',
                value: ({ interval }) => showInterval(interval),
              },
              {
                label: 'Tag Count',
                value: 'tagCount',
              },
            ]}
          />
        </TableWrapper>
      )}
    </LoadingWrapper>
  );
};

export default ImageRepositoriesTable;
