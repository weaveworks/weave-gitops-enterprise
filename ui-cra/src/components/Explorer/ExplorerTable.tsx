import { Box } from '@material-ui/core';
import {
  Flex,
  Icon,
  IconType,
  Link,
  formatURL,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { getKindRoute } from '../../utils/nav';
// @ts-ignore
import { DataTable } from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import { QueryState } from './hooks';

type Props = {
  className?: string;
  onColumnHeaderClick?: (field: Field) => void;
  rows: Object[];
  queryState: QueryState;
  enableBatchSync?: boolean;
  sortField?: string;
};

const debouncedInputHandler = _.debounce((fn, val) => {
  fn(val);
}, 500);

function ExplorerTable({
  className,
  rows,
  enableBatchSync,
  sortField,
  onColumnHeaderClick,
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
    <div className={className}>
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
            label: 'Status',
            sortValue: () => 'status',
            defaultSort: sortField === 'status',
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
          {
            label: 'Message',
            value: 'message',
            defaultSort: sortField === 'message',
          },
        ]}
        rows={r}
        hideSearchAndFilters
        onColumnHeaderClick={onColumnHeaderClick}
        hasCheckboxes={enableBatchSync}
      />
    </div>
  );
}

export default styled(ExplorerTable).attrs({ className: ExplorerTable.name })`
  td:nth-child(6) {
    white-space: pre-wrap;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }

  flex: 1;
`;
