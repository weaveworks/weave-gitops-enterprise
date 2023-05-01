import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import qs from 'query-string';
import { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';

export type QueryState = {
  query: string;
  pinnedTerms: string[];
  filters: { label: string; value: any }[];
  limit: number;
  offset: number;
  selectedFilter: string;
  orderBy: string;
  orderAscending: boolean;
};

function initialTerms(search: string) {
  const parsed: { q?: string } = qs.parse(search);

  return parsed.q ? parsed.q.split(',') : [];
}

const DEFAULT_LIMIT = 25;

type Config = {
  enableURLState?: boolean;
  filters?: { label: string; value: any }[];
};

export function useQueryState(
  cfg: Config,
): [QueryState, (s: QueryState) => void] {
  const history = useHistory();

  const [queryState, setQueryState] = useState<QueryState>({
    query: '',
    pinnedTerms: cfg.enableURLState
      ? initialTerms(history.location.search)
      : [],
    limit: DEFAULT_LIMIT,
    offset: 0,
    filters: cfg.filters || [],
    selectedFilter: '',
    orderBy: 'name',
    orderAscending: true,
  });

  useEffect(() => {
    if (!cfg.enableURLState) {
      return;
    }

    if (queryState.pinnedTerms.length === 0 && queryState.offset === 0) {
      history.replace(history.location.pathname);
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
        q: queryState.pinnedTerms.join(','),
        limit,
        offset,
        ascending: queryState.orderAscending,
      },
      { skipNull: true },
    );

    history.replace(`?${q}`);
  }, [
    history,
    cfg.enableURLState,
    queryState.offset,
    queryState.limit,
    queryState.pinnedTerms,
    queryState.orderAscending,
  ]);

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
    let nextVal = [val];

    if (val === '') {
      nextVal = [];
    }

    setQueryState({
      ...queryState,
      offset: 0,
      pinnedTerms: nextVal,
      selectedFilter: val,
    });
  };
