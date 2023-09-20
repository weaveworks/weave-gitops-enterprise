// @ts-ignore
import { Facet } from '../../api/query/query.pb';
import { useListFacets, useQueryService } from '../../hooks/query';
import ExplorerTable, { FieldWithIndex } from './ExplorerTable';
import FilterDrawer from './FilterDrawer';
import Filters from './Filters';
import {
  QueryStateProvider,
  columnHeaderHandler,
  useGetUnstructuredObjects,
} from './hooks';
import PaginationControls from './PaginationControls';
import QueryInput from './QueryInput';
import QueryStateChips from './QueryStateChips';
import { QueryStateManager, URLQueryStateManager } from './QueryStateManager';
import { IconButton } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import { useState } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';

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

  const { data, error } = useQueryService({
    terms: queryState.terms,
    filters: queryState.filters,
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
    category,
  });

  const unst = useGetUnstructuredObjects(data?.objects || []);
  const rows = _.map(data?.objects, (o: any) => ({
    ...o,
    parsed: unst[o.id],
  }));

  const filteredFacets = filterFacetsForCategory(facetsRes?.facets, category);

  return (
    <QueryStateProvider manager={manager}>
      <div className={className}>
        {error && <Alert severity="error">{error.message}</Alert>}
        <Flex align wide end>
          <QueryStateChips />
          <IconButton onClick={() => setFilterDrawerOpen(!filterDrawerOpen)}>
            <Icon size="normal" type={IconType.FilterIcon} color="neutral30" />
          </IconButton>
        </Flex>
        <Flex wide>
          <ExplorerTable
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
