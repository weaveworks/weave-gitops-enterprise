import _ from 'lodash';
import { useQuery } from 'react-query';
import { Query, QueryOpts, QueryResponse } from '../api/query/query.pb';

function convertToOpts(query: string): QueryOpts[] {
  if (!query) {
    return [{ key: '', value: '' }];
  }
  const opts = _.map(query.split(','), term => {
    const [key, value] = term.split(':');
    return {
      key,
      value,
    };
  });

  return opts;
}

export function useQueryService(query: string) {
  const api = Query;

  return useQuery<QueryResponse, Error>(
    ['query', query],
    () => {
      const opts = convertToOpts(query);

      return api.DoQuery({
        query: opts,
      });
    },
    {
      cacheTime: Infinity,
      staleTime: Infinity,
      retry: false,
      keepPreviousData: true,
    },
  );
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
