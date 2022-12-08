import {
  DataTable,
  filterByStatusCallback,
  filterConfig,
  formatURL,
  KubeStatusIndicator,
  Link,
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
  const { data } = useFeatureFlags();
  const flags = data.flags;
  let filterState = {
    ...filterConfig(rows, 'namespace'),
    ...filterConfig(rows, 'Cluster', tf => tf.clusterName),
    ...filterConfig(rows, 'Source', tf => tf.sourceRef.name),
    ...filterConfig(rows, 'Status', filterByStatusCallback),
  };
  if (flags.WEAVE_GITOPS_FEATURE_TENANCY === 'true') {
    filterState = {
      ...filterState,
      ...filterConfig(rows, 'tenant'),
    };
  }
  return (
    <DataTable
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
        ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === 'true'
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
        },
        {
          value: (tf: TerraformObject) => (
            <KubeStatusIndicator
              conditions={tf.conditions || []}
              suspended={tf.suspended}
            />
          ),
          label: 'Status',
        },
        {
          value: (tf: TerraformObject) => computeMessage(tf.conditions as any),
          label: 'Message',
        },
      ]}
      rows={rows}
      filters={filterState}
    />
  );
}

export default styled(TerraformListTable).attrs({
  className: TerraformListTable.name,
})``;
