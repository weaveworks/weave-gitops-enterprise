import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
// @ts-ignore
import { IconButton } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import _ from 'lodash';
import { useState } from 'react';
import styled from 'styled-components';
import { Facet } from '../../api/query/query.pb';
import { useListFacets, useQueryService } from '../../hooks/query';
import ExplorerTable from './ExplorerTable';
import FilterDrawer from './FilterDrawer';
import Filters from './Filters';
import PaginationControls from './PaginationControls';
import QueryInput from './QueryInput';
import {
  columnHeaderHandler,
  filterChangeHandler,
  textInputHandler,
  useQueryState,
} from './hooks';

type Props = {
  className?: string;
  category?: 'automation' | 'source';
  enableBatchSync?: boolean;
};

function Explorer({ className, category, enableBatchSync }: Props) {
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const { data: facetsRes } = useListFacets();
  const [queryState, setQueryState] = useQueryState({
    enableURLState: false,
  });

  const filters = _.reduce(
    queryState.filters,
    (result, f) => {
      const re = /[%w|+](.+?):(.*)/g;

      const matches = re.exec(f);

      if (matches) {
        const [, key, value] = matches;

        result[`${key}:${value}`] = true;
      }

      return result;
    },
    {} as { [key: string]: boolean },
  );

  const { data, error } = useQueryService({
    terms: queryState.terms,
    filters: queryState.filters,
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
    category,
  });

  const filteredFacets = filterFacetsForCategory(facetsRes?.facets, category);

  return (
    <div className={className}>
      {error && <Alert severity="error">{error.message}</Alert>}
      <Flex wide end>
        <IconButton onClick={() => setFilterDrawerOpen(!filterDrawerOpen)}>
          <Icon size="normal" type={IconType.FilterIcon} />
        </IconButton>
      </Flex>
      <Flex wide>
        <ExplorerTable
          queryState={queryState}
          rows={data?.objects || []}
          onColumnHeaderClick={columnHeaderHandler(queryState, setQueryState)}
          enableBatchSync={enableBatchSync}
        />

        <FilterDrawer
          onClose={() => setFilterDrawerOpen(false)}
          open={filterDrawerOpen}
        >
          <QueryInput
            queryState={queryState}
            onTextInputChange={textInputHandler(queryState, setQueryState)}
          />

          <Filters
            facets={filteredFacets || []}
            onFilterSelect={filterChangeHandler(queryState, setQueryState)}
            state={filters}
          />
        </FilterDrawer>
      </Flex>

      <PaginationControls
        queryState={queryState}
        setQueryState={setQueryState}
        count={data?.objects?.length || 0}
      />
    </div>
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
};

function filterFacetsForCategory(
  facets?: Facet[],
  category?: 'automation' | 'source',
): Facet[] {
  if (!category) {
    return _.sortBy(facets, 'field') as Facet[];
  }

  const withoutKind = _.filter(facets, facet => {
    return facet.field?.toLowerCase() !== 'kind';
  });

  const kindFacets = _.map(categoryKinds[category], k => k.toLowerCase());

  withoutKind.unshift({
    field: 'Kind',
    values: kindFacets,
  });

  return _.sortBy(withoutKind, 'field');
}
