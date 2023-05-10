import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import _ from 'lodash';
import qs from 'query-string';
import { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';

export type QueryState = {
  terms: string;
  filters: string[];
  limit: number;
  offset: number;
  orderBy: string;
  orderAscending: boolean;
};

const DEFAULT_LIMIT = 25;

type Config = {
  enableURLState?: boolean;
};

export function useQueryState(
  cfg: Config,
): [QueryState, (s: QueryState) => void] {
  const history = useHistory();

  const { terms, qFilters, limit, orderBy, ascending, offset } = qs.parse(
    history.location.search,
    { arrayFormat: 'comma' },
  );

  const l = parseInt(limit as string);
  const o = parseInt(offset as string);

  const [queryState, setQueryState] = useState<QueryState>({
    terms: (terms as string) || '',
    limit: !_.isNaN(l) ? l : DEFAULT_LIMIT,
    offset: !_.isNaN(o) ? o : 0,
    filters: qFilters ? _.split(qFilters as string, ',') : [],
    orderBy: (orderBy as string) || '',
    orderAscending: ascending === 'true',
  });

  usePersistURL(history, queryState, cfg.enableURLState);

  return [queryState, setQueryState];
}

export function usePersistURL(
  history: ReturnType<typeof useHistory>,
  queryState: QueryState,
  enabled?: boolean,
) {
  useEffect(() => {
    if (!enabled) {
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
        terms: queryState.terms,
        qFilters: queryState.filters,
        limit,
        offset,
        ascending: queryState.orderAscending,
        orderBy: queryState.orderBy,
      },
      { skipNull: true, skipEmptyString: true, arrayFormat: 'comma' },
    );

    history.replace(`?${q}`);
  }, [enabled, history, queryState]);
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
  (filters: { [key: string]: boolean }) => {
    const existing = [...queryState.filters];

    _.each(filters, (v, k) => {
      if (_.includes(existing, k) && !v) {
        _.remove(existing, f => f === k);
        return;
      }

      if (v) {
        existing.push(k);
      }
    });

    setQueryState({
      ...queryState,
      filters: existing,
    });
  };

export const textInputHandler =
  (queryState: QueryState, setQueryState: (next: QueryState) => void) =>
  (terms: string) => {
    setQueryState({
      ...queryState,
      terms,
    });
  };
