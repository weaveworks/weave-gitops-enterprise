import { Alert } from '@material-ui/lab';
import _ from 'lodash';
import styled from 'styled-components';
import { GlobalOperand, useQueryService } from '../../hooks/query';
import ExplorerTable from './ExplorerTable';
import {
  columnHeaderHandler,
  filterChangeHandler,
  useQueryState,
} from './hook';
import QueryBuilder from './QueryBuilder';

type Props = {
  className?: string;
  scopedKinds: string[];
};

function ScopedExploreUI({ className, scopedKinds }: Props) {
  const [queryState, setQueryState] = useQueryState({
    enableURLState: false,
    filters: [
      {
        label: 'Failed',
        value: 'status:Failed',
      },
      ..._.map(scopedKinds, k => ({ label: k, value: `kind:${k}` })),
    ],
  });

  // If kind filter is selected, we have to change some query logic.
  const filterSelected = !!_.find(
    queryState.pinnedTerms,
    t => _.includes(t, 'kind:') || _.includes(t, 'status:'),
  );

  // Always add the kind filter since we are "scoped",
  // unless the user has already selected a kind filter.
  const terms = _.concat(
    queryState.pinnedTerms,
    filterSelected ? [] : _.map(scopedKinds, k => `kind:${k}`),
  );

  const { data, error, isLoading } = useQueryService({
    query: terms.join(','),
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: `${queryState.orderBy} ${
      queryState.orderDescending ? 'desc' : 'asc'
    }`,
    globalOperandOverride: filterSelected
      ? GlobalOperand.and
      : GlobalOperand.or,
  });

  if (isLoading) {
    return null;
  }

  if (error) {
    return <Alert severity="error">Error: {error.message}</Alert>;
  }

  return (
    <div className={className}>
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
        hideTextInput
      />
      <ExplorerTable
        className={className}
        rows={data?.objects || []}
        onColumnHeaderClick={columnHeaderHandler(queryState, setQueryState)}
      />
    </div>
  );
}

export default styled(ScopedExploreUI).attrs({
  className: ScopedExploreUI.name,
})``;
