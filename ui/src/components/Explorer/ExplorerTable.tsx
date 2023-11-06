import { Box } from '@material-ui/core';
import {
  DataTable,
  Flex,
  formatURL,
  Icon,
  IconType,
  Link,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { getKindRoute, Routes } from '../../utils/nav';
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
  linkToObject?: boolean;
};

function ExplorerTable({
  className,
  rows,
  enableBatchSync,
  sortField,
  onColumnHeaderClick,
  extraColumns = [],
  linkToObject = true,
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
        } else if (page === Routes.Templates) {
          url = formatURL(page, {
            search: o.name + '_',
            filters: 'namespace: ' + o.namespace + '_',
          });
        } else {
          url = formatURL(page, {
            name: o.name,
            namespace: o.namespace,
            clusterName: o.cluster,
          });
        }

        return linkToObject ? <Link to={url}>{o.name}</Link> : <>{o.name}</>;
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
      maxWidth: 600,
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

  /* Moving the sync/pause buttons to the left */
  & > div:first-child {
    justify-content: flex-start;
  }
`;
