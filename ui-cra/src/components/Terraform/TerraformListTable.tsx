import {
  DataTable,
  filterByStatusCallback,
  filterConfig,
  formatURL,
  KubeStatusIndicator,
  Link,
  statusSortHelper,
  useFeatureFlags,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { TerraformObject } from '../../api/terraform/types.pb';
import { computeMessage } from '../../utils/conditions';
import { getKindRoute, Routes } from '../../utils/nav';

type Props = {
  className?: string;
  rows?: TerraformObject[];
};

function TerraformListTable({ className, rows }: Props) {
  const { isFlagEnabled } = useFeatureFlags();
  let filterState = {
    ...filterConfig(rows, 'namespace'),
    ...filterConfig(rows, 'Cluster', tf => tf.clusterName),
    ...filterConfig(rows, 'Source', tf => tf.sourceRef.name),
    ...filterConfig(rows, 'Status', filterByStatusCallback),
  };
  if (isFlagEnabled('WEAVE_GITOPS_FEATURE_TENANCY')) {
    filterState = {
      ...filterState,
      ...filterConfig(rows, 'tenant'),
    };
  }

  const kindRows = rows?.map(row => {
    return { ...row, type: 'Terraform' };
  });

  return (
    <DataTable
      hasCheckboxes
      className={className}
      fields={[
        {
          value: ({ name, namespace, clusterName }: TerraformObject) => (
            <Link
              to={formatURL(Routes.TerraformDetail, {
                name,
                namespace,
                clusterName,
              })}
            >
              {name}
            </Link>
          ),
          label: 'Name',
          sortValue: ({ name }: TerraformObject) => name,
          textSearchable: true,
        },
        { value: 'namespace', label: 'Namespace' },
        ...(isFlagEnabled('WEAVE_GITOPS_FEATURE_TENANCY')
          ? [{ label: 'Tenant', value: 'tenant' }]
          : []),
        { value: 'clusterName', label: 'Cluster' },
        {
          label: 'Source',
          value: (tf: TerraformObject) => {
            const route = getKindRoute(tf.sourceRef?.kind as string);

            const { name, namespace } = tf.sourceRef || {};

            const u = formatURL(route, {
              clusterName: tf.clusterName,
              name,
              namespace,
            });

            return <Link to={u}>{name}</Link>;
          },
          sortValue: ({ sourceRef }: TerraformObject) => sourceRef?.name,
        },
        {
          value: (tf: TerraformObject) => (
            <KubeStatusIndicator
              conditions={tf.conditions || []}
              suspended={tf.suspended}
            />
          ),
          label: 'Status',
          sortValue: statusSortHelper,
        },
        {
          value: (tf: TerraformObject) => computeMessage(tf.conditions as any),
          label: 'Message',
        },
      ]}
      rows={kindRows}
      filters={filterState}
    />
  );
}

export default styled(TerraformListTable).attrs({
  className: TerraformListTable.name,
})``;
