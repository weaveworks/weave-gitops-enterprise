import { Box } from '@material-ui/core';
import {
  DataTable,
  Flex,
  formatURL,
  Icon,
  Link,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { getKindRoute, Routes } from '../../utils/nav';
import { getIndicatorInfo } from '../../utils/status';
import { QueryState } from './hooks';

export type ExplorerField = Field & {
  id: string;
  index?: number;
};

type Props = {
  className?: string;
  fields: ExplorerField[];
  onColumnHeaderClick?: (field: Field) => void;
  rows: Object[];
  queryState: QueryState;
  enableBatchSync?: boolean;
};

export const defaultExplorerFields: ExplorerField[] = [
  {
    id: 'name',
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

      return <Link to={url}>{o.name}</Link>;
    },
  },
  {
    id: 'kind',
    label: 'Kind',
    value: 'kind',
  },
  {
    id: 'namespace',
    label: 'Namespace',
    value: 'namespace',
  },
  {
    id: 'clusterName',
    label: 'Cluster',
    value: 'clusterName',
  },
  {
    id: 'tenant',
    label: 'Tenant',
    value: 'tenant',
  },
  {
    id: 'status',
    label: 'Status',
    value: (o: Object) => {
      if (o.status === '-') {
        return '-';
      }

      return (
        <Flex align>
          <Box marginRight={1}>
            <Icon
              size={24}
              {...getIndicatorInfo(o?.status)}
            />
          </Box>
          {o?.status}
        </Flex>
      );
    },
  },
  {
    id: 'message',
    label: 'Message',
    value: 'message',
    maxWidth: 600,
  },
];

export function addFieldsWithIndex(
  fields: ExplorerField[],
  extraFieldsWithIndex: ExplorerField[],
) {
  const newFields = [...fields];
  // Allows for columns to be added anywhere in the table.
  // We sort them here because if we mutate fields out of order,
  // the column order won't be accurate. See test case for example.
  for (const extra of _.sortBy(extraFieldsWithIndex, 'index')) {
    if (typeof extra.index !== 'undefined') {
      newFields.splice(extra.index, 0, extra);
    } else {
      newFields.push(extra);
    }
  }

  return newFields;
}

function ExplorerTable({
  className,
  rows,
  enableBatchSync,
  fields,
  onColumnHeaderClick,
}: Props) {
  const r: Object[] = _.map(rows, o => ({
    // Doing some things here to make this work with the DataTable.
    // It handles rendering the sync/pause buttons.
    ...o,
    uid: o.id,
    clusterName: o.cluster,
    type: o.kind,
  }));

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
