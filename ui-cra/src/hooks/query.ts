import _ from 'lodash';
import { useQuery } from 'react-query';
import { Query, QueryResponse } from '../api/query/query.pb';

type QueryOpts = {
  terms?: string;
  filters?: string[];
  limit: number;
  offset: number;
  orderBy?: string;
  category?: string;
  ascending?: boolean;
};

export function useQueryService({
  terms,
  filters,
  limit,
  offset,
  orderBy,
  category,
  ascending,
}: QueryOpts) {
  const api = Query;

  if (category) {
    filters = _.concat(filters || [], ['+category:' + category]);
  }

  return useQuery<QueryResponse, Error>(
    ['query', { terms, filters, limit, offset, orderBy, ascending }],
    () => {
      return api.DoQuery({
        terms,
        filters,
        limit,
        offset,
        orderBy,
        ascending,
      });
    },
    {
      keepPreviousData: true,
      retry: false,
      refetchInterval: Infinity,
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
