import { useQuery } from 'react-query';
import { Query, QueryOpts, QueryResponse } from '../api/query/query.pb';

// query looks like "key:value,key:value"
function convertToOpts(query: string): QueryOpts {
  if (!query) {
    return { key: '', value: '' };
  }

  return { key: '', value: '' };
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
      keepPreviousData: true,
    },
  );
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
