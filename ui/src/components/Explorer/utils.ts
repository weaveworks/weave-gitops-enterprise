import { QueryState } from './hooks';
import qs from 'query-string';

export function linkToExplorer(path: string, q: QueryState): string {
  const { filters, ...rest } = q;
  const search = qs.stringify(
    {
      ...rest,
      //   Adding a namespace to avoid conflicting with DataTable
      qFilters: filters,
    },
    { skipEmptyString: true, skipNull: true, arrayFormat: 'comma' },
  );

  return `${path}?${search}`;
}
