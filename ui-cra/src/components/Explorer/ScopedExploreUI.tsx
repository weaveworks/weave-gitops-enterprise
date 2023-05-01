import { Box } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import _ from 'lodash';
import styled from 'styled-components';
import { useQueryService } from '../../hooks/query';
import ExplorerTable from './ExplorerTable';
import {
  columnHeaderHandler,
  filterChangeHandler,
  useQueryState,
} from './hooks';
import PaginationControls from './PaginationControls';
import QueryBuilder from './QueryBuilder';

type Props = {
  className?: string;
  enableBatchSync?: boolean;
  category: string;
};

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

function ScopedExploreUI({ className, category, enableBatchSync }: Props) {
  const kinds = (categoryKinds as any)[category];

  const [queryState, setQueryState] = useQueryState({
    enableURLState: false,
    filters: [
      ..._.map(kinds, k => ({
        label: k,
        value: `+kind:${k}`,
      })),
    ],
  });

  const { data, error, isLoading } = useQueryService({
    query: queryState.pinnedTerms.join(','),
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
    category,
  });

  if (isLoading) {
    return null;
  }

  if (error) {
    return <Alert severity="error">Error: {error.message}</Alert>;
  }

  return (
    <div className={className}>
      <Box marginY={2}>
        <QueryBuilder
          busy={isLoading}
          query={queryState.query}
          filters={queryState.filters}
          selectedFilter={queryState.selectedFilter}
          pinnedTerms={queryState.pinnedTerms}
          onChange={(query, pinnedTerms) => {
            setQueryState({ ...queryState, query, pinnedTerms });
          }}
          onSubmit={pinnedTerms => {
            setQueryState({ ...queryState, pinnedTerms });
          }}
          onFilterSelect={filterChangeHandler(queryState, setQueryState)}
        />
      </Box>

      <ExplorerTable
        className={className}
        rows={data?.objects || []}
        onColumnHeaderClick={columnHeaderHandler(queryState, setQueryState)}
        enableBatchSync={enableBatchSync}
      />
      <PaginationControls
        queryState={queryState}
        setQueryState={setQueryState}
        count={data?.objects?.length || 0}
      />
    </div>
  );
}

export default styled(ScopedExploreUI).attrs({
  className: ScopedExploreUI.name,
})``;
