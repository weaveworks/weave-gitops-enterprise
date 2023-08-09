import {
  DataTable,
  Flex,
  Icon,
  IconType,
  Link,
  RequestStateHandler,
  Severity,
  Text,
  Timestamp,
  V2Routes,
  formatURL,
} from '@weaveworks/weave-gitops';
import { useListFacets } from '../../../hooks/query';
import { URLQueryStateManager } from '../../Explorer/QueryStateManager';
import {
  QueryState,
  QueryStateProvider,
  columnHeaderHandler,
} from '../../Explorer/hooks';

// @ts-ignore
import { IconButton } from '@material-ui/core';
import { RequestError } from '@weaveworks/weave-gitops/ui/lib/types';
import { useState } from 'react';
import FilterDrawer from '../../Explorer/FilterDrawer';
import Filters from '../../Explorer/Filters';
import PaginationControls from '../../Explorer/PaginationControls';
import QueryInput from '../../Explorer/QueryInput';
import QueryStateChips from '../../Explorer/QueryStateChips';
import { TableWrapper } from '../../Shared';
import { QueryResponse } from '../../../api/query/query.pb';

type AuditProps = {
  data: QueryResponse | undefined;
  error: Error | null;
  isLoading: boolean;
  manager: URLQueryStateManager;
  queryState: QueryState;
  setQueryState: (queryState: QueryState) => void;
};

export const AuditTable = ({
  data,
  error,
  isLoading,
  manager,
  queryState,
  setQueryState,
}: AuditProps) => {
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);

  const { data: facetsRes } = useListFacets();
  const filteredFacets = facetsRes?.facets?.filter(f => f.field !== 'kind');

  const rows = data?.objects?.map(obj => {
    const { unstructured, cluster } = obj;
    const details = JSON.parse(unstructured || '');
    const {
      metadata: {
        annotations: { category, policy_name, severity, policy_id },
        creationTimestamp,
      },
      involvedObject: { namespace, name, kind },
      message,
    } = details;
    return {
      message,
      cluster,
      category,
      policy_name,
      severity,
      creationTimestamp,
      policy_id,
      namespace,
      name,
      kind,
    };
  });
  return (
    <QueryStateProvider manager={manager}>
      <RequestStateHandler error={error as RequestError} loading={isLoading}>
        <Flex wide column>
          <Flex align wide end>
            <QueryStateChips />
            <IconButton onClick={() => setFilterDrawerOpen(!filterDrawerOpen)}>
              <Icon
                size="normal"
                type={IconType.FilterIcon}
                color="neutral30"
              />
            </IconButton>
          </Flex>
          <Flex wide>
            <TableWrapper id="auditViolations-list">
              <DataTable
                key={rows?.length}
                rows={rows}
                fields={[
                  {
                    label: 'Message',
                    value: ({ message }) => (
                      <span title={message}>{message}</span>
                    ),
                    textSearchable: true,
                    maxWidth: 300,
                    sortValue: ({ message }) => message,
                  },
                  {
                    label: 'Cluster',
                    value: 'cluster',
                    sortValue: ({ cluster }) => cluster,
                  },
                  {
                    label: 'Application',
                    value: ({ namespace, name, kind }) =>
                      kind === 'Kustomization' || kind === 'HelmRelease'
                        ? `${namespace}/${name}`
                        : '-',
                    sortValue: ({ namespace, name }) => `${namespace}/${name}`,
                    maxWidth: 150,
                  },
                  {
                    label: 'Severity',
                    value: ({ severity }) => (
                      <Severity severity={severity || ''} />
                    ),
                    sortValue: ({ severity }) => severity,
                  },
                  {
                    label: 'Category',
                    value: ({ category }) => (
                      <span title={category}>{category}</span>
                    ),
                    sortValue: ({ category }) => category,
                    maxWidth: 100,
                  },

                  {
                    label: 'Violated Policy',
                    value: ({ policy_name, cluster, policy_id }) => (
                      <Link
                        to={formatURL(V2Routes.PolicyDetailsPage, {
                          clusterName: cluster,
                          id: policy_id,
                          name: policy_name,
                        })}
                        data-policy-name={policy_name}
                      >
                        <Text capitalize semiBold>
                          {policy_name}
                        </Text>
                      </Link>
                    ),
                    sortValue: ({ policy_name }) => policy_name,
                    maxWidth: 200,
                  },

                  {
                    label: 'Violation Time',
                    value: ({ creationTimestamp }) => (
                      <Timestamp time={creationTimestamp} />
                    ),
                    defaultSort: true,
                    sortValue: ({ creationTimestamp }) => {
                      const t =
                        creationTimestamp &&
                        new Date(creationTimestamp).getTime();
                      return t * -1;
                    },
                  },
                ]}
                hideSearchAndFilters
                onColumnHeaderClick={columnHeaderHandler(
                  queryState,
                  setQueryState,
                )}
              />
            </TableWrapper>
            <FilterDrawer
              onClose={() => setFilterDrawerOpen(false)}
              open={filterDrawerOpen}
            >
              <QueryInput />

              <Filters facets={filteredFacets || []} />
            </FilterDrawer>
          </Flex>
          <PaginationControls
            queryState={queryState}
            setQueryState={setQueryState}
            count={data?.objects?.length || 0}
          />
        </Flex>
      </RequestStateHandler>
    </QueryStateProvider>
  );
};
