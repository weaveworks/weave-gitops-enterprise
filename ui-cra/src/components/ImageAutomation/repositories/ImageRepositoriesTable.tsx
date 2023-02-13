import {
  DataTable,
  filterConfig,
  formatURL,
  Kind,
  KubeStatusIndicator,
} from '@weaveworks/weave-gitops';
import { Source } from '@weaveworks/weave-gitops/ui/lib/types';
import { Link } from 'react-router-dom';
import { useListImageAutomation } from '../../../contexts/ImageAutomation';
import { Routes } from '../../../utils/nav';
import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';
import { showInterval } from '../time';

const ImageRepositoriesTable = () => {
  const { data, isLoading, error } = useListImageAutomation(
    Kind.ImageRepository,
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
                to={formatURL(Routes.ImageAutomationRepositoriesDetails, {
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
            value: (s: Source) => showInterval(s.interval),
          },
          {
            label: 'Tag Count',
            value: 'tagCount',
          },
        ]}
      />
    </LoadingWrapper>
  );
};

export default ImageRepositoriesTable;
