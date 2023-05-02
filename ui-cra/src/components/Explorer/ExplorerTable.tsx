import { Box } from '@material-ui/core';
import {
  Flex,
  formatURL,
  Icon,
  IconType,
  Link,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { getKindRoute } from '../../utils/nav';
// @ts-ignore
import { DataTable } from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';

type Props = {
  className?: string;
  onColumnHeaderClick?: (field: Field) => void;
  rows: Object[];
  enableBatchSync?: boolean;
};

function ExplorerTable({
  className,
  onColumnHeaderClick,
  rows,
  enableBatchSync,
}: Props) {
  const r: Object[] = _.map(rows, o => ({
    // Doing some things here to make this work with the DataTable.
    // It handles rendering the sync/pause buttons.
    ...o,
    uid: `${o.cluster}/${o.apiGroup}/${o.kind}/${o.namespace}/${o.name}`,
    clusterName: o.cluster,
    type: o.kind,
  }));

  return (
    <DataTable
      className={className}
      fields={[
        {
          label: 'Name',
          value: (o: Object) => {
            const page = getKindRoute(o?.kind as string);

            const url = formatURL(page, {
              name: o.name,
              namespace: o.namespace,
              clusterName: o.cluster,
            });

            return <Link to={url}>{o.name}</Link>;
          },
          sortValue: () => 'name',
        },
        { label: 'Kind', value: 'kind' },
        { label: 'Namespace', value: 'namespace' },
        { label: 'Cluster', value: 'clusterName' },
        {
          label: 'Status',
          sortValue: () => 'status',
          value: (o: Object) => (
            <Flex align>
              <Box marginRight={1}>
                <Icon
                  size={24}
                  color={
                    o?.status === 'Success'
                      ? 'successOriginal'
                      : 'alertOriginal'
                  }
                  type={
                    o?.status === 'Success'
                      ? IconType.SuccessIcon
                      : IconType.ErrorIcon
                  }
                />
              </Box>

              {o?.status}
            </Flex>
          ),
        },
        { label: 'Message', value: 'message' },
      ]}
      rows={r}
      disableSort
      onColumnHeaderClick={onColumnHeaderClick}
      hasCheckboxes={enableBatchSync}
    />
  );
}

export default styled(ExplorerTable).attrs({ className: ExplorerTable.name })``;
