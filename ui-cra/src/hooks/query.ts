import _ from 'lodash';
import { useEffect, useState } from 'react';
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

// Copied and TS-ified from https://usehooks.com/useDebounce/
export function useDebounce<T>(value: T, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
