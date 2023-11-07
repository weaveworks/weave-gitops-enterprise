// @ts-ignore
import { CircularProgress, IconButton } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useState } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Facet } from '../../api/query/query.pb';
import { useListFacets, useQueryService } from '../../hooks/query';
import ExplorerTable, { FieldWithIndex } from './ExplorerTable';
import FilterDrawer from './FilterDrawer';
import Filters from './Filters';
import {
  columnHeaderHandler,
  QueryStateProvider,
  useGetUnstructuredObjects,
} from './hooks';
import PaginationControls from './PaginationControls';
import QueryInput from './QueryInput';
import QueryStateChips from './QueryStateChips';
import { QueryStateManager, URLQueryStateManager } from './QueryStateManager';

type Props = {
  className?: string;
  category?: 'automation' | 'source' | 'gitopsset' | 'template';
  enableBatchSync?: boolean;
  manager?: QueryStateManager;
  extraColumns?: FieldWithIndex[];
  linkToObject?: boolean;
};

function Explorer({
  className,
  category,
  enableBatchSync,
  manager,
  extraColumns,
  linkToObject,
}: Props) {
  const history = useHistory();
  if (!manager) {
    manager = new URLQueryStateManager(history);
  }

  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const { data: facetsRes } = useListFacets();
  const queryState = manager.read();
  const setQueryState = manager.write;

  const { data, error, isLoading, isRefetching, isPreviousData } =
    useQueryService({
      terms: queryState.terms,
      filters: queryState.filters,
      limit: queryState.limit,
      offset: queryState.offset,
      orderBy: queryState.orderBy,
      ascending: queryState.orderAscending,
      category,
    });

  // This will be true when the query has changed, but the data hasn't been fetched yet.
  // Allows us to animate the table while the query is being worked on.
  // It will be false on background fetches that happen at a regular interval (without user interaction).
  const isRespondingToQuery = isRefetching && isPreviousData;

  const unst = useGetUnstructuredObjects(data?.objects || []);
  const rows = _.map(data?.objects, (o: any) => ({
    ...o,
    parsed: unst[o.id],
  }));

  const filteredFacets = filterFacetsForCategory(facetsRes?.facets, category);

  if (isLoading) {
    return (
      // Set min-width here to fix a weird stuttering issue where the spinner had
      // inconsistent width while spinning.
      <Flex wide center style={{ minWidth: '100%' }}>
        <CircularProgress />
      </Flex>
    );
  }

  return (
    <QueryStateProvider manager={manager}>
      <div className={className}>
        {error && <Alert severity="error">{error.message}</Alert>}
        <Flex align wide>
          <div style={{ marginLeft: '0 auto', width: 80 }}>
            <CircularProgress
              size={24}
              style={{ display: isRespondingToQuery ? 'block' : 'none' }}
            />
          </div>
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
        </Flex>
        <Flex wide>
          <ExplorerTableWithBusyAnimation
            busy={isRespondingToQuery}
            queryState={queryState}
            rows={rows}
            onColumnHeaderClick={columnHeaderHandler(queryState, setQueryState)}
            enableBatchSync={enableBatchSync}
            sortField={queryState.orderBy}
            extraColumns={extraColumns}
            linkToObject={linkToObject}
          />

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
      </div>
    </QueryStateProvider>
  );
}

export default styled(Explorer).attrs({ className: Explorer.name })`
  width: 100%;
`;

// Gray out the table while we are responding to a query. This is a visual indication to the user the explorer is "thinking".
// This will animate on query changes (including ordering), but not on refretches.
const ExplorerTableWithBusyAnimation = styled(ExplorerTable)<{ busy: boolean }>`
  table tbody {
    opacity: ${props => (props.busy ? '0.5' : '1')};
  }
`;

const categoryKinds = {
  automation: ['Kustomization', 'HelmRelease'],
  source: [
    'GitRepository',
    'HelmRepository',
    'Bucket',
    'HelmChart',
    'OCIRepository',
  ],
  gitopsset: ['GitOpsSet'],
  template: ['Template'],
};

function filterFacetsForCategory(
  facets?: Facet[],
  category?: 'automation' | 'source' | 'gitopsset' | 'template',
): Facet[] {
  if (!category) {
    return _.sortBy(facets, 'field') as Facet[];
  }

  const withoutKind = _.filter(facets, facet => {
    return facet.field?.toLowerCase() !== 'kind';
  });

  const kindFacets = _.map(categoryKinds[category], k => k);

  withoutKind.unshift({
    field: 'kind',
    values: kindFacets,
  });

  return _.sortBy(withoutKind, 'field');
}
