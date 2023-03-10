import _ from 'lodash';
import { useQuery } from 'react-query';
import { Query, QueryOpts } from '../api/query/query.pb';

function convertToOpts(terms: string[]): QueryOpts[] {
  const opts = _.map(terms, term => {
    const [key, value] = term.split(':');
    return {
      key,
      value,
    };
  });

  return opts;
}

export function useQueryService(terms?: string[]) {
  const api = Query;

  const val = convertToOpts(terms);
  // const val = useDebounce(, 500);

  return useQuery(['query', val], () => {
    if (!val) {
      return Promise.resolve(null);
    }

    return api.DoQuery({
      query: val,
    });
  });
}

// Copied and TS-ified from https://usehooks.com/useDebounce/
// export function useDebounce<T>(value: T, delay: number) {
//   if (process.env.NODE_ENV === 'test') {
//     return value;
//   }

//   const [debouncedValue, setDebouncedValue] = useState(value);

//   useEffect(() => {
//     const handler = setTimeout(() => {
//       setDebouncedValue(value);
//     }, delay);
//     return () => {
//       clearTimeout(handler);
//     };
//   }, [value, delay]);

//   return debouncedValue;
// }

export function useListAccessRules() {
  const api = Query;

  return useQuery(['listAccessRules'], () => api.DebugGetAccessRules({}));
}
