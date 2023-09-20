import { useListImageObjects } from '../../../contexts/ImageAutomation';
import { TableWrapper } from '../../Shared';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import {
  DataTable,
  filterConfig,
  formatURL,
  ImageUpdateAutomation,
  Kind,
  KubeStatusIndicator,
  Link,
  showInterval,
  SourceLink,
  V2Routes,
} from '@weaveworks/weave-gitops';
import moment from 'moment';

const ImageAutomationUpdatesTable = () => {
  const { data, isLoading, error } = useListImageObjects(
    ImageUpdateAutomation,
    Kind.ImageUpdateAutomation,
  );
  const initialFilterState = {
    ...filterConfig(data?.objects, 'name'),
  };
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="image-update-list">
        <DataTable
          filters={initialFilterState}
          rows={data?.objects}
          fields={[
            {
              label: 'Name',
              value: ({ name, namespace, clusterName }) => (
                <Link
                  to={formatURL(V2Routes.ImageAutomationUpdatesDetails, {
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
              value: ({ conditions, suspended }) => (
                <KubeStatusIndicator
                  short
                  conditions={conditions}
                  suspended={suspended}
                />
              ),
              defaultSort: true,
            },
            {
              label: 'Source',
              value: ({ sourceRef, clusterName }) => (
                <SourceLink sourceRef={sourceRef} clusterName={clusterName} />
              ),
            },
            {
              label: 'Interval',
              value: ({ interval }) => showInterval(interval),
            },
            {
              label: 'Last Run',
              value: ({ lastAutomationRunTime }) => (
                <span>
                  {lastAutomationRunTime
                    ? moment(lastAutomationRunTime).fromNow()
                    : ''}
                </span>
              ),
            },
          ]}
        />
      </TableWrapper>
    </LoadingWrapper>
  );
};

export default ImageAutomationUpdatesTable;
