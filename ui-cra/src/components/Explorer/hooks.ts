import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import qs from 'query-string';
import { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';

export type QueryState = {
  query: string;
  filters: { label: string; value: any }[];
  limit: number;
  offset: number;
  selectedFilter: string;
  orderBy: string;
  orderAscending: boolean;
};

const DEFAULT_LIMIT = 25;

type Config = {
  enableURLState?: boolean;
  filters?: { label: string; value: any }[];
};

export function useQueryState(
  cfg: Config,
): [QueryState, (s: QueryState) => void] {
  const history = useHistory();

  const { q, limit, orderBy, selectedFilter, ascending, offset } = qs.parse(
    history.location.search,
  );

  const l = parseInt(limit as string);
  const o = parseInt(offset as string);

  const [queryState, setQueryState] = useState<QueryState>({
    query: (q as string) || '',
    limit: !_.isNaN(l) ? l : DEFAULT_LIMIT,
    offset: !_.isNaN(o) ? o : 0,
    filters: cfg.filters || [],
    selectedFilter: (selectedFilter as string) || '',
    orderBy: (orderBy as string) || '',
    orderAscending: ascending === 'true',
  });

  useEffect(() => {
    if (!cfg.enableURLState) {
      return;
    }
    let offset: any = queryState.offset;
    let limit: any = queryState.limit;
    if (queryState.offset === 0) {
      offset = null;
      limit = null;
    }
    const q = qs.stringify(
      {
        q: queryState.query,
        limit,
        offset,
        ascending: queryState.orderAscending,
        orderBy: queryState.orderBy,
        selectedFilter: queryState.selectedFilter,
      },
      { skipNull: true, skipEmptyString: true },
    );
    history.replace(`?${q}`);
  }, [history, cfg.enableURLState, queryState]);

  return [queryState, setQueryState];
}

export const columnHeaderHandler =
  (queryState: QueryState, setQueryState: (next: QueryState) => void) =>
  (field: Field) => {
    let col = _.isFunction(field.value)
      ? field.sortValue && field.sortValue(field.value)
      : field.value;

    // Override column name to maintain compatibility with the DataTable sync buttons
    if (col === 'clusterName') {
      col = 'cluster';
    }

    setQueryState({
      ...queryState,
      orderBy: col as string,
      orderAscending:
        queryState.orderBy === col ? !queryState.orderAscending : false,
    });
  };

export const filterChangeHandler =
  (queryState: QueryState, setQueryState: (next: QueryState) => void) =>
  (val: string) => {
    setQueryState({
      ...queryState,
      query: val,
      offset: 0,
      selectedFilter: val,
    });
  };
