import { Box } from '@material-ui/core';
import {
  Flex,
  Icon,
  IconType,
  Link,
  V2Routes,
  formatURL,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { Routes, getKindRoute } from '../../utils/nav';
// @ts-ignore
import { DataTable } from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import { QueryState } from './hooks';

export type FieldWithIndex = Field & { index?: number };

type Props = {
  className?: string;
  onColumnHeaderClick?: (field: Field) => void;
  rows: Object[];
  queryState: QueryState;
  enableBatchSync?: boolean;
  sortField?: string;
  extraColumns?: FieldWithIndex[];
};

function ExplorerTable({
  className,
  rows,
  enableBatchSync,
  sortField,
  onColumnHeaderClick,
  extraColumns = [],
}: Props) {
  const r: Object[] = _.map(rows, o => ({
    // Doing some things here to make this work with the DataTable.
    // It handles rendering the sync/pause buttons.
    ...o,
    uid: o.id,
    clusterName: o.cluster,
    type: o.kind,
  }));

  const fields: FieldWithIndex[] = [
    {
      label: 'Name',
      value: (o: Object) => {
        const page = getKindRoute(o?.kind as string);

        let url: string;
        if (page === V2Routes.NotImplemented) {
          url = formatURL(Routes.ExplorerView, {
            kind: o.kind,
            name: o.name,
            namespace: o.namespace,
            clusterName: o.cluster,
          });
        } else if (page == Routes.Templates){
          url = formatURL(page, {
            name: o.name,
            namespace: o.namespace,
          });
        } else {
          url = formatURL(page, {
            name: o.name,
            namespace: o.namespace,
            clusterName: o.cluster,
          });
        }

        return <Link to={url}>{o.name}</Link>;
      },
      sortValue: () => 'name',
      defaultSort: sortField === 'name',
    },
    { label: 'Kind', value: 'kind', defaultSort: sortField === 'kind' },
    {
      label: 'Namespace',
      value: 'namespace',
      defaultSort: sortField === 'namespace',
    },
    {
      label: 'Cluster',
      value: 'clusterName',
      defaultSort: sortField === 'clusterName',
    },
    {
      label: 'Tenant',
      value: 'tenant',
      defaultSort: sortField === 'tenant',
    },
    {
      label: 'Status',
      sortValue: () => 'status',
      defaultSort: sortField === 'status',
      value: (o: Object) => {
        if (o.status === '-') {
          return '-';
        }

        return (
          <Flex align>
            <Box marginRight={1}>
              <Icon
                size={24}
                color={
                  o?.status === 'Success' ? 'successOriginal' : 'alertOriginal'
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
        );
      },
    },
    {
      label: 'Message',
      value: 'message',
      defaultSort: sortField === 'message',
    },
  ];

  // Allows for columns to be added anywhere in the table.
  // We sort them here because if we mutate fields out of order,
  // the column order won't be accurate. See test case for example.
  for (const extra of _.sortBy(extraColumns, 'index')) {
    if (typeof extra.index !== 'undefined') {
      fields.splice(extra.index, 0, extra);
    } else {
      fields.push(extra);
    }
  }

  return (
    <DataTable
      className={className}
      fields={fields}
      rows={r}
      hideSearchAndFilters
      onColumnHeaderClick={onColumnHeaderClick}
      hasCheckboxes={enableBatchSync}
      disableSort
    />
  );
}

export default styled(ExplorerTable).attrs({ className: ExplorerTable.name })`
  width: 100%;

  td:nth-child(6),
  td:nth-child(7) {
    white-space: pre-wrap;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }

  /* Moving the sync/pause buttons to the left */
  & > div:first-child {
    justify-content: flex-start;
  }
`;
