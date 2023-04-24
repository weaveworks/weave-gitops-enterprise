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
  scopedKinds: string[];
  enableBatchSync?: boolean;
};

function ScopedExploreUI({ className, scopedKinds, enableBatchSync }: Props) {
  const [queryState, setQueryState] = useQueryState({
    enableURLState: false,
    filters: [..._.map(scopedKinds, k => ({ label: k, value: `kind:${k}` }))],
  });

  const { data, error, isLoading } = useQueryService({
    query: queryState.pinnedTerms.join(','),
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: `${queryState.orderBy} ${
      queryState.orderDescending ? 'desc' : 'asc'
    }`,

    scopedKinds,
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
          onPin={pinnedTerms => {
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
