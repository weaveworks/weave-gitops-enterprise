import {
  DataTable,
  filterByStatusCallback,
  filterConfig,
  formatURL,
  KubeStatusIndicator,
  statusSortHelper,
  Timestamp,
} from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import { FC } from 'react';
import { Link } from 'react-router-dom';
import { GitOpsSet, ResourceRef } from '../../api/gitopssets/types.pb';
import { EnabledComponent } from '../../api/query/query.pb';
import { useListGitOpsSets } from '../../hooks/gitopssets';
import { useIsEnabledForComponent } from '../../hooks/query';
import { computeMessage, Condition } from '../../utils/conditions';
import { Routes } from '../../utils/nav';
import Explorer from '../Explorer/Explorer';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

export const getInventory = (gs: GitOpsSet | undefined) => {
  const entries = gs?.inventory || [];
  return Array.from(
    new Set(
      entries.map((entry: ResourceRef) => {
        // entry is namespace_name_group_kind, but name can contain '_' itself
        const parts = entry?.id?.split('_');
        const kind = parts?.[parts.length - 1];
        const group = parts?.[parts.length - 2];
        return { group, version: entry.version, kind };
      }),
    ),
  );
};

const GitOpsSets: FC = () => {
  const isExplorerEnabled = useIsEnabledForComponent(
    EnabledComponent.gitopssets,
  );
  const { isLoading, data } = useListGitOpsSets({
    enabled: !isExplorerEnabled,
  });

  const gitopssets = data?.gitopssets;

  const initialFilterState = {
    ...filterConfig(gitopssets, 'status', filterByStatusCallback),
    ...filterConfig(gitopssets, 'type'),
    ...filterConfig(gitopssets, 'namespace'),
    ...filterConfig(gitopssets, 'tenant'),
    ...filterConfig(gitopssets, 'clusterName'),
  };

  const fields: Field[] = [
    {
      label: 'Name',
      value: ({ name, namespace, clusterName }: GitOpsSet) => (
        <Link
          to={formatURL(Routes.GitOpsSetDetail, {
            name,
            namespace,
            clusterName,
          })}
        >
          {name}
        </Link>
      ),
      sortValue: ({ name }) => name,
      textSearchable: true,
      maxWidth: 600,
    },
    {
      label: 'Kind',
      value: 'type',
    },
    {
      label: 'Namespace',
      value: 'namespace',
    },
    { label: 'Tenant', value: 'tenant' },
    { label: 'Cluster', value: 'clusterName' },
    {
      label: 'Status',
      value: (gs: GitOpsSet) =>
        gs?.conditions && gs?.conditions?.length > 0 ? (
          <KubeStatusIndicator
            short
            conditions={gs.conditions}
            suspended={false}
          />
        ) : null,
      sortValue: statusSortHelper,
      defaultSort: true,
    },
    {
      label: 'Message',
      value: (gs: GitOpsSet) =>
        (gs?.conditions && computeMessage(gs?.conditions as Condition[])) || '',
      sortValue: ({ conditions }) => computeMessage(conditions),
      maxWidth: 600,
    },
    {
      label: 'Revision',
      maxWidth: 36,
      value: 'lastAppliedRevision',
    },
    {
      label: 'Last Updated',
      value: (gs: GitOpsSet) => (
        <Timestamp
          time={
            _.get(_.find(gs?.conditions, { type: 'Ready' }), 'timestamp') || ''
          }
        />
      ),
      sortValue: (gs: GitOpsSet) => {
        return _.get(_.find(gs.conditions, { type: 'Ready' }), 'timestamp');
      },
    },
  ];

  return (
    <Page
      path={[
        {
          label: 'GitOpsSets',
        },
      ]}
      loading={isLoading}
    >
      <NotificationsWrapper errors={data?.errors}>
        {isExplorerEnabled ? (
          <Explorer category="gitopsset" enableBatchSync={false} />
        ) : (
          <DataTable
            fields={fields}
            rows={data?.gitopssets}
            filters={initialFilterState}
          />
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default GitOpsSets;
