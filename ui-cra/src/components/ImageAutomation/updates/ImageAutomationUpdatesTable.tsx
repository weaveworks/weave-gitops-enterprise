import {
  DataTable,
  filterConfig,
  formatURL,
  KubeStatusIndicator,
  SourceLink,
  Timestamp,
} from '@weaveworks/weave-gitops';
import { Interval } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { Link } from 'react-router-dom';
import { useListImageAutomation } from '../../../contexts/ImageAutomation';
import { Routes } from '../../../utils/nav';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

const ImageAutomationUpdatesTable = () => {
  const { data, isLoading, error } = useListImageAutomation(
    'ImageUpdateAutomation', //Kind.ImageUpdateAutomation
  );
  const initialFilterState = {
    ...filterConfig(data?.objects, 'name'),
  };
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      <DataTable
        filters={initialFilterState}
        rows={data?.objects}
        fields={[
          {
            label: 'Name',
            value: ({ name, namespace, clusterName }) => (
              <Link
                to={formatURL(Routes.ImageAutomationUpdatesDetails, {
                  name: name,
                  namespace: namespace,
                  clusterName: clusterName,
                })}
              >
                {name}
              </Link>
            ),
            textSearchable: true,
            maxWidth: 600,
          },
          {
            label: 'Namespace',
            value: 'namespace',
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
              <Timestamp time={lastAutomationRunTime} />
            ),
          },
        ]}
      />
    </LoadingWrapper>
  );
};
export function showInterval(interval?: Interval): string {
  const parts = [];
  if (!interval) return '--';
  if (interval.hours !== '0') {
    parts.push(`${interval.hours}h`);
  }

  if (interval.minutes !== '0' || parts.length > 0) {
    parts.push(`${interval.minutes}m`);
  }

  if (interval.seconds !== '0' || parts.length > 0) {
    parts.push(`${interval.seconds}s`);
  }

  return parts.join(' ');
}
export default ImageAutomationUpdatesTable;
