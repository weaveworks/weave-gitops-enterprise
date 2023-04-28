import { useQuery } from 'react-query';
import { Query, QueryResponse } from '../api/query/query.pb';

type QueryOpts = {
  query: string;
  limit: number;
  offset: number;
  orderBy?: string;
  scopedKinds?: string[];
};

export function useQueryService({
  query,
  limit,
  offset,
  orderBy,
  scopedKinds,
}: QueryOpts) {
  const api = Query;

  let q = query;

  if (scopedKinds) {
    for (const k of scopedKinds) {
      q += `+kind:${k}`;
    }
  }

  return useQuery<QueryResponse, Error>(
    ['query', { query, limit, offset, orderBy }],
    () => {
      return api.DoQuery({
        query: q,
        limit,
        offset,
        orderBy,
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
