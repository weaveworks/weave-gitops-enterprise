import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  DataTable,
  filterConfig,
  KubeStatusIndicator,
  filterByStatusCallback,
  statusSortHelper,
  Timestamp,
  formatURL,
} from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import { useListGitOpsSets } from '../../hooks/gitopssets';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import { GitOpsSet, ResourceRef } from '../../api/gitopssets/types.pb';
import { computeMessage } from '../Clusters';
import _ from 'lodash';
import { Routes } from '../../utils/nav';
import { TableWrapper } from '../Shared';

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

const GitopsSets: FC = () => {
  const { isLoading, data } = useListGitOpsSets();

  const gitopssets = data?.gitopssets;

  let initialFilterState = {
    ...filterConfig(gitopssets, 'status', filterByStatusCallback),
    ...filterConfig(gitopssets, 'type'),
    ...filterConfig(gitopssets, 'namespace'),
    ...filterConfig(gitopssets, 'tenant'),
    ...filterConfig(gitopssets, 'clusterName'),
  };

  let fields: Field[] = [
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
        (gs?.conditions && computeMessage(gs?.conditions)) || '',
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
    <PageTemplate
      documentTitle="GitOpsSets"
      path={[
        {
          label: 'GitOpsSets',
        },
      ]}
    >
      <ContentWrapper errors={data?.errors}>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <TableWrapper id="gitopssets-list">
            <DataTable
              fields={fields}
              rows={data?.gitopssets}
              filters={initialFilterState}
            />
          </TableWrapper>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitopsSets;
