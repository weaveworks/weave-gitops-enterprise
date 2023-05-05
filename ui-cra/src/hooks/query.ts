import { useQuery } from 'react-query';
import { Query, QueryResponse } from '../api/query/query.pb';

type QueryOpts = {
  query: string;
  limit: number;
  offset: number;
  orderBy?: string;
  category?: string;
  ascending?: boolean;
};

export function useQueryService({
  query,
  limit,
  offset,
  orderBy,
  category,
  ascending,
}: QueryOpts) {
  const api = Query;

  let q = query;

  if (category) {
    q += ' +category:' + category;
  }

  return useQuery<QueryResponse, Error>(
    ['query', { query, limit, offset, orderBy, ascending }],
    () => {
      return api.DoQuery({
        query: q,
        limit,
        offset,
        orderBy,
        ascending,
      });
    },
    {
      keepPreviousData: true,
      retry: false,
      refetchInterval: 5000,
    },
  );
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}

export function useListFacets() {
  const api = Query;

  return useQuery(['facets'], () => api.ListFacets({}), {
    refetchIntervalInBackground: false,
    refetchInterval: Infinity,
  });
}
