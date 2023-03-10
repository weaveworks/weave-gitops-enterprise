import { useQuery } from 'react-query';
import { Query } from '../api/query/query.pb';

export function useQueryService() {
  const api = Query;

  return useQuery(['query'], () => api.DoQuery({}));
}

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
